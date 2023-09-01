package pg_handler

import (
	"github.com/DIMO-Network/trips-api/internal/services/bundlr"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/gofiber/fiber/v2"
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

	allSegments, err := models.Trips(
		models.TripWhere.UserDeviceID.EQ(deviceID),
	).All(c.Context(), h.pg.DB)
	if err != nil {
		return c.JSON(err)
	}

	return c.JSON(allSegments)
}
