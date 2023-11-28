package api

import (
	"database/sql"
	"math"
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

const pageSize = 100

// GetVehicleTrips returns a page of the given vehicle's trips.
//
//	@Description	Lists vehicle trips.
//	@Produce		json
//	@Security		BearerAuth
//	@Param			tokenId	path		int	true	"Vehicle token id"
//	@Param			page	query		int	false	"Page of trips to retrieve. Defaults to 1."
//	@Success		200		{object}	api.VehicleTripsResp
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

	mods := []qm.QueryMod{
		qm.OrderBy(models.TripColumns.EndTime + " DESC"),
		qm.Limit(pageSize),
		qm.Offset((p.Page - 1) * pageSize),
	}

	v, err := models.Vehicles(
		models.VehicleWhere.TokenID.EQ(tokenID),
		qm.Load(
			models.VehicleRels.VehicleTokenTrips,
			mods...,
		),
	).One(c.Context(), h.pg.DB.DBS().Reader)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "No vehicle with that token id.")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	out := VehicleTripsResp{
		Trips:       make([]VehicleTripResp, len(v.R.VehicleTokenTrips)),
		CurrentPage: p.Page,
		TotalPages:  int(math.Ceil(float64(totalCount) / pageSize)),
	}

	for i, tripRow := range v.R.VehicleTokenTrips {
		tripOut := VehicleTripResp{
			ID: tripRow.ID,
			Start: Endpoint{
				Time: tripRow.StartTime,
			},
		}

		if tripRow.EndTime.Valid {
			tripOut.End = &Endpoint{
				Time: tripRow.EndTime.Time,
			}
		}

		out.Trips[i] = tripOut
	}

	return c.JSON(out)
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
	Page int `json:"page"`
}

type VehicleTripsResp struct {
	Trips       []VehicleTripResp `json:"trips"`
	TotalPages  int               `json:"totalPages" example:"1"`
	CurrentPage int               `json:"currentPage" example:"1"`
}

type VehicleTripResp struct {
	ID    string    `json:"id" example:"2Y83IHPItgk0uHD7hybGnA776Bo"`
	Start Endpoint  `json:"start"`
	End   *Endpoint `json:"end"`
}

type Endpoint struct {
	Time time.Time `json:"time" example:"2023-05-04T09:00:00Z"`
}
