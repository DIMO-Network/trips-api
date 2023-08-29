package pg

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"

	pb_devices "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

// Store connected to postgres db containing trip information and validates user
type Store struct {
	DB                 *sql.DB
	DevicesAPIGRPCAddr string
}

func New(settings *config.Settings) (*Store, error) {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		settings.DBHost,
		settings.DBPort,
		settings.DBUser,
		settings.DBPassword,
		settings.DBName,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	return &Store{
		DB:                 db,
		DevicesAPIGRPCAddr: settings.DevicesAPIGRPCAddr,
	}, nil
}

func (s Store) GetOrGenerateEncryptionKey(ctx context.Context, deviceID string, grpc pb_devices.UserDeviceServiceClient) (*models.Vehicle, error) {
	vehicle, err := models.Vehicles(models.VehicleWhere.UserDeviceID.EQ(deviceID)).One(ctx, s.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			userDevice, err := grpc.GetUserDevice(ctx, &pb_devices.GetUserDeviceRequest{
				Id: deviceID,
			})
			if err != nil {
				return nil, err
			}

			return s.GenerateKey(ctx, deviceID, *userDevice.TokenId, grpc)
		}
		return nil, err
	}
	return vehicle, err
}

func (s Store) GenerateKey(ctx context.Context, deviceID string, tokenID uint64, grpc pb_devices.UserDeviceServiceClient) (*models.Vehicle, error) {
	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		return nil, err
	}

	v := models.Vehicle{
		TokenID:       types.NewDecimal(new(decimal.Big).SetUint64(tokenID)),
		UserDeviceID:  deviceID,
		EncryptionKey: encryptionKey,
	}
	if err := v.Insert(ctx, s.DB, boil.Whitelist(models.VehicleColumns.UserDeviceID, models.VehicleColumns.TokenID, models.VehicleColumns.EncryptionKey)); err != nil {
		return nil, err
	}
	return &v, nil

}
