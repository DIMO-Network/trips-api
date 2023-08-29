package consumer

import (
	"context"
	"math/big"
	"time"

	pb_devices "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/kafka"
	"github.com/DIMO-Network/trips-api/internal/services/bundlr"
	es_store "github.com/DIMO-Network/trips-api/internal/services/es"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type Consumer struct {
	logger *zerolog.Logger
	es     *es_store.Client
	pg     *pg_store.Store
	bundlr *bundlr.Client
	grpc   pb_devices.UserDeviceServiceClient
}

type SegmentEvent struct {
	Start    PointTime `json:"start"`
	End      PointTime `json:"end"`
	DeviceID string    `json:"deviceID"`
}

type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type PointTime struct {
	Point Point     `json:"point"`
	Time  time.Time `json:"time"`
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

const UserDeviceMintEventType = "com.dimo.zone.device.mint"

func New(es *es_store.Client, bundlrClient *bundlr.Client, pg *pg_store.Store, devicesGRPC pb_devices.UserDeviceServiceClient, logger *zerolog.Logger) *Consumer {
	return &Consumer{logger, es, pg, bundlrClient, devicesGRPC}
}

func Start[A any](ctx context.Context, config kafka.Config, handler func(context.Context, *shared.CloudEvent[A]) error, logger *zerolog.Logger) {
	if err := kafka.Consume(ctx, config, handler, logger); err != nil {
		logger.Fatal().Err(err).Msgf("Couldn't start %s consumer.", config.Group)
	}
	logger.Info().Msgf("%s consumer started.", config.Group)
}

func (c *Consumer) CompletedSegment(ctx context.Context, event *shared.CloudEvent[SegmentEvent]) error {
	response, err := c.es.FetchData(ctx, event.Data.DeviceID, event.Data.Start.Time, event.Data.End.Time)
	if err != nil {
		return err
	}

	vehicleData, err := c.pg.GetOrGenerateEncryptionKey(ctx, event.Data.DeviceID, c.grpc)
	if err != nil {
		return err
	}

	dataItem, nonce, err := c.bundlr.PrepareData(response, vehicleData.EncryptionKey, vehicleData.UserDeviceID, event.Data.Start.Time, event.Data.End.Time)
	if err != nil {
		return err
	}

	segment := models.Trip{
		VehicleTokenID: types.NullDecimal(vehicleData.TokenID),
		UserDeviceID:   event.Data.DeviceID,
		ID:             ksuid.New().String(),
		Start:          event.Data.Start.Time,
		End:            null.TimeFrom(event.Data.End.Time),
		Nonce:          nonce,
		BundlrID:       null.StringFrom(dataItem.Id.Base64()),
	}
	if err := segment.Insert(
		ctx,
		c.pg.DB,
		boil.Whitelist(
			models.TripColumns.VehicleTokenID,
			models.TripColumns.ID,
			models.TripColumns.UserDeviceID,
			models.TripColumns.Nonce,
			models.TripColumns.BundlrID,
			models.TripColumns.Start,
			models.TripColumns.End)); err != nil {
		return err
	}

	// upload

	return nil
}

func (c *Consumer) VehicleEvent(ctx context.Context, event *shared.CloudEvent[UserDeviceMintEvent]) error {
	if event.Type == UserDeviceMintEventType {
		c.logger.Info().Str("device", event.Data.Device.ID).Msg("vehicle node minted event recieved")
		if _, err := c.pg.GenerateKey(ctx, event.Data.Device.ID, event.Data.NFT.TokenID.Uint64(), c.grpc); err != nil {
			c.logger.Err(err).Str("device", event.Data.Device.ID).Msg("encryption key generation failed")
			return err
		}
	}
	return nil
}
