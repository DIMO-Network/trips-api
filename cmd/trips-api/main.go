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
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

type tripOngoing struct {
	Resp bool `json:"tripOngoing"`
}

type deviceTrip struct {
	Start time.Time `json:"tripStart"`
	End   time.Time `json:"tripEnd"`
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
	done   chan bool
}

func (p *TripEventProcessor) uniqueTripID(userID string, startTime time.Time) uint32 {
	start := startTime.Format("2006-01-02 15:04:05")
	h := fnv.New32()
	h.Write([]byte(userID + start))
	return h.Sum32()
}

func (p *TripEventProcessor) allOngoingTrips(c *fiber.Ctx) error {

	ongoingTrips := make([]string, 0)
	sqlQuery := `SELECT DISTINCT deviceid FROM fulltrips WHERE tripend IS NULL;`
	rows, err := p.db.Query(sqlQuery)
	if err != nil {
		return err
	}

	for rows.Next() {
		var device string
		err := rows.Scan(&device)
		if err != nil {
			return err
		}
		ongoingTrips = append(ongoingTrips, device)
	}

	return c.JSON(ongoingTrips)
}

func (p *TripEventProcessor) deviceTripOngoing(c *fiber.Ctx) error {

	response := tripOngoing{}

	deviceID := c.Params("deviceID")
	sqlQuery := `SELECT 
					CASE WHEN tripend IS NULL THEN TRUE
					ELSE FALSE 
					END AS ongoingtrip
					FROM fulltrips
					WHERE deviceid = $1
					ORDER BY tripend DESC
					LIMIT 1;`
	rows, err := p.db.Query(sqlQuery, deviceID)
	if err != nil {
		return err
	}

	for rows.Next() {
		err := rows.Scan(&response.Resp)
		if err != nil {
			return err
		}
	}

	return c.JSON(response)
}

func (p *TripEventProcessor) allDeviceTrips(c *fiber.Ctx) error {

	response := []deviceTrip{}
	deviceID := c.Params("deviceID")
	sqlQuery := `SELECT 
					tripstart, tripend
					FROM fulltrips
					WHERE deviceid = $1
					AND tripend IS NOT NULL
					ORDER BY tripend DESC;`
	rows, err := p.db.Query(sqlQuery, deviceID)
	if err != nil {
		return err
	}

	for rows.Next() {
		trp := deviceTrip{}
		err := rows.Scan(&trp.Start, &trp.End)
		if err != nil {
			return err
		}
		response = append(response, trp)
	}

	return c.JSON(response)
}

// process messages until ctrl-c is pressed
func (p *TripEventProcessor) runProcessor(ctx context.Context, brokers []string) {
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
		defer close(p.done)
		if err := proc.Run(ctx); err != nil {
			p.logger.Fatal().Err(err).Msg("Processor terminated with an error.")
		} else {
			p.logger.Info().Msg("Processor shut down cleanly.")
		}
	}()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

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
	defer func() {
		if err := db.Close(); err != nil {
			logger.Err(err).Msg("Database handle did not shut down gracefully.")
		}
	}()

	tep := &TripEventProcessor{logger: &logger, db: db, done: make(chan bool)}
	brokers := strings.Split(settings.KafkaBrokers, ",")
	go tep.runProcessor(ctx, brokers) // press ctrl-c to stop

	app := fiber.New()
	app.Use(cors.New(cors.Config{AllowOrigins: "*"}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	app.Get("/ongoing/all", tep.allOngoingTrips)
	app.Get("/ongoing/:deviceID", tep.deviceTripOngoing)
	app.Get("/alltrips/:deviceID", tep.allDeviceTrips)

	go func() {
		app.Listen(":8000")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	logger.Info().Str("signal", sig.String()).Msg("Got signal, terminating.")
	cancel()
	<-tep.done
	err = app.Shutdown()
	if err != nil {
		logger.Err(err).Msg("HTTP server did not shut down gracefully.")
	}
}
