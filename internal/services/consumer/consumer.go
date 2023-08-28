package consumer

import (
	"context"
	"time"

	pb_devices "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/kafka"
	"github.com/DIMO-Network/trips-api/internal/services/bundlr"
	es_store "github.com/DIMO-Network/trips-api/internal/services/es"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/rs/zerolog"
)

type Consumer struct {
	logger *zerolog.Logger
	es     *es_store.Client
	pg     *pg_store.Store
	bundlr *bundlr.Client
	grpc   pb_devices.UserDeviceServiceClient
}

type SegmentEvent struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	DeviceID string    `json:"deviceID"`
}

func New(es *es_store.Client, bundlrClient *bundlr.Client, pg *pg_store.Store, devicesGRPC pb_devices.UserDeviceServiceClient, logger *zerolog.Logger) *Consumer {
	return &Consumer{logger, es, pg, bundlrClient, devicesGRPC}
}

func Start(ctx context.Context, config kafka.Config, handler func(context.Context, any) error, logger *zerolog.Logger) {
	if err := kafka.Consume(ctx, config, handler, logger); err != nil {
		logger.Fatal().Err(err).Msg("Couldn't start segment consumer.")
	}
	logger.Info().Msg("segment consumer started.")
}

func (c *Consumer) CompletedSegment(ctx context.Context, e any) error {
	event := e.(*shared.CloudEvent[SegmentEvent])
	response, err := c.es.FetchData(ctx, event.Data.DeviceID, event.Data.Start, event.Data.End)
	if err != nil {
		return err
	}

	if _, _, err := c.bundlr.PrepareData(response, event.Data.DeviceID, event.Data.Start, event.Data.End); err != nil {
		return err
	}

	// userDevice, err := c.grpc.GetUserDevice(ctx, &pb_devices.GetUserDeviceRequest{
	// 	Id: event.Data.DeviceID,
	// })
	// if err != nil {
	// 	return err
	// }

	// err = c.pg.StoreSegmentMetadata(ctx, *userDevice.TokenId, encryptionKey, response, dataItem.Id.Base64())
	// if err != nil {
	// 	return err
	// }
	// upload

	return nil
}

func (c *Consumer) VehicleEvent(ctx context.Context, e any) error {

	// populate vehicles table: token_id, user_device_id, encryption_key
	// look at devices-api
	return nil
}
