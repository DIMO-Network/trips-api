package main

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/lib/pq"
	"github.com/lovoo/goka"
	"github.com/pressly/goose"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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

func (p *TripEventProcessor) UpdateCompletedTrip(ctx context.Context, trp TripStatus) error {
	tid := strconv.Itoa(int(p.uniqueTripID(trp.DeviceID, trp.Start)))

	tm, err := models.FindFulltrip(ctx, p.db, tid)
	if err != nil {
		return err
	}

	tm.TripEnd = null.TimeFrom(trp.End)
	_, err = tm.Update(ctx, p.db, boil.Infer())

	return err
}

func (p *TripEventProcessor) BeginNewTrip(ctx context.Context, trp TripStatus) error {
	tid := strconv.Itoa(int(p.uniqueTripID(trp.DeviceID, trp.Start)))

	newTrip := models.Fulltrip{
		TripID:    tid,
		DeviceID:  null.StringFrom(trp.DeviceID),
		TripStart: null.TimeFrom(trp.Start),
	}
	return newTrip.Insert(ctx, p.db, boil.Infer())
}

func (p *TripEventProcessor) listenForTrips(ctx goka.Context, msg any) {
	ongoingTrip := msg.(*TripStatus)

	if !ongoingTrip.End.IsZero() {
		err := p.UpdateCompletedTrip(ctx.Context(), *ongoingTrip)
		if err != nil {
			p.logger.Err(err)
		}
		return
	}

	err := p.BeginNewTrip(ctx.Context(), *ongoingTrip)
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

	resp, err := models.Fulltrips(models.FulltripWhere.TripEnd.IsNull()).All(c.Context(), p.db)
	if err != nil {
		return err
	}

	for _, trp := range resp {
		ongoingTrips = append(ongoingTrips, trp.DeviceID.String)
	}

	return c.JSON(ongoingTrips)
}

func (p *TripEventProcessor) deviceTripOngoing(c *fiber.Ctx) error {
	deviceID := c.Params("deviceID")

	mods := []qm.QueryMod{
		models.FulltripWhere.DeviceID.EQ(null.StringFrom(deviceID)),
		models.FulltripWhere.TripEnd.IsNull(),
	}

	resp, err := models.Fulltrips(mods...).All(c.Context(), p.db)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

func (p *TripEventProcessor) allDeviceTrips(c *fiber.Ctx) error {
	deviceID := c.Params("deviceID")

	mods := []qm.QueryMod{
		models.FulltripWhere.DeviceID.EQ(null.StringFrom(deviceID)),
		models.FulltripWhere.TripStart.IsNotNull(),
		models.FulltripWhere.TripEnd.IsNotNull(),
		qm.OrderBy(models.FulltripColumns.TripEnd + " DESC"),
	}

	resp, err := models.Fulltrips(mods...).All(c.Context(), p.db)
	if err != nil {
		return err
	}

	return c.JSON(resp)
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

	arg := ""
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}
	switch arg {
	case "migrate":
		command := "up"
		if len(os.Args) > 2 {
			command = os.Args[2]
			if command == "down-to" || command == "up-to" {
				command = command + " " + os.Args[3]
			}
		}
		migrateDatabase(logger, &settings, command, "trips_api")
	default:
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
}

func migrateDatabase(logger zerolog.Logger, settings *config.Settings, command, schemaName string) {
	var db *sql.DB

	// setup database
	db, err := sql.Open("postgres", settings.GetWriterDSN(true))
	defer func() {
		if err := db.Close(); err != nil {
			logger.Fatal().Msgf("goose: failed to close DB: %v\n", err)
		}
	}()
	if err != nil {
		logger.Fatal().Msgf("failed to open db connection: %v\n", err)
	}
	if err = db.Ping(); err != nil {
		logger.Fatal().Msgf("failed to ping db: %v\n", err)
	}

	// set default
	if command == "" {
		command = "up"
	}
	// must create schema so that can set migrations table to that schema
	_, err = db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;", schemaName))
	if err != nil {
		logger.Fatal().Err(err).Msgf("could not create schema, %s", schemaName)
	}
	goose.SetTableName(fmt.Sprintf("%s.migrations", schemaName))
	if err := goose.Run(command, db, "migrations"); err != nil {
		logger.Fatal().Msgf("failed to apply go code migrations: %v\n", err)
	}
}
