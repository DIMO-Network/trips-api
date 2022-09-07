package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/lovoo/goka"
	"github.com/rs/zerolog"
)

var (
	brokers                = []string{"localhost:9092"}
	group      goka.Group  = "trips-api"
	tripStatus goka.Stream = "topic.device.trip.event"
)

var tripStatusCodec = &shared.JSONCodec[TripStatus]{}

type TripStatus struct {
	DeviceID string
	Start    time.Time
	End      time.Time
}

func (p *TripEventProcessor) StoreTrip(trp TripStatus) error {

	if !trp.End.IsZero() {
		query := `INSERT INTO fulltrips (deviceid, tripstart, tripend, tripid) VALUES ($1, $2, $3, $4) `
		_, err := p.db.Exec(query, trp.DeviceID, trp.Start, trp.End, p.TripCount)
		if err != nil {
			return err
		}
		p.TripCount++
	}
	return nil
}

func (p *TripEventProcessor) listenForCompletedTrips(ctx goka.Context, msg any) {

	completedTrip := msg.(*TripStatus)
	err := p.StoreTrip(*completedTrip)
	if err != nil {
		p.logger.Err(err)
	}

}

type TripEventProcessor struct {
	logger    *zerolog.Logger
	db        *sql.DB
	TripCount int
}

// process messages until ctrl-c is pressed
func (p *TripEventProcessor) runProcessor() {
	// Define a new processor group. The group defines all inputs, outputs, and
	// serialization formats. The group-table topic is "example-group-table".
	g := goka.DefineGroup(group,
		goka.Input(tripStatus, tripStatusCodec, p.listenForCompletedTrips),
	)

	proc, err := goka.NewProcessor(brokers, g)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Failed to create processor.")
	}
	ctx, cancel := context.WithCancel(context.Background())
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
		logger.Fatal().Err(err).Msg("could not load settings")
	}
	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		settings.DBHost,
		settings.DBPort,
		settings.DBUser,
		settings.DBPassword,
		settings.DBName,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	tp := &TripEventProcessor{logger: &logger, db: db}
	tp.runProcessor() // press ctrl-c to stop
}
