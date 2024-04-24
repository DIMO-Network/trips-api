package helper

import (
	"time"

	"github.com/DIMO-Network/trips-api/models"
	"github.com/umahmood/haversine"
	"github.com/volatiletech/sqlboiler/v4/types/pgeo"
)

type VehicleTrips struct {
	Trips       []*TripDetails `json:"trips"`
	TotalPages  int            `json:"totalPages" example:"1"`
	CurrentPage int            `json:"currentPage" example:"1"`
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

func TripGeos(trip *models.Trip, details *TripDetails, prevTripEnd pgeo.NullPoint) *TripDetails {
	if trip.StartPosition.Valid {
		details.Start.Actual.Latitude = trip.StartPosition.Y
		details.Start.Actual.Longitude = trip.StartPosition.X

		if prevTripEnd.Valid {
			if distMiles, _ := haversine.Distance(
				haversine.Coord{Lat: prevTripEnd.Y, Lon: prevTripEnd.X},
				haversine.Coord{Lat: details.Start.Actual.Latitude, Lon: details.Start.Actual.Longitude},
			); distMiles < .25 {
				// if the trip start location is less than 0.25 miles from the previous trip end location
				// user prior trip end as the new start estimate
				// NOTE that this does not tell us what the estimated start time is
				details.Start.Estimate = &PointTime{
					Longitude: prevTripEnd.X,
					Latitude:  prevTripEnd.Y,
				}
			}
		}
	}

	if trip.EndPosition.Valid {
		details.End.Actual.Latitude = trip.EndPosition.Y
		details.End.Actual.Longitude = trip.EndPosition.X
	}

	return details
}
