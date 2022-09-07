package main

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/trips-api/internal/config"
	_ "github.com/lib/pq"
	"github.com/lovoo/goka"
	"github.com/rs/zerolog"
)

var (
	group     goka.Group  = "trips-api"
	tripTopic goka.Stream = "topic.device.trip.event"
)

var tripStatusCodec = &shared.JSONCodec[TripStatus]{}

type TripStatus struct {
	DeviceID string
	Start    time.Time
	End      time.Time
}

func (p *TripEventProcessor) UpdateCompletedTrip(trp TripStatus) error {
	query := `UPDATE fulltrips SET tripend = $1 WHERE deviceid = $2 AND tripid = $3`
	_, err := p.db.Exec(query, trp.End, trp.DeviceID, p.uniqueTripID(trp.DeviceID, trp.Start))
	if err != nil {
		return err
	}
	return nil
}

func (p *TripEventProcessor) BeginNewTrip(trp TripStatus) error {

	query := `INSERT INTO fulltrips (deviceid, tripstart, tripid) VALUES ($1, $2, $3)`
	_, err := p.db.Exec(query, trp.DeviceID, trp.Start, p.uniqueTripID(trp.DeviceID, trp.Start))
	if err != nil {
		return err
	}
	return nil
}

func (p *TripEventProcessor) listenForTrips(ctx goka.Context, msg any) {
	ongoingTrip := msg.(*TripStatus)

	if !ongoingTrip.End.IsZero() {
		err := p.UpdateCompletedTrip(*ongoingTrip)
		if err != nil {
			p.logger.Err(err)
		}
		return
	}

	err := p.BeginNewTrip(*ongoingTrip)
	if err != nil {
		p.logger.Err(err)
	}

}

type TripEventProcessor struct {
	logger *zerolog.Logger
	db     *sql.DB
}

func (p *TripEventProcessor) uniqueTripID(userID string, startTime time.Time) uint32 {
	start := startTime.Format("2006-01-02 15:04:05")
	h := fnv.New32()
	h.Write([]byte(userID + start))
	return h.Sum32()
}

// process messages until ctrl-c is pressed
func (p *TripEventProcessor) runProcessor(brokers []string) {
	// Define a new processor group. The group defines all inputs, outputs, and
	// serialization formats. The group-table topic is "example-group-table".
	g := goka.DefineGroup(group,
		goka.Input(tripTopic, tripStatusCodec, p.listenForTrips),
	)

	proc, err := goka.NewProcessor(brokers, g)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Failed to create processor.")
	}
	ctx, cancel := context.WithCancel(context.Background())

	// Closure of this channel tells us that proc.Run has exited.
	done := make(chan bool)

	go func() {
		defer close(done)
		if err = proc.Run(ctx); err != nil {
			p.logger.Fatal().Err(err).Msg("Processor terminated with an error.")
		} else {
			p.logger.Info().Msg("Processor shut down cleanly.")
		}
	}()

	wait := make(chan os.Signal, 1)
	signal.Notify(wait, syscall.SIGINT, syscall.SIGTERM)
	<-wait
	cancel() // Stop Goka processor proc.
	<-done
}

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "trips-api").Logger()

	settings, err := shared.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed loading settings.")
	}
	logger.Info().Interface("settings", settings).Msg("Settings loaded.")

	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		settings.DBHost,
		settings.DBPort,
		settings.DBUser,
		settings.DBPassword,
		settings.DBName,
	)
	fmt.Println(psqlInfo)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to creating database handle.")
	}

	err = db.Ping()
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't reach database.")
	}

	tep := &TripEventProcessor{logger: &logger, db: db}
	brokers := strings.Split(settings.KafkaBrokers, ",")
	tep.runProcessor(brokers) // press ctrl-c to stop
}
