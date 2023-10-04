package consumer

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/trips-api/internal/services/bundlr"
	es_store "github.com/DIMO-Network/trips-api/internal/services/es"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types/pgeo"
)

type Consumer struct {
	logger           *zerolog.Logger
	es               *es_store.Client
	pg               *pg_store.Store
	bundlr           *bundlr.Client
	dataFetchEnabled bool
	workerCount      int
	bundlrEnabled    bool
}

type SegmentEvent struct {
	Start    bundlr.PointTime `json:"start"`
	End      bundlr.PointTime `json:"end"`
	DeviceID string           `json:"deviceID"`
}

type UserDeviceMintEvent struct {
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"userId"`
	Device    struct {
		ID string `json:"id"`
	} `json:"device"`
	NFT struct {
		TokenID *big.Int `json:"tokenId"`
	} `json:"nft"`
}

const WorkerPoolSize = 20
const UserDeviceMintEventType = "com.dimo.zone.device.mint"

func New(es *es_store.Client, bundlrClient *bundlr.Client, pg *pg_store.Store, logger *zerolog.Logger, dataFetchEnabled bool, workerCount int, bundlrEnabled bool) *Consumer {
	return &Consumer{logger, es, pg, bundlrClient, dataFetchEnabled, workerCount, bundlrEnabled}
}

func (c *Consumer) CompletedSegment(ctx context.Context, event shared.CloudEvent[SegmentEvent]) error {
	v, err := models.Vehicles(models.VehicleWhere.UserDeviceID.EQ(event.Data.DeviceID)).One(ctx, c.pg.DB)
	if err != nil {
		return fmt.Errorf("failed to find vehicle %s: %w", event.Subject, err)
	}

	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		return fmt.Errorf("couldn't produce random key: %w", err)
	}

	var bundlrID null.String

	if c.dataFetchEnabled {
		response, err := c.es.FetchData(ctx, event.Data.DeviceID, event.Data.Start.Time, event.Data.End.Time)
		if err != nil {
			return fmt.Errorf("call to Elasticsearch failed: %w", err)
		}

		dataItem, err := c.bundlr.PrepareData(response, encryptionKey, v.TokenID, event.Data.Start.Time, event.Data.End.Time)
		if err != nil {
			return fmt.Errorf("assembly for Bundlr failed: %w", err)
		}

		if c.bundlrEnabled {
			if err := c.bundlr.Upload(dataItem); err != nil {
				return err
			}
		}

		bundlrID = null.StringFrom(dataItem.Id.Base64())

		c.logger.Info().Msgf("https://devnet.bundlr.network/%s", dataItem.Id.Base64())
	}

	segment := models.Trip{
		VehicleTokenID: v.TokenID,
		EncryptionKey:  null.BytesFrom(encryptionKey),
		ID:             ksuid.New().String(),
		StartTime:      event.Data.Start.Time,
		EndTime:        null.TimeFrom(event.Data.End.Time),
		StartPosition:  pointToDB(event.Data.Start.Point),
		EndPosition:    pgeo.NewNullPoint(pointToDB(event.Data.End.Point), true),
		BundlrID:       bundlrID,
	}

	if err := segment.Insert(ctx, c.pg.DB, boil.Infer()); err != nil {
		return fmt.Errorf("failed to insert new trip: %w", err)
	}

	return nil
}

func (c *Consumer) VehicleEvent(ctx context.Context, event shared.CloudEvent[UserDeviceMintEvent]) error {
	if event.Type == UserDeviceMintEventType {
		return c.pg.StoreVehicle(ctx, event.Data.Device.ID, int(event.Data.NFT.TokenID.Int64()))
	}
	return nil
}

func pointToDB(p bundlr.Point) pgeo.Point {
	return pgeo.NewPoint(p.Longitude, p.Latitude)
}
