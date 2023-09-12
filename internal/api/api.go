package api

import (
	"context"

	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/DIMO-Network/trips-api/internal/utils"
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
func (h *Handler) AllSegments(c *fiber.Ctx) error {
	deviceID := c.Params("id")
	segments, err := h.fetchSegments(c.Context(), &deviceID, nil)
	if err != nil {
		return c.JSON(err)
	}

	return c.JSON(segments)
}

// SingleSegment godoc
// @Description returns metadata for segment associated with vehicle and tripID
// @Tags        vehicles-segments
// @Produce     json
// @Security    BearerAuth
// @Router      /vehicles/:id/segments/:tripID [get]
func (h *Handler) SingleSegment(c *fiber.Ctx) error {
	deviceID := c.Params("id")
	tripID := c.Params("tripID")

	segments, err := h.fetchSegments(c.Context(), &deviceID, &tripID)
	if err != nil {
		return c.JSON(err)
	}

	return c.JSON(segments)
}

func (h *Handler) fetchSegments(ctx context.Context, deviceID, tripID *string) ([]utils.SegmentSummary, error) {
	mods := []qm.QueryMod{
		models.VehicleWhere.UserDeviceID.EQ(*deviceID),
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
	}

	if tripID != nil {
		mods = append(mods, qm.Load(models.VehicleRels.VehicleTokenTrips, models.TripWhere.ID.EQ(*tripID)))
	}

	segments, err := models.Vehicles(mods...).One(ctx, h.pg.DB)
	if err != nil {
		return nil, err
	}

	segmentsSummary := make([]utils.SegmentSummary, 0)
	for _, s := range segments.R.VehicleTokenTrips {
		segmentsSummary = append(segmentsSummary, utils.SegmentSummary{
			TripID:   s.ID,
			DeviceID: *deviceID,
			BundlrID: s.BundlrID.String,
			Start: utils.PointTime{
				Point: utils.Point{
					Latitude:  s.StartPosition.Y,
					Longitude: s.StartPosition.X,
				},
				Time: s.StartTime,
			},
			End: utils.PointTime{
				Point: utils.Point{
					Latitude:  s.EndPosition.Y,
					Longitude: s.EndPosition.X,
				},
				Time: s.EndTime.Time,
			},
		})
	}

	return segmentsSummary, nil
}
