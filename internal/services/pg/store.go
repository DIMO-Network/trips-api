package pg

import (
	"context"

	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// Store connected to postgres db containing trip information and validates user
type Store struct {
	DB db.Store
}

func New(settings *config.Settings) (*Store, error) {
	dbs := db.NewDbConnectionFromSettings(context.Background(), &settings.DB, true)

	return &Store{
		DB: dbs,
	}, nil
}

func (s Store) StoreVehicle(ctx context.Context, userDeviceID string, tokenID int) error {
	v := models.Vehicle{
		UserDeviceID: userDeviceID,
		TokenID:      tokenID,
	}

	return v.Upsert(ctx, s.DB.DBS().Writer, false, []string{models.VehicleColumns.TokenID}, boil.None(), boil.Infer())
}
