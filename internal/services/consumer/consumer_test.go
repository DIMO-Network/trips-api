package consumer

import (
	"context"
	"testing"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/DIMO-Network/trips-api/internal/test"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/sqlboiler/v4/types/pgeo"
)

var (
	migrationsDirRelPath    = "../../../migrations"
	testExpectedStartAndEnd = "testExpectedStartAndEnd"
	testEstimateStart       = "testEstimateStart"
	testFailToEstimateStart = "testFailToEstimateStart"
)

var createDevice shared.CloudEvent[UserDeviceMintEvent] = shared.CloudEvent[UserDeviceMintEvent]{
	Type: UserDeviceMintEventType,
	Data: UserDeviceMintEvent{
		Device: struct {
			ID string "json:\"id\""
		}{
			ID: ksuid.New().String(),
		},
		NFT: struct {
			TokenID int "json:\"tokenId\""
		}{
			TokenID: 1,
		},
	},
}

var startLat = 40.744331740800455
var startLon = -73.98043334522801
var beginSegment shared.CloudEvent[SegmentEvent] = shared.CloudEvent[SegmentEvent]{
	Data: SegmentEvent{
		ID:       testExpectedStartAndEnd,
		DeviceID: createDevice.Data.Device.ID,
		Start: Endpoint{
			Time: time.Now().Add(-time.Hour * 3 * 24).UTC(),
			Location: &Location{
				Latitude:  startLat,
				Longitude: startLon,
			},
		},
	},
}

var endLat = 33.84805567103969
var endLon = -118.39318923141917
var completeSegment shared.CloudEvent[SegmentEvent] = shared.CloudEvent[SegmentEvent]{
	Data: SegmentEvent{
		ID:        testExpectedStartAndEnd,
		DeviceID:  createDevice.Data.Device.ID,
		Completed: true,
		End: Endpoint{
			Time: time.Now().Add(-time.Hour * 30).UTC(),
			Location: &Location{
				Latitude:  endLat,
				Longitude: endLon,
			},
		},
		Start: beginSegment.Data.Start,
	},
}

var estStartLat = 33.850422561365455
var estStartLon = -118.3962470088937
var newTrpEstimatestart shared.CloudEvent[SegmentEvent] = shared.CloudEvent[SegmentEvent]{
	Data: SegmentEvent{
		ID:       testEstimateStart,
		DeviceID: createDevice.Data.Device.ID,
		Start: Endpoint{
			Time: time.Now().Add(-time.Minute * 3).UTC(),
			Location: &Location{
				Latitude:  estStartLat,
				Longitude: estStartLon,
			},
		},
	},
}

var noEstStartLat = 33.95737251631686
var noEstStartLon = -118.44861917146383
var newTrpNoStartEstimate shared.CloudEvent[SegmentEvent] = shared.CloudEvent[SegmentEvent]{
	Data: SegmentEvent{
		ID:       testFailToEstimateStart,
		DeviceID: createDevice.Data.Device.ID,
		Start: Endpoint{
			Time: time.Now().Add(-time.Minute * 5).UTC(),
			Location: &Location{
				Latitude:  noEstStartLat,
				Longitude: noEstStartLon,
			},
		},
	},
}

func Test_CreateVehicle(t *testing.T) {
	ctx := context.Background()

	pdb := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	consumer := Consumer{
		logger: &zerolog.Logger{},
		pg: &pg.Store{
			DB: pdb,
		},
	}

	if err := consumer.VehicleEvent(ctx, createDevice); err != nil {
		t.Fatal(err)
	}

	v, _ := models.Vehicles().One(ctx, pdb.DBS().Reader)
	assert.Equal(t, v.UserDeviceID, createDevice.Data.Device.ID)
}

func Test_TripStartTripWithGeos(t *testing.T) {
	ctx := context.Background()

	pdb := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	consumer := Consumer{
		logger: &zerolog.Logger{},
		pg: &pg.Store{
			DB: pdb,
		},
	}

	if err := consumer.VehicleEvent(ctx, createDevice); err != nil {
		t.Fatal(err)
	}

	if err := consumer.ProcessSegmentEvent(ctx, beginSegment); err != nil {
		t.Fatal(err)
	}
	trp, _ := models.Trips(models.TripWhere.ID.EQ(beginSegment.Data.ID)).One(ctx, pdb.DBS().Reader)
	assert.Equal(t, trp.StartPosition.X, beginSegment.Data.Start.Location.Longitude)
	assert.Equal(t, trp.StartPosition.Y, beginSegment.Data.Start.Location.Latitude)
	assert.Equal(t, trp.StartTime, beginSegment.Data.Start.Time)

	if err := consumer.ProcessSegmentEvent(ctx, completeSegment); err != nil {
		t.Fatal(err)
	}
	err := trp.Reload(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)
	assert.Equal(t, trp.EndPosition.X, completeSegment.Data.End.Location.Longitude)
	assert.Equal(t, trp.EndPosition.Y, completeSegment.Data.End.Location.Latitude)
	assert.Equal(t, trp.EndTime.Time, completeSegment.Data.End.Time)

}

func Test_NewTripEstimateStart(t *testing.T) {
	ctx := context.Background()
	pdb := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	consumer := Consumer{
		logger: &zerolog.Logger{},
		pg: &pg.Store{
			DB: pdb,
		},
	}

	if err := consumer.VehicleEvent(ctx, createDevice); err != nil {
		t.Fatal(err)
	}

	if err := consumer.ProcessSegmentEvent(ctx, beginSegment); err != nil {
		t.Fatal(err)
	}

	if err := consumer.ProcessSegmentEvent(ctx, completeSegment); err != nil {
		t.Fatal(err)
	}

	if err := consumer.ProcessSegmentEvent(ctx, newTrpEstimatestart); err != nil {
		t.Fatal(err)
	}
	estTrp, _ := models.Trips(models.TripWhere.ID.EQ(newTrpEstimatestart.Data.ID)).One(ctx, pdb.DBS().Reader)
	assert.Equal(t, estTrp.StartPositionEstimate.X, completeSegment.Data.End.Location.Longitude)
	assert.Equal(t, estTrp.StartPositionEstimate.Y, completeSegment.Data.End.Location.Latitude)

}

func Test_NewTripDontEstimateStart(t *testing.T) {
	ctx := context.Background()
	pdb := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	consumer := Consumer{
		logger: &zerolog.Logger{},
		pg: &pg.Store{
			DB: pdb,
		},
	}

	if err := consumer.VehicleEvent(ctx, createDevice); err != nil {
		t.Fatal(err)
	}

	if err := consumer.ProcessSegmentEvent(ctx, beginSegment); err != nil {
		t.Fatal(err)
	}

	if err := consumer.ProcessSegmentEvent(ctx, completeSegment); err != nil {
		t.Fatal(err)
	}

	if err := consumer.ProcessSegmentEvent(ctx, newTrpNoStartEstimate); err != nil {
		t.Fatal(err)
	}

	newTripNoEst, _ := models.Trips(models.TripWhere.ID.EQ(newTrpNoStartEstimate.Data.ID)).One(ctx, pdb.DBS().Reader)
	assert.Equal(t, newTripNoEst.StartPositionEstimate, pgeo.NullPoint{})
	assert.Equal(t, newTripNoEst.StartPosition.X, newTrpNoStartEstimate.Data.Start.Location.Longitude)
	assert.Equal(t, newTripNoEst.StartPosition.Y, newTrpNoStartEstimate.Data.Start.Location.Latitude)
}

func Test_StartLocationNotIncludedInFirstEvent(t *testing.T) {
	ctx := context.Background()
	pdb := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	consumer := Consumer{
		logger: &zerolog.Logger{},
		pg: &pg.Store{
			DB: pdb,
		},
	}

	beginSegment.Data.Start.Latitude = nil
	beginSegment.Data.Start.Longitude = nil

	if err := consumer.VehicleEvent(ctx, createDevice); err != nil {
		t.Fatal(err)
	}

	if err := consumer.ProcessSegmentEvent(ctx, beginSegment); err != nil {
		t.Fatal(err)
	}

	trp, _ := models.Trips(models.TripWhere.ID.EQ(beginSegment.Data.ID)).One(ctx, pdb.DBS().Reader)
	assert.False(t, trp.StartPosition.Valid)
	assert.Equal(t, trp.StartTime, beginSegment.Data.Start.Time)

	if err := consumer.ProcessSegmentEvent(ctx, completeSegment); err != nil {
		t.Fatal(err)
	}
	err := trp.Reload(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)
	assert.Equal(t, trp.StartPositionEstimate.X, *completeSegment.Data.Start.Longitude)
	assert.Equal(t, trp.StartPositionEstimate.Y, *completeSegment.Data.Start.Latitude)
	assert.Equal(t, trp.EndPosition.X, *completeSegment.Data.End.Longitude)
	assert.Equal(t, trp.EndPosition.Y, *completeSegment.Data.End.Latitude)
	assert.Equal(t, trp.EndTime.Time, completeSegment.Data.End.Time)

}
