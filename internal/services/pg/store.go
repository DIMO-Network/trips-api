package pg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
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

func (s Store) StoreVehicle(ctx context.Context, userDeviceID string, tokenID int) error {
	v := models.Vehicle{
		UserDeviceID: userDeviceID,
		TokenID:      tokenID,
	}

	return v.Insert(ctx, s.DB, boil.Infer())
}
