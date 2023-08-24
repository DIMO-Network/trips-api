package pg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/ericlagergren/decimal"
	"github.com/tidwall/gjson"
	"github.com/uber/h3-go/v3"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

// Store connected to postgres db containing trip information and validates user
type Store struct {
	db                 *sql.DB
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
		db:                 db,
		DevicesAPIGRPCAddr: settings.DevicesAPIGRPCAddr,
	}, nil
}

func (s *Store) StoreSegmentMetadata(ctx context.Context, vehicleTokenId uint64, encryptionKey string, response []byte, bundlrID string) error {
	n := gjson.GetBytes(response, "hits.hits.#").Int()
	startLat := gjson.GetBytes(response, "hits.hits.0._source.data.latitude").Float()
	startLon := gjson.GetBytes(response, "hits.hits.0._source.data.longitude").Float()
	startTime := gjson.GetBytes(response, "hits.hits.0._source.data.timestamp").Time()
	endLat := gjson.GetBytes(response, fmt.Sprintf("hits.hits.%d._source.data.latitude", n-1)).Float()
	endLon := gjson.GetBytes(response, fmt.Sprintf("hits.hits.%d._source.data.longitude", n-1)).Float()
	endTime := gjson.GetBytes(response, fmt.Sprintf("hits.hits.%d._source.data.timestamp", n-1)).Time()

	startHex := h3.FromGeo(h3.GeoCoord{Latitude: startLat, Longitude: startLon}, 6)
	endHex := h3.FromGeo(h3.GeoCoord{Latitude: endLat, Longitude: endLon}, 6)

	trp := models.Trip{
		VehicleTokenID: types.NewDecimal(decimal.New(int64(vehicleTokenId)+startTime.Unix(), 0)),
		Start:          startTime,
		StartHex:       int(startHex),
		End:            endTime,
		EndHex:         int(endHex),
		BunldrID:       bundlrID,
	}

	return trp.Insert(ctx, s.db, boil.Infer())
}
