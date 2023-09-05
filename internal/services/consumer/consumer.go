package consumer

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

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
)

type Consumer struct {
	logger *zerolog.Logger
	es     *es_store.Client
	pg     *pg_store.Store
	bundlr *bundlr.Client
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

func New(es *es_store.Client, bundlrClient *bundlr.Client, pg *pg_store.Store, logger *zerolog.Logger) *Consumer {
	return &Consumer{logger, es, pg, bundlrClient}
}

func Start[A any](ctx context.Context, config kafka.Config, handler func(context.Context, *shared.CloudEvent[A]) error, logger *zerolog.Logger) {
	if err := kafka.Consume(ctx, config, handler, logger); err != nil {
		logger.Fatal().Err(err).Msgf("Couldn't start %s consumer.", config.Group)
	}
	logger.Info().Msgf("%s consumer started.", config.Group)
}

func (c *Consumer) CompletedSegment(ctx context.Context, event *shared.CloudEvent[SegmentEvent]) error {
	v, err := models.Vehicles(models.VehicleWhere.UserDeviceID.EQ(event.Data.DeviceID)).One(ctx, c.pg.DB)
	if err != nil {
		return err
	}

	response, err := c.es.FetchData(ctx, event.Data.DeviceID, event.Data.Start.Time, event.Data.End.Time)
	if err != nil {
		return err
	}

	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		return err
	}

	dataItem, err := c.bundlr.PrepareData(response, encryptionKey, v.TokenID, event.Data.Start.Time, event.Data.End.Time)
	if err != nil {
		return err
	}

	segment := models.Trip{
		VehicleTokenID: v.TokenID,
		EncryptionKey:  null.BytesFrom(encryptionKey),
		ID:             ksuid.New().String(),
		Start:          event.Data.Start.Time,
		End:            null.TimeFrom(event.Data.End.Time),
		BundlrID:       null.StringFrom(dataItem.Id.Base64()),
	}
	if err := segment.Insert(
		ctx,
		c.pg.DB,
		boil.Infer()); err != nil {
		return err
	}

	err = c.bundlr.Upload(dataItem)
	if err != nil {
		return err
	}

	c.logger.Info().Msgf("https://devnet.bundlr.network/%s", dataItem.Id.Base64())
	return nil
}

func (c *Consumer) VehicleEvent(ctx context.Context, event *shared.CloudEvent[UserDeviceMintEvent]) error {
	if event.Type == UserDeviceMintEventType {
		return c.pg.StoreVehicle(ctx, event.Data.Device.ID, int(event.Data.NFT.TokenID.Int64()))
	}
	return nil
}
