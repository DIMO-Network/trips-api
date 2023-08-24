package pg

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"

	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/ericlagergren/decimal"
	"github.com/segmentio/ksuid"
	"github.com/tidwall/gjson"
	"github.com/uber/h3-go/v3"
	"github.com/volatiletech/null/v8"
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
	// there should be a validation step in the processor to make sure we aren't saving an zerotimes/ coords
	n := gjson.GetBytes(response, "hits.hits.#").Int()
	startLat := gjson.GetBytes(response, "hits.hits.0._source.data.latitude").Float()
	startLon := gjson.GetBytes(response, "hits.hits.0._source.data.longitude").Float()
	startTime := gjson.GetBytes(response, "hits.hits.0._source.data.timestamp").Time()
	endLat := gjson.GetBytes(response, fmt.Sprintf("hits.hits.%d._source.data.latitude", n-1)).Float()
	endLon := gjson.GetBytes(response, fmt.Sprintf("hits.hits.%d._source.data.longitude", n-1)).Float()
	endTime := gjson.GetBytes(response, fmt.Sprintf("hits.hits.%d._source.data.timestamp", n-1)).Time()

	startHex := h3.FromGeo(h3.GeoCoord{Latitude: startLat, Longitude: startLon}, 6)
	endHex := h3.FromGeo(h3.GeoCoord{Latitude: endLat, Longitude: endLon}, 6)

	fmt.Println(int(startHex), startHex, endHex, startTime, endTime, bundlrID)

	trp := models.Fulltrip{
		TripID:         ksuid.New().String(),
		TripStart:      null.TimeFrom(startTime),
		TripEnd:        null.TimeFrom(endTime),
		VehicleTokenID: types.NewNullDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(int64(vehicleTokenId)), 0)),
		StartHex:       null.IntFrom(int(startHex)),
		EndHex:         null.IntFrom(int(endHex)),
		BundlrID:       null.StringFrom(bundlrID),
	}

	return trp.Insert(ctx, s.db, boil.Infer())
}