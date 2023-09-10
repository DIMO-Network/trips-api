package consumer

import (
	"context"
	"crypto/rand"
	"math/big"
	"sync"
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

const WorkerPoolSize = 9
const UserDeviceMintEventType = "com.dimo.zone.device.mint"

func New(es *es_store.Client, bundlrClient *bundlr.Client, pg *pg_store.Store, logger *zerolog.Logger) *Consumer {
	return &Consumer{logger, es, pg, bundlrClient}
}

func Start[A any](ctx context.Context, config kafka.Config, handler func(context.Context, int, chan shared.CloudEvent[A], *sync.WaitGroup, *zerolog.Logger), taskChan chan shared.CloudEvent[A], wg *sync.WaitGroup, logger *zerolog.Logger) {
	l := logger.With().Str("group", config.Group).Logger()

	for i := 0; i < WorkerPoolSize; i++ {
		l.Info().Msgf("starting worker %d", i+1)
		wg.Add(1)
		go handler(ctx, i, taskChan, wg, &l)
	}

	if err := kafka.Consume(ctx, config, func(ctx context.Context, evt *shared.CloudEvent[A]) error {
		taskChan <- *evt
		return nil
	}, &l); err != nil {
		l.Err(err).Msg("unable to consume")
		return
	}
}

func (c *Consumer) CompletedSegment(ctx context.Context, workerNum int, taskChan chan shared.CloudEvent[SegmentEvent], wg *sync.WaitGroup, logger *zerolog.Logger) {
	defer wg.Done()
	for event := range taskChan {
		v, err := models.Vehicles(models.VehicleWhere.UserDeviceID.EQ(event.Data.DeviceID)).One(ctx, c.pg.DB)
		if err != nil {
			logger.Err(err).Msg("unable to find vehicle using device ID")
			return
		}

		response, err := c.es.FetchData(ctx, event.Data.DeviceID, event.Data.Start.Time, event.Data.End.Time)
		if err != nil {
			logger.Err(err).Msg("unable to fetch data from elasticsearch")
		}

		encryptionKey := make([]byte, 32)
		if _, err := rand.Read(encryptionKey); err != nil {
			logger.Err(err).Msg("unable to make encryption key")
		}

		dataItem, err := c.bundlr.PrepareData(response, encryptionKey, v.TokenID, event.Data.Start.Time, event.Data.End.Time)
		if err != nil {
			logger.Err(err).Msg("unable to prepare data")
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
			logger.Err(err).Msg("unable to insert segment to trips table")
		}

		// upload
	}
	logger.Info().Int("workerNum", workerNum).Msg("shutdown")
}

func (c *Consumer) VehicleEvent(ctx context.Context, workerNum int, taskChan chan shared.CloudEvent[UserDeviceMintEvent], wg *sync.WaitGroup, logger *zerolog.Logger) {
	defer wg.Done()
	for event := range taskChan {
		if event.Type == UserDeviceMintEventType {
			err := c.pg.StoreVehicle(ctx, event.Data.Device.ID, int(event.Data.NFT.TokenID.Int64()))
			if err != nil {
				logger.Err(err).Msg("unable to store vehicle information")
			}
		}
		continue
	}
	logger.Info().Int("workerNum", workerNum).Msg("shutdown")
}
