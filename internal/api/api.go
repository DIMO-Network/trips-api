package api

import (
	"database/sql"
	"math"
	"strconv"
	"time"

	"github.com/DIMO-Network/trips-api/internal/api/types"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types/pgeo"
)

type Handler struct {
	pg *pg_store.Store
}

type VehicleTrips struct {
	Trips       []TripDetails `json:"trips"`
	TotalPages  int           `json:"totalPages" example:"1"`
	CurrentPage int           `json:"currentPage" example:"1"`
}

type TripDetails struct {
	ID    string    `json:"id" example:"2Y83IHPItgk0uHD7hybGnA776Bo"`
	Start TripStart `json:"start"`
	End   TripEnd   `json:"end"`
}

type TripStart struct {
	Time              time.Time `json:"time"`
	Location          *Location `json:"location,omitempty"`
	EstimatedLocation *Location `json:"estimatedLocation"`
}

type TripEnd struct {
	Time     time.Time `json:"time"`
	Location *Location `json:"location,omitempty"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func NewHandler(pgStore *pg_store.Store) *Handler {
	return &Handler{pgStore}
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

	v, err := models.Vehicles(
		models.VehicleWhere.TokenID.EQ(tokenID),
		qm.Load(
			models.VehicleRels.VehicleTokenTrips,
			models.TripWhere.EndTime.IsNotNull(),
			qm.OrderBy(models.TripColumns.EndTime+" DESC"),
			qm.Limit(pageSize),
			qm.Offset((p.Page-1)*pageSize),
		),
	).One(c.Context(), h.pg.DB.DBS().Reader)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "No vehicle with that token id.")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	resp := types.VehicleTrips{
		Trips:       make([]types.TripDetails, len(v.R.VehicleTokenTrips)),
		CurrentPage: p.Page,
		TotalPages:  int(math.Ceil(float64(totalCount) / pageSize)),
	}

	for i, trp := range v.R.VehicleTokenTrips {
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
