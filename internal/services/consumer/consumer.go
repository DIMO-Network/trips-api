package consumer

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/trips-api/internal/geo"
	"github.com/DIMO-Network/trips-api/internal/services/bundlr"
	es_store "github.com/DIMO-Network/trips-api/internal/services/es"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Endpoint struct {
	Time     time.Time `json:"time"`
	Location *Location `json:"location"`
}

type SegmentEvent struct {
	ID        string   `json:"id"`
	DeviceID  string   `json:"deviceId"`
	Completed bool     `json:"completed"`
	Start     Endpoint `json:"start"`
	End       Endpoint `json:"end"`
}

type UserDeviceMintEvent struct {
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"userId"`
	Device    struct {
		ID string `json:"id"`
	} `json:"device"`
	NFT struct {
		TokenID int `json:"tokenId"`
	} `json:"nft"`
}

const UserDeviceMintEventType = "com.dimo.zone.device.mint"

func New(es *es_store.Client, bundlrClient *bundlr.Client, pg *pg_store.Store, logger *zerolog.Logger, dataFetchEnabled bool, workerCount int, bundlrEnabled bool) *Consumer {
	return &Consumer{logger, es, pg, bundlrClient, dataFetchEnabled, workerCount, bundlrEnabled}
}

func (c *Consumer) ProcessSegmentEvent(ctx context.Context, event shared.CloudEvent[SegmentEvent]) error {
	if event.Data.Completed {
		return c.CompleteSegment(ctx, event)
	}
	return c.BeginSegment(ctx, event)
}

func (c *Consumer) BeginSegment(ctx context.Context, event shared.CloudEvent[SegmentEvent]) error {
	veh, err := models.Vehicles(
		models.VehicleWhere.UserDeviceID.EQ(event.Data.DeviceID),
		qm.Load(
			models.VehicleRels.VehicleTokenTrips,
			models.TripWhere.EndTime.IsNotNull(),
			qm.OrderBy(models.TripColumns.EndTime+" DESC"),
			qm.Limit(1),
		),
	).One(ctx, c.pg.DB.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to find vehicle %s: %w", event.Subject, err)
		}
		return err
	}

	segment := models.Trip{
		ID:             event.Data.ID,
		VehicleTokenID: veh.TokenID,
		StartTime:      event.Data.Start.Time,
		StartPosition:  nullLocationToDB(event.Data.Start.Location),
	}

	if segment.StartPosition.Valid && len(veh.R.VehicleTokenTrips) > 0 {
		if lastLoc := veh.R.VehicleTokenTrips[0].EndPosition; lastLoc.Valid && geo.InterpolateTripStart(lastLoc.Point, segment.StartPosition.Point) {
			segment.StartPositionEstimate = lastLoc
		}
	}

	return segment.Insert(ctx, c.pg.DB.DBS().Writer, boil.Infer())
}

func (c *Consumer) CompleteSegment(ctx context.Context, event shared.CloudEvent[SegmentEvent]) error {
	segment, err := models.Trips(
		models.TripWhere.ID.EQ(event.Data.ID),
	).One(ctx, c.pg.DB.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no segment with id  %s: %w", event.Data.ID, err)
		}
		return fmt.Errorf("error fetching segment %s: %w", event.Data.ID, err)
	}
	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		return fmt.Errorf("couldn't produce random key: %w", err)
	}

	segment.EncryptionKey = null.BytesFrom(encryptionKey)
	segment.EndTime = null.TimeFrom(event.Data.End.Time)
	segment.EndPosition = nullLocationToDB(event.Data.End.Location)

	if !segment.StartPosition.Valid && event.Data.Start.Location != nil {
		veh, err := models.Vehicles(
			models.VehicleWhere.UserDeviceID.EQ(event.Data.DeviceID),
			qm.Load(
				models.VehicleRels.VehicleTokenTrips,
				models.TripWhere.EndTime.IsNotNull(),
				qm.OrderBy(models.TripColumns.EndTime+" DESC"),
				qm.Limit(1),
			),
		).One(ctx, c.pg.DB.DBS().Reader)
		if err != nil {
			c.logger.Error().Err(err).Msg("failed to find vehicle")
		}

		switch {
		case len(veh.R.VehicleTokenTrips) > 0:
			estLoc := nullLocationToDB(event.Data.Start.Location)
			if lastLoc := veh.R.VehicleTokenTrips[0].EndPosition; lastLoc.Valid && geo.InterpolateTripStart(lastLoc.Point, estLoc.Point) {
				segment.StartPositionEstimate = lastLoc
			}
		default:
			segment.StartPositionEstimate = nullLocationToDB(event.Data.Start.Location)
		}
	}

	if c.dataFetchEnabled {
		response, err := c.es.FetchData(ctx, event.Data.DeviceID, segment.StartTime, event.Data.End.Time)
		if err != nil {
			return fmt.Errorf("call to Elasticsearch failed: %w", err)
		}

		dataItem, err := c.bundlr.PrepareData(response, encryptionKey, segment.VehicleTokenID, segment.StartTime, event.Data.End.Time)
		if err != nil {
			return fmt.Errorf("assembly for Bundlr failed: %w", err)
		}

		if c.bundlrEnabled {
			if err := c.bundlr.Upload(dataItem); err != nil {
				return fmt.Errorf("bundlr upload failed: %w", err)
			}
		}

		segment.BundlrID = null.StringFrom(dataItem.Id.Base64())
		c.logger.Info().Msgf("https://devnet.bundlr.network/%s", segment.BundlrID.String)
	}

	if _, err := segment.Update(ctx, c.pg.DB.DBS().Writer,
		boil.Whitelist(
			models.TripColumns.EncryptionKey,
			models.TripColumns.EndTime,
			models.TripColumns.BundlrID,
			models.TripColumns.EndPosition,
			models.TripColumns.StartPositionEstimate),
	); err != nil {
		return fmt.Errorf("error updating segment %s: %w", event.Data.ID, err)
	}
	return nil
}

func (c *Consumer) VehicleEvent(ctx context.Context, event shared.CloudEvent[UserDeviceMintEvent]) error {
	if event.Type == UserDeviceMintEventType {
		if err := c.pg.StoreVehicle(ctx, event.Data.Device.ID, event.Data.NFT.TokenID); err != nil {
			return fmt.Errorf("failed to store vehicle: %w", err)
		}

		c.logger.Debug().Int("tokenId", event.Data.NFT.TokenID).Str("userDeviceId", event.Data.Device.ID).Msg("Id mapping stored.")
		return nil
	}
	return nil
}

func nullLocationToDB(l *Location) pgeo.NullPoint {
	if l == nil {
		return pgeo.NullPoint{}
	}
	return pgeo.NewNullPoint(pgeo.NewPoint(l.Longitude, l.Latitude), true)
}
