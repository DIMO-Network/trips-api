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

// TODO (AE): this is ugly
func Test_DontEstimateTripStart(t *testing.T) {
	ctx := context.Background()

	var createDevice shared.CloudEvent[UserDeviceMintEvent]
	var beginSegment shared.CloudEvent[SegmentEvent]
	var completeSegment shared.CloudEvent[SegmentEvent]
	var newTrpEstimatestart shared.CloudEvent[SegmentEvent]
	var newTrpNoStartEstimate shared.CloudEvent[SegmentEvent]

	createDevice.Type = UserDeviceMintEventType
	createDevice.Data.Device.ID = ksuid.New().String()
	createDevice.Data.NFT.TokenID = 1

	beginSegment.Data.DeviceID = createDevice.Data.Device.ID
	beginSegment.Data.ID = testExpectedStartAndEnd
	beginSegment.Data.Start.Time = time.Now().Add(-time.Hour * 3 * 24).UTC()
	beginSegment.Data.Start.Latitude = 40.744331740800455
	beginSegment.Data.Start.Longitude = -73.98043334522801

	completeSegment.Data.Completed = true
	completeSegment.Data.ID = testExpectedStartAndEnd
	completeSegment.Data.DeviceID = createDevice.Data.Device.ID
	completeSegment.Data.End.Time = time.Now().Add(-time.Minute * 30).UTC()
	completeSegment.Data.End.Latitude = 33.84805567103969
	completeSegment.Data.End.Longitude = -118.39318923141917

	newTrpEstimatestart.Data.DeviceID = createDevice.Data.Device.ID
	newTrpEstimatestart.Data.ID = testEstimateStart
	newTrpEstimatestart.Data.Start.Time = time.Now().Add(-time.Minute * 3).UTC()
	newTrpEstimatestart.Data.Start.Latitude = 33.850422561365455
	newTrpEstimatestart.Data.Start.Longitude = -118.3962470088937

	newTrpNoStartEstimate.Data.DeviceID = createDevice.Data.Device.ID
	newTrpNoStartEstimate.Data.ID = testFailToEstimateStart
	newTrpNoStartEstimate.Data.Start.Time = time.Now().Add(-time.Minute * 5).UTC()
	newTrpNoStartEstimate.Data.Start.Latitude = 33.95737251631686
	newTrpNoStartEstimate.Data.Start.Longitude = -118.44861917146383

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

	if err := consumer.BeginSegment(ctx, beginSegment); err != nil {
		t.Fatal(err)
	}
	trp, _ := models.Trips().One(ctx, pdb.DBS().Reader)
	assert.Equal(t, trp.StartPosition.X, beginSegment.Data.Start.Longitude)
	assert.Equal(t, trp.StartPosition.Y, beginSegment.Data.Start.Latitude)
	assert.Equal(t, trp.StartTime, beginSegment.Data.Start.Time)

	if err := consumer.CompleteSegment(ctx, completeSegment); err != nil {
		t.Fatal(err)
	}
	err := trp.Reload(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)
	assert.Equal(t, trp.EndPosition.X, completeSegment.Data.End.Longitude)
	assert.Equal(t, trp.EndPosition.Y, completeSegment.Data.End.Latitude)
	assert.Equal(t, trp.EndTime.Time, completeSegment.Data.End.Time)

	if err := consumer.BeginSegment(ctx, newTrpEstimatestart); err != nil {
		t.Fatal(err)
	}
	estTrp, _ := models.Trips(models.TripWhere.ID.EQ(newTrpEstimatestart.Data.ID)).One(ctx, pdb.DBS().Reader)
	assert.Equal(t, estTrp.StartPositionEstimate.X, completeSegment.Data.End.Longitude)
	assert.Equal(t, estTrp.StartPositionEstimate.Y, completeSegment.Data.End.Latitude)

	if err := consumer.BeginSegment(ctx, newTrpNoStartEstimate); err != nil {
		t.Fatal(err)
	}

	newTripNoEst, _ := models.Trips(models.TripWhere.ID.EQ(newTrpNoStartEstimate.Data.ID)).One(ctx, pdb.DBS().Reader)
	assert.Equal(t, newTripNoEst.StartPositionEstimate, pgeo.NullPoint{})
	assert.Equal(t, newTripNoEst.StartPosition.X, newTrpNoStartEstimate.Data.Start.Longitude)
	assert.Equal(t, newTripNoEst.StartPosition.Y, newTrpNoStartEstimate.Data.Start.Latitude)

}
