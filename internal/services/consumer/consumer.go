package consumer

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"time"

	pb_devices "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/trips-api/internal/services/bundlr"
	es_store "github.com/DIMO-Network/trips-api/internal/services/es"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type Consumer struct {
	logger           *zerolog.Logger
	es               *es_store.Client
	pg               *pg_store.Store
	bundlr           *bundlr.Client
	devicesClient    pb_devices.UserDeviceServiceClient
	dataFetchEnabled bool
	workerCount      int
	bundlrEnabled    bool
}

type Endpoint struct {
	Time time.Time `json:"time"`
}

type SegmentEvent struct {
	TripID       string   `json:"id"`
	UserDeviceID string   `json:"deviceId"`
	Completed    bool     `json:"completed"`
	Start        Endpoint `json:"start"`
	End          Endpoint `json:"end"`
}

type UserDeviceMintEvent struct {
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"userId"`
	Device    struct {
		ID string `json:"id"`
	} `json:"device"`
	NFT struct {
		TokenID int            `json:"tokenId"`
		Owner   common.Address `json:"address"`
	} `json:"nft"`
}

const UserDeviceMintEventType = "com.dimo.zone.device.mint"

func New(es *es_store.Client, bundlrClient *bundlr.Client, pg *pg_store.Store, devicesClient pb_devices.UserDeviceServiceClient, logger *zerolog.Logger, dataFetchEnabled bool, workerCount int, bundlrEnabled bool) *Consumer {
	return &Consumer{logger, es, pg, bundlrClient, devicesClient, dataFetchEnabled, workerCount, bundlrEnabled}
}

func (c *Consumer) ProcessSegmentEvent(ctx context.Context, event shared.CloudEvent[SegmentEvent]) error {
	if event.Data.Completed {
		return c.CompleteSegment(ctx, event)
	}

	ud, err := c.devicesClient.GetUserDevice(ctx, &pb_devices.GetUserDeviceRequest{Id: event.Data.UserDeviceID})
	if err != nil {
		return fmt.Errorf("failed to get user device %s: %w", event.Data.UserDeviceID, err)
	}

	if ud.TokenId == nil || ud.OwnerAddress == nil {
		return fmt.Errorf("failed to store segment; missing token id or owner address for device %s", event.Data.UserDeviceID)

	}

	return c.BeginSegment(ctx, event, int64(*ud.TokenId), ud.OwnerAddress)
}

func (c *Consumer) BeginSegment(ctx context.Context, event shared.CloudEvent[SegmentEvent], vehicleTokenId int64, owner []byte) error {
	segment := models.Trip{
		ID:             event.Data.TripID,
		VehicleTokenID: int(vehicleTokenId),
		OwnerAddress:   null.BytesFrom(owner),
		StartTime:      event.Data.Start.Time,
	}

	return segment.Insert(ctx, c.pg.DB.DBS().Writer, boil.Infer())
}

func (c *Consumer) CompleteSegment(ctx context.Context, event shared.CloudEvent[SegmentEvent]) error {
	segment, err := models.Trips(models.TripWhere.ID.EQ(event.Data.TripID)).One(ctx, c.pg.DB.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to find segment %s: %w", event.Data.TripID, err)
		}
		return err
	}

	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		return fmt.Errorf("couldn't produce random key: %w", err)
	}

	if c.dataFetchEnabled {
		response, err := c.es.FetchData(ctx, event.Data.UserDeviceID, segment.StartTime, event.Data.End.Time)
		if err != nil {
			return fmt.Errorf("call to Elasticsearch failed: %w", err)
		}

		dataItem, err := c.bundlr.PrepareData(response, encryptionKey, segment.VehicleTokenID, segment.StartTime, event.Data.End.Time)
		if err != nil {
			return fmt.Errorf("assembly for Bundlr failed: %w", err)
		}

		if c.bundlrEnabled {
			if err := c.bundlr.Upload(dataItem); err != nil {
				return err
			}
		}

		segment.BundlrID = null.StringFrom(dataItem.Id.Base64())
		c.logger.Info().Msgf("https://devnet.bundlr.network/%s", segment.BundlrID.String)
	}

	segment.EncryptionKey = null.BytesFrom(encryptionKey)
	segment.EndTime = null.TimeFrom(event.Data.End.Time)
	_, err = segment.Update(ctx, c.pg.DB.DBS().Writer, boil.Whitelist(models.TripColumns.EncryptionKey, models.TripColumns.EndTime, models.TripColumns.BundlrID))
	return err
}
