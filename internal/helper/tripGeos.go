package tripgeos

import (
	"time"

	"github.com/umahmood/haversine"
	"github.com/volatiletech/sqlboiler/v4/types/pgeo"
)

type VehicleTrips struct {
	Trips       []TripDetails `json:"trips"`
	TotalPages  int           `json:"totalPages" example:"1"`
	CurrentPage int           `json:"currentPage" example:"1"`
}

type TripDetails struct {
	TripID string    `json:"id" example:"2Y83IHPItgk0uHD7hybGnA776Bo"`
	Start  TripEvent `json:"start"`
	End    TripEvent `json:"end"`
}

type TripEvent struct {
	Estimate *PointTime `json:"estimate,omitempty"`
	Actual   PointTime  `json:"actual"`
}

type PointTime struct {
	Time      time.Time `json:"time,omitempty" example:"2023-05-04T09:00:00Z"`
	Latitude  float64   `json:"latitude,omitempty" example:"37.7749"`
	Longitude float64   `json:"longitude,omitempty" example:"-122.4194"`
}

func InterpolateTripStart(prevTripEnd, newTripStart pgeo.NullPoint) bool {
	if !prevTripEnd.Valid || !newTripStart.Valid {
		return false
	}

	if distMiles, _ := haversine.Distance(
		haversine.Coord{
			Lat: prevTripEnd.Y,
			Lon: prevTripEnd.X,
		},
		haversine.Coord{
			Lat: newTripStart.Y,
			Lon: newTripStart.X,
		},
	); distMiles < 0.25 {
		return true
	}

	return false
}
