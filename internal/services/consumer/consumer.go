package consumer

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"math/big"
	"time"

	pb_devices "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/kafka"
	"github.com/DIMO-Network/trips-api/internal/services/bundlr"
	es_store "github.com/DIMO-Network/trips-api/internal/services/es"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
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

type VehicleNodeMinted struct {
	TokenId *big.Int
	Owner   common.Address
}

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

	v, err := getVehicleData(ctx, c.pg.DB, c.grpc, event.Data.DeviceID)
	if err != nil {
		return err
	}

	dataItem, nonce, err := c.bundlr.PrepareData(response, v.EncryptionKey, v.UserDeviceID, event.Data.Start.Time, event.Data.End.Time)
	if err != nil {
		return err
	}

	segment := models.Trip{
		VehicleTokenID: types.NullDecimal(v.TokenID),
		ID:             event.Data.DeviceID,
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

func (c *Consumer) VehicleEvent(ctx context.Context, e *shared.CloudEvent[VehicleNodeMinted]) error {
	// event := e.(*shared.CloudEvent[VehicleNodeMinted])

	// populate vehicles table: token_id, user_device_id, encryption_key
	// look at devices-api
	return nil
}

func getVehicleData(ctx context.Context, db *sql.DB, grpc pb_devices.UserDeviceServiceClient, deviceID string) (*models.Vehicle, error) {
	v, err := models.Vehicles(models.VehicleWhere.UserDeviceID.EQ(deviceID)).One(ctx, db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			userDevice, err := grpc.GetUserDevice(ctx, &pb_devices.GetUserDeviceRequest{
				Id: deviceID,
			})
			if err != nil {
				return nil, err
			}

			encryptionKey := make([]byte, 32)
			if _, err := rand.Read(encryptionKey); err != nil {
				return nil, err
			}

			v := models.Vehicle{
				TokenID:       types.NewDecimal(new(decimal.Big).SetUint64(*userDevice.TokenId)),
				UserDeviceID:  deviceID,
				EncryptionKey: encryptionKey,
			}
			if err := v.Insert(ctx, db, boil.Whitelist(models.VehicleColumns.UserDeviceID, models.VehicleColumns.TokenID, models.VehicleColumns.EncryptionKey)); err != nil {
				return nil, err
			}
			return &v, nil
		}
		return nil, err
	}
	return v, err
}
