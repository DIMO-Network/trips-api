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
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var (
	migrationsDirRelPath = "../../../migrations"
)

var createDevice shared.CloudEvent[UserDeviceMintEvent] = shared.CloudEvent[UserDeviceMintEvent]{
	Type: UserDeviceMintEventType,
	Data: UserDeviceMintEvent{
		Device: struct {
			ID string `json:"id"`
		}{
			ID: ksuid.New().String(),
		},
		NFT: struct {
			TokenID int `json:"tokenId"`
		}{
			TokenID: 1,
		},
	},
}

var segment1 shared.CloudEvent[SegmentEvent] = shared.CloudEvent[SegmentEvent]{
	Data: SegmentEvent{
		ID:       ksuid.New().String(),
		DeviceID: createDevice.Data.Device.ID,
		Start: Endpoint{
			Time: time.Date(2023, 8, 16, 12, 15, 2, 0, time.UTC),
			Location: &Location{
				Latitude:  40.744331740800455,
				Longitude: -73.98043334522801,
			},
		},
		End: Endpoint{
			Time: time.Date(2023, 8, 18, 8, 15, 2, 0, time.UTC),
			Location: &Location{
				Latitude:  33.84805567103969,
				Longitude: -118.39318923141917,
			},
		},
	},
}

var segment2 shared.CloudEvent[SegmentEvent] = shared.CloudEvent[SegmentEvent]{
	Data: SegmentEvent{
		ID:       ksuid.New().String(),
		DeviceID: createDevice.Data.Device.ID,
		Start: Endpoint{
			Time: time.Date(2023, 8, 18, 8, 18, 2, 0, time.UTC),
			Location: &Location{
				Latitude:  33.850422561365455,
				Longitude: -118.3962470088937,
			},
		},
		End: Endpoint{
			Time: time.Date(2023, 8, 18, 8, 25, 2, 0, time.UTC),
			Location: &Location{
				Latitude:  33.8544585026455,
				Longitude: -118.39821832237583,
			},
		},
	},
}

var segment3 shared.CloudEvent[SegmentEvent] = shared.CloudEvent[SegmentEvent]{
	Data: SegmentEvent{
		ID:       ksuid.New().String(),
		DeviceID: createDevice.Data.Device.ID,
		Start: Endpoint{
			Time: time.Date(2023, 8, 18, 8, 18, 2, 0, time.UTC),
			Location: &Location{
				Latitude:  33.95737251631686,
				Longitude: -118.44861917146383,
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

// First trip a user takes
// Includes geo data
func Test_TripWithGeos(t *testing.T) {
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

	if err := consumer.ProcessSegmentEvent(ctx, segment1); err != nil {
		t.Fatal(err)
	}
	trp, _ := models.Trips(models.TripWhere.ID.EQ(segment1.Data.ID)).One(ctx, pdb.DBS().Reader)
	assert.Equal(t, trp.StartPosition.X, segment1.Data.Start.Location.Longitude)
	assert.Equal(t, trp.StartPosition.Y, segment1.Data.Start.Location.Latitude)
	assert.Equal(t, trp.StartPosition.Y, segment1.Data.Start.Location.Latitude)
	assert.Equal(t, trp.StartTime, segment1.Data.Start.Time)
	assert.Equal(t, false, trp.DroppedData)

	segment1.Data.Completed = true
	if err := consumer.ProcessSegmentEvent(ctx, segment1); err != nil {
		t.Fatal(err)
	}
	err := trp.Reload(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)
	assert.Equal(t, trp.EndPosition.X, segment1.Data.End.Location.Longitude)
	assert.Equal(t, trp.EndPosition.Y, segment1.Data.End.Location.Latitude)
	assert.Equal(t, trp.EndTime.Time, segment1.Data.End.Time)
	assert.Equal(t, false, trp.DroppedData)

}

// New user trip (has a prev trip)
// Current trip includes start geo
// Interpolate start estimate based on end of last trip
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

	trp1 := models.Trip{
		ID:             ksuid.New().String(),
		VehicleTokenID: createDevice.Data.NFT.TokenID,
		StartTime:      segment1.Data.Start.Time,
		EndTime:        null.TimeFrom(segment1.Data.End.Time),
		StartPosition:  nullLocationToDB(segment1.Data.Start.Location),
		EndPosition:    nullLocationToDB(segment1.Data.End.Location),
	}

	if err := trp1.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	segment2.Data.Completed = false
	if err := consumer.ProcessSegmentEvent(ctx, segment2); err != nil {
		t.Fatal(err)
	}
	estTrp, _ := models.Trips(models.TripWhere.ID.EQ(segment2.Data.ID)).One(ctx, pdb.DBS().Reader)
	assert.Equal(t, estTrp.StartPositionEstimate.X, segment1.Data.End.Location.Longitude)
	assert.Equal(t, estTrp.StartPositionEstimate.Y, segment1.Data.End.Location.Latitude)
	assert.Equal(t, estTrp.DroppedData, true)

}

// New user trip (has a prev trip)
// Current trip includes start geo
// Do not interpolate start estimate bc new start loc is too far from prev trip end
// Still indicate that we have dropped data
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

	trp1 := models.Trip{
		ID:             ksuid.New().String(),
		VehicleTokenID: createDevice.Data.NFT.TokenID,
		StartTime:      segment1.Data.Start.Time,
		EndTime:        null.TimeFrom(segment1.Data.End.Time),
		StartPosition:  nullLocationToDB(segment1.Data.Start.Location),
		EndPosition:    nullLocationToDB(segment1.Data.End.Location),
	}

	if err := trp1.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	if err := consumer.ProcessSegmentEvent(ctx, segment3); err != nil {
		t.Fatal(err)
	}

	newTripNoEst, _ := models.Trips(models.TripWhere.ID.EQ(segment3.Data.ID)).One(ctx, pdb.DBS().Reader)
	assert.False(t, newTripNoEst.StartPositionEstimate.Valid)
	assert.Equal(t, newTripNoEst.StartPosition.X, segment3.Data.Start.Location.Longitude)
	assert.Equal(t, newTripNoEst.StartPosition.Y, segment3.Data.Start.Location.Latitude)
	assert.Equal(t, true, newTripNoEst.DroppedData)
}

// New user trip
// Trip start event does not include geo
// Trip completion event includes start geo, populate Start Estimate with this
func Test_StartLocationNotIncludedInFirstEvent(t *testing.T) {
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

	tripStartPayload := segment1
	tripStartPayload.Data.Start.Location = nil
	tripStartPayload.Data.Completed = false
	if err := consumer.ProcessSegmentEvent(ctx, tripStartPayload); err != nil {
		t.Fatal(err)
	}

	trp, _ := models.Trips(models.TripWhere.ID.EQ(segment1.Data.ID)).One(ctx, pdb.DBS().Reader)
	assert.False(t, trp.StartPosition.Valid)
	assert.Equal(t, trp.StartTime, segment1.Data.Start.Time)

	segment1.Data.Completed = true
	if err := consumer.ProcessSegmentEvent(ctx, segment1); err != nil {
		t.Fatal(err)
	}
	err := trp.Reload(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)
	assert.Equal(t, trp.StartPositionEstimate.X, segment1.Data.Start.Location.Longitude)
	assert.Equal(t, trp.StartPositionEstimate.Y, segment1.Data.Start.Location.Latitude)
	assert.Equal(t, trp.EndPosition.X, segment1.Data.End.Location.Longitude)
	assert.Equal(t, trp.EndPosition.Y, segment1.Data.End.Location.Latitude)
	assert.True(t, trp.EndTime.Time.Equal(segment1.Data.End.Time))
	assert.True(t, trp.DroppedData)
}

// New user trip (has prev trip)
// Trip start event does not include geo
// Trip completion event includes start geo
// Based on trip completion, estimate a start loc based on prev trip and new trip start
func Test_EstimateStartOnCompletion(t *testing.T) {
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

	trp1 := models.Trip{
		ID:             ksuid.New().String(),
		VehicleTokenID: createDevice.Data.NFT.TokenID,
		StartTime:      segment1.Data.Start.Time,
		EndTime:        null.TimeFrom(segment1.Data.End.Time),
		StartPosition:  nullLocationToDB(segment1.Data.Start.Location),
		EndPosition:    nullLocationToDB(segment1.Data.End.Location),
	}

	if err := trp1.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	completed := segment2
	segment2.Data.Start.Location = nil
	if err := consumer.ProcessSegmentEvent(ctx, segment2); err != nil {
		t.Fatal(err)
	}

	estTrp, _ := models.Trips(models.TripWhere.ID.EQ(segment2.Data.ID)).One(ctx, pdb.DBS().Reader)
	assert.False(t, estTrp.StartPosition.Valid)
	assert.Equal(t, estTrp.StartTime, segment2.Data.Start.Time)
	assert.True(t, estTrp.DroppedData)

	completed.Data.Completed = true
	if err := consumer.ProcessSegmentEvent(ctx, completed); err != nil {
		t.Fatal(err)
	}
	err := estTrp.Reload(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)
	assert.False(t, estTrp.StartPosition.Valid)
	assert.True(t, estTrp.StartPositionEstimate.Valid)
	assert.Equal(t, estTrp.StartPositionEstimate.X, trp1.EndPosition.X)
	assert.Equal(t, estTrp.StartPositionEstimate.Y, trp1.EndPosition.Y)
	assert.Equal(t, estTrp.EndPosition.X, completed.Data.End.Location.Longitude)
	assert.Equal(t, estTrp.EndPosition.Y, completed.Data.End.Location.Latitude)
	assert.True(t, estTrp.EndTime.Time.Equal(segment2.Data.End.Time))

}
