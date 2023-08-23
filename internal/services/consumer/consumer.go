package consumer

import (
	"context"
	"strings"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/kafka"
	"github.com/DIMO-Network/trips-api/internal/config"
	es_store "github.com/DIMO-Network/trips-api/internal/services/es"
	"github.com/DIMO-Network/trips-api/internal/services/uploader"
	"github.com/rs/zerolog"
)

type CompletedSegmentConsumer struct {
	config kafka.Config
	logger *zerolog.Logger
	es     *es_store.Store
	*uploader.Uploader
}

type SegmentEvent struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	DeviceID string    `json:"deviceID"`
}

func New(es *es_store.Store, uploader *uploader.Uploader, settings *config.Settings, logger *zerolog.Logger) (*CompletedSegmentConsumer, error) {
	kc := kafka.Config{
		Brokers: strings.Split(settings.KafkaBrokers, ","),
		Topic:   settings.TripEventTopic,
		Group:   "segmenter",
	}

	return &CompletedSegmentConsumer{kc, logger, es, uploader}, nil
}

func (c *CompletedSegmentConsumer) Start(ctx context.Context) {
	if err := kafka.Consume(ctx, c.config, c.ingest, c.logger); err != nil {
		c.logger.Fatal().Err(err).Msg("Couldn't start segment consumer.")
	}
	c.logger.Info().Msg("segment consumer started.")
}

func (c *CompletedSegmentConsumer) ingest(_ context.Context, event *shared.CloudEvent[SegmentEvent]) error {
	response, err := c.es.FetchData(event.Data.DeviceID, event.Data.Start.Format(time.RFC3339), event.Data.End.Format(time.RFC3339))
	if err != nil {
		return err
	}

	_, err = c.PrepareData(response, event.Data.Start.Format(time.RFC3339), event.Data.End.Format(time.RFC3339))
	if err != nil {
		return err
	}

	return nil
}
