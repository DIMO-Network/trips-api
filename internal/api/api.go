package api

import (
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

// Segments godoc
// @Description details for all segments associated with vehicles
// @Tags        vehicles-segments
// @Produce     json
// @Security    BearerAuth
// @Router      /vehicles/:id/segments [get]
func (h *Handler) Segments(c *fiber.Ctx) error {
	deviceID := c.Params("id")
	segments, err := models.Vehicles(
		models.VehicleWhere.UserDeviceID.EQ(deviceID),
		qm.Load(
			models.VehicleRels.VehicleTokenTrips,
			qm.Select(
				models.TripColumns.ID,
				models.TripColumns.VehicleTokenID,
				models.TripColumns.BundlrID,
				models.TripColumns.StartTime,
				models.TripColumns.StartPosition,
				models.TripColumns.EndTime,
				models.TripColumns.EndPosition,
			),
			qm.OrderBy(models.TripColumns.EndTime+" DESC")),
	).One(c.Context(), h.pg.DB)
	if err != nil {
		return c.JSON(err)
	}

	// returns null for encryption key
	return c.JSON(segments.R.VehicleTokenTrips)
}
