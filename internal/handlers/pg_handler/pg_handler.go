package pg_handler

import (
	"context"

	"github.com/DIMO-Network/trips-api/internal/helpers"
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
func (h *Handler) AllSegments(c *fiber.Ctx) error {
	deviceID := c.Params("id")

	segments, err := h.fetchSegments(c.Context(), &deviceID, nil)
	if err != nil {
		return c.JSON(err)
	}

	return c.JSON(segments)
}

// SingleSegment godoc
// @Description returns metadata for segment associated with device and tripID
// @Tags        user-segments
// @Produce     json
// @Security    BearerAuth
// @Router      /devices/:id/segments/:tripID [get]
func (h *Handler) SingleSegment(c *fiber.Ctx) error {
	deviceID := c.Params("id")
	tripID := c.Params("tripID")

	segments, err := h.fetchSegments(c.Context(), &deviceID, &tripID)
	if err != nil {
		return c.JSON(err)
	}

	return c.JSON(segments)
}

func (h *Handler) fetchSegments(ctx context.Context, deviceID, tripID *string) ([]helpers.SegmentSummary, error) {
	mods := []qm.QueryMod{
		models.VehicleWhere.UserDeviceID.EQ(*deviceID),
		qm.Load(models.VehicleRels.VehicleTokenTrips),
	}

	if tripID != nil {
		mods = append(mods, qm.Load(models.VehicleRels.VehicleTokenTrips, models.TripWhere.ID.EQ(*tripID)))
	}

	segments, err := models.Vehicles(mods...).One(ctx, h.pg.DB)
	if err != nil {
		return nil, err
	}

	segmentsSummary := make([]helpers.SegmentSummary, 0)
	for _, s := range segments.R.VehicleTokenTrips {
		segmentsSummary = append(segmentsSummary, helpers.SegmentSummary{
			TripID:   s.ID,
			DeviceID: *deviceID,
			BundlrID: s.BundlrID.String,
			Start: helpers.PointTime{
				Point: helpers.Point{
					Latitude:  s.StartPosition.Y,
					Longitude: s.StartPosition.X,
				},
				Time: s.Start,
			},
			End: helpers.PointTime{
				Point: helpers.Point{
					Latitude:  s.EndPosition.Y,
					Longitude: s.EndPosition.X,
				},
				Time: s.End.Time,
			},
		})
	}

	return segmentsSummary, nil
}
