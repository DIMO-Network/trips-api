package api

import (
	"math"
	"strconv"
	"time"

	"github.com/DIMO-Network/trips-api/internal/api/types"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types/pgeo"
)

type Handler struct {
	pg     *pg_store.Store
	logger *zerolog.Logger
}

func NewHandler(pgStore *pg_store.Store, logger *zerolog.Logger) *Handler {
	return &Handler{pgStore, logger}
}

const pageSize = 100

// GetVehicleTrips returns a page of the given vehicle's trips.
//
//	@Description	Lists vehicle trips.
//	@Produce		json
//	@Security		BearerAuth
//	@Param			tokenId	path		int	true	"Vehicle token id"
//	@Param			page	query		int	false	"Page of trips to retrieve. Defaults to 1."
//	@Success		200		{object}	types.VehicleTrips
//	@Router			/vehicle/{tokenId}/trips [get]
func (h *Handler) GetVehicleTrips(c *fiber.Ctx) error {
	rawTokenID := c.Params("tokenID")
	tokenID, err := strconv.Atoi(rawTokenID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse vehicle token id.")
	}

	var p Params
	err = validateQueryParams(&p, c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse query params.")
	}

	totalCount, err := models.Trips(
		models.TripWhere.VehicleTokenID.EQ(tokenID),
	).Count(c.Context(), h.pg.DB.DBS().Reader)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	start := time.Now()
	trips, err := models.Trips(
		models.TripWhere.VehicleTokenID.EQ(tokenID),
		models.TripWhere.EndTime.IsNotNull(),
		qm.OrderBy(models.TripColumns.EndTime+" DESC"),
		qm.Limit(pageSize),
		qm.Offset((p.Page-1)*pageSize),
	).All(c.Context(), h.pg.DB.DBS().Reader)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	h.logger.Info().Int("vehicleTokenId", tokenID).Str("duration", time.Since(start).String()).Msg("Ran trips query.")

	resp := types.VehicleTrips{
		Trips:       make([]types.TripDetails, len(trips)),
		CurrentPage: p.Page,
		TotalPages:  int(math.Ceil(float64(totalCount) / pageSize)),
	}

	for i, trp := range trips {
		resp.Trips[i] = types.TripDetails{
			ID: trp.ID,
			Start: types.TripStart{
				Time:              trp.StartTime,
				Location:          nullLocationToAPI(trp.StartPosition),
				EstimatedLocation: nullLocationToAPI(trp.StartPositionEstimate),
			},
			End: types.TripEnd{
				Time:     trp.EndTime.Time,
				Location: nullLocationToAPI(trp.EndPosition),
			},
			Dropped: trp.DroppedData,
		}
	}

	return c.JSON(resp)
}

func validateQueryParams(p *Params, c *fiber.Ctx) error {
	err := c.QueryParser(p)
	if err != nil {
		return err
	}

	if p.Page < 1 {
		p.Page = 1
	}
	return nil
}

type Params struct {
	Page int `query:"page"`
}

func nullLocationToAPI(l pgeo.NullPoint) *types.Location {
	if l.Valid {
		return &types.Location{Latitude: l.Y, Longitude: l.X}
	}
	return nil
}
