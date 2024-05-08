package geo

import (
	"github.com/umahmood/haversine"
	"github.com/volatiletech/sqlboiler/v4/types/pgeo"
)

const InterpolationThresholdMiles = 0.25

func InterpolateTripStart(prevTripEnd, newTripStart pgeo.Point) bool {
	distMiles, _ := haversine.Distance(
		haversine.Coord{
			Lat: prevTripEnd.Y,
			Lon: prevTripEnd.X,
		},
		haversine.Coord{
			Lat: newTripStart.Y,
			Lon: newTripStart.X,
		},
	)

	return distMiles < InterpolationThresholdMiles
}
