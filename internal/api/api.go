package api

import (
	"database/sql"
	"strconv"
	"time"

	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Handler struct {
	pg *pg_store.Store
}

func NewHandler(pgStore *pg_store.Store) *Handler {
	return &Handler{pgStore}
}

const maxPageSize = 100

// Segments godoc
// @Description details for all segments associated with vehicles
// @Tags        vehicles-segments
// @Produce     json
// @Security    BearerAuth
// @Router      /vehicle/:tokenID/trips [get]
func (h *Handler) Segments(c *fiber.Ctx) error {
	rawTokenID := c.Params("tokenID")
	tokenID, err := strconv.Atoi(rawTokenID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse vehicle token id.")
	}

	v, err := models.Vehicles(
		models.VehicleWhere.TokenID.EQ(tokenID),
		qm.Load(
			models.VehicleRels.VehicleTokenTrips,
			qm.OrderBy(models.TripColumns.EndTime+" DESC"),
			qm.Limit(maxPageSize),
		),
	).One(c.Context(), h.pg.DB.DBS().Reader)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "No vehicle with that token id.")
		}
	}

	trips := v.R.VehicleTokenTrips

	out := VehicleTripsResp{
		Trips: make([]VehicleTripResp, len(trips)),
	}

	for i, t := range v.R.VehicleTokenTrips {
		out.Trips[i] = VehicleTripResp{
			Start: PointTime{
				Time: t.StartTime,
				Location: Location{
					Latitude:  t.StartPosition.Y,
					Longitude: t.StartPosition.X,
				},
			},
			End: PointTime{
				Time: t.EndTime.Time,
				Location: Location{
					Latitude:  t.EndPosition.Y,
					Longitude: t.EndPosition.X,
				},
			},
		}
	}

	return c.JSON(out)
}

type VehicleTripsResp struct {
	Trips []VehicleTripResp `json:"trips"`
}

type VehicleTripResp struct {
	Start PointTime `json:"start"`
	End   PointTime `json:"end"`
}

type PointTime struct {
	Time     time.Time `json:"time"`
	Location Location  `json:"location"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
