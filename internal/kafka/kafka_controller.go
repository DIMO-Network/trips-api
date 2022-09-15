package kafka

import (
	"context"
	"database/sql"
	"hash/fnv"
	"strconv"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/lovoo/goka"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var (
	group     goka.Group  = "trips-api"
	tripTopic goka.Stream = "topic.device.trip.event"
)

type TripEvent struct {
	DeviceID string    `json:"deviceId"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
}

type TripEventProcessor struct {
	logger *zerolog.Logger
	db     *sql.DB
	Done   chan bool
}

func NewTripEventProcessor(logger *zerolog.Logger, db *sql.DB) *TripEventProcessor {
	return &TripEventProcessor{logger: logger, db: db, Done: make(chan bool)}
}

var tripStatusCodec = &shared.JSONCodec[shared.CloudEvent[TripEvent]]{}

func (p *TripEventProcessor) uniqueTripID(userID string, startTime time.Time) string {
	start := startTime.Format("2006-01-02 15:04:05")
	h := fnv.New32()
	h.Write([]byte(userID + start))
	tid := strconv.Itoa(int(h.Sum32()))
	return tid
}

func (p *TripEventProcessor) updateCompletedTrip(ctx context.Context, trp TripEvent) error {

	tm, err := models.FindFulltrip(ctx, p.db, p.uniqueTripID(trp.DeviceID, trp.Start))
	if err != nil {
		return err
	}

	tm.TripEnd = null.TimeFrom(trp.End)
	_, err = tm.Update(ctx, p.db, boil.Infer())

	return err
}

func (p *TripEventProcessor) beginNewTrip(ctx context.Context, trp TripEvent) error {
	newTrip := models.Fulltrip{
		TripID:    p.uniqueTripID(trp.DeviceID, trp.Start),
		DeviceID:  null.StringFrom(trp.DeviceID),
		TripStart: null.TimeFrom(trp.Start),
	}
	return newTrip.Insert(ctx, p.db, boil.Infer())
}

func (p *TripEventProcessor) listenForTrips(ctx goka.Context, msg any) {
	event := msg.(*shared.CloudEvent[TripEvent])
	ongoingTrip := event.Data

	p.logger.Info().Str("userDeviceId", ctx.Key()).Interface("event", event.Data).Msg("Received trip event.")

	if !ongoingTrip.End.IsZero() {
		err := p.updateCompletedTrip(ctx.Context(), ongoingTrip)
		if err != nil {
			p.logger.Err(err).Str("userDeviceId", ctx.Key()).Msg("Failed updating trip.")
		}
		return
	}

	err := p.beginNewTrip(ctx.Context(), ongoingTrip)
	if err != nil {
		p.logger.Err(err).Str("userDeviceId", ctx.Key()).Msg("Failed creating trip.")
	}

}

// process messages until ctrl-c is pressed
func (p *TripEventProcessor) RunProcessor(ctx context.Context, brokers []string) {
	// Define a new processor group. The group defines all inputs, outputs, and
	// serialization formats. The group-table topic is "example-group-table".
	g := goka.DefineGroup(group,
		goka.Input(tripTopic, tripStatusCodec, p.listenForTrips),
	)

	proc, err := goka.NewProcessor(brokers, g)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Failed to create processor.")
	}

	go func() {
		defer close(p.Done)
		if err := proc.Run(ctx); err != nil {
			p.logger.Fatal().Err(err).Msg("Processor terminated with an error.")
		} else {
			p.logger.Info().Msg("Processor shut down cleanly.")
		}
	}()
}
