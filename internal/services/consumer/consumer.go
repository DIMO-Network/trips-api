package consumer

import (
	"context"
	"strings"
	"time"

	pb_devices "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/kafka"
	"github.com/DIMO-Network/trips-api/internal/config"
	es_store "github.com/DIMO-Network/trips-api/internal/services/es"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/DIMO-Network/trips-api/internal/services/uploader"
	"github.com/rs/zerolog"
)

type CompletedSegmentConsumer struct {
	config kafka.Config
	logger *zerolog.Logger
	es     *es_store.Store
	pg     *pg_store.Store
	grpc   pb_devices.UserDeviceServiceClient
	*uploader.Uploader
}

type SegmentEvent struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	DeviceID string    `json:"deviceID"`
}

func New(es *es_store.Store, uploader *uploader.Uploader, pg *pg_store.Store, grpcDevices pb_devices.UserDeviceServiceClient, settings *config.Settings, logger *zerolog.Logger) (*CompletedSegmentConsumer, error) {
	kc := kafka.Config{
		Brokers: strings.Split(settings.KafkaBrokers, ","),
		Topic:   settings.TripEventTopic,
		Group:   "segmenter",
	}

	return &CompletedSegmentConsumer{kc, logger, es, pg, grpcDevices, uploader}, nil
}

func (c *CompletedSegmentConsumer) Start(ctx context.Context) {
	if err := kafka.Consume(ctx, c.config, c.ingest, c.logger); err != nil {
		c.logger.Fatal().Err(err).Msg("Couldn't start segment consumer.")
	}
	c.logger.Info().Msg("segment consumer started.")
}

func (c *CompletedSegmentConsumer) ingest(ctx context.Context, event *shared.CloudEvent[SegmentEvent]) error {
	response, err := c.es.FetchData(event.Data.DeviceID, event.Data.Start.Format(time.RFC3339), event.Data.End.Format(time.RFC3339))
	if err != nil {
		return err
	}

	dataItem, encryptionKey, err := c.PrepareData(response, event.Data.DeviceID, event.Data.Start.Format(time.RFC3339), event.Data.End.Format(time.RFC3339))
	if err != nil {
		return err
	}

	userDevice, err := c.grpc.GetUserDevice(ctx, &pb_devices.GetUserDeviceRequest{
		Id: event.Data.DeviceID,
	})
	if err != nil {
		return err
	}

	err = c.pg.StoreSegmentMetadata(ctx, *userDevice.TokenId, encryptionKey, response, dataItem.Id.Base64())
	if err != nil {
		return err
	}
	// upload

	return nil
}
