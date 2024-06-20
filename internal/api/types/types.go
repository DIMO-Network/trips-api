package types

import "time"

type VehicleTrips struct {
	Trips       []TripDetails `json:"trips"`
	TotalPages  int           `json:"totalPages" example:"1"`
	CurrentPage int           `json:"currentPage" example:"1"`
}

type TripDetails struct {
	ID      string    `json:"id" example:"2Y83IHPItgk0uHD7hybGnA776Bo"`
	Start   TripStart `json:"start"`
	End     TripEnd   `json:"end"`
	Dropped bool      `json:"droppedData"`
}

type TripStart struct {
	Time              time.Time `json:"time"`
	Location          *Location `json:"location,omitempty"`
	EstimatedLocation *Location `json:"estimatedLocation,omitempty"`
}

type TripEnd struct {
	Time     time.Time `json:"time"`
	Location *Location `json:"location,omitempty"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
