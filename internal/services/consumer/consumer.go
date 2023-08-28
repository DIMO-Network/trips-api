package consumer

import (
	"context"
	"strings"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/kafka"
	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/DIMO-Network/trips-api/internal/services/bundlr"
	es_store "github.com/DIMO-Network/trips-api/internal/services/es"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/rs/zerolog"
)

type CompletedSegmentConsumer struct {
	config kafka.Config
	logger *zerolog.Logger
	es     *es_store.Client
	pg     *pg_store.Store
	bundlr *bundlr.Client
}

type SegmentEvent struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	DeviceID string    `json:"deviceID"`
}

func New(es *es_store.Client, bundlrClient *bundlr.Client, pg *pg_store.Store, settings *config.Settings, logger *zerolog.Logger) (*CompletedSegmentConsumer, error) {
	kc := kafka.Config{
		Brokers: strings.Split(settings.KafkaBrokers, ","),
		Topic:   settings.TripEventTopic,
		Group:   "segmenter",
	}

	return &CompletedSegmentConsumer{kc, logger, es, pg, bundlrClient}, nil
}

func (c *CompletedSegmentConsumer) Start(ctx context.Context) {
	if err := kafka.Consume(ctx, c.config, c.ingest, c.logger); err != nil {
		c.logger.Fatal().Err(err).Msg("Couldn't start segment consumer.")
	}
	c.logger.Info().Msg("segment consumer started.")
}

func (c *CompletedSegmentConsumer) ingest(ctx context.Context, event *shared.CloudEvent[SegmentEvent]) error {
	response, err := c.es.FetchData(ctx, event.Data.DeviceID, event.Data.Start, event.Data.End)
	if err != nil {
		return err
	}

	if _, _, err := c.bundlr.PrepareData(response, event.Data.DeviceID, event.Data.Start, event.Data.End); err != nil {
		return err
	}

	return nil
}
