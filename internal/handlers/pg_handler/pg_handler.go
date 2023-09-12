package pg_handler

import (
	"github.com/DIMO-Network/trips-api/internal/services/bundlr"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Handler struct {
	pg     *pg_store.Store
	bundlr *bundlr.Client
}

func New(pgStore *pg_store.Store, bundlrClient *bundlr.Client) *Handler {
	return &Handler{pgStore, bundlrClient}
}

// Segments godoc
// @Description details for all segments associated with device
// @Tags        user-segments
// @Produce     json
// @Security    BearerAuth
// @Router      /devices/:id/segments [get]
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
			)),
		qm.OrderBy(models.TripColumns.EndTime+" DESC"),
	).One(c.Context(), h.pg.DB)
	if err != nil {
		return c.JSON(err)
	}

	// returns null for cols not in select
	return c.JSON(segments.R.VehicleTokenTrips)
}
