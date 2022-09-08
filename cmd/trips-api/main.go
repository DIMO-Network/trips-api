package main

import (
	"context"
	"database/sql"
	"hash/fnv"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/DIMO-Network/trips-api/internal/database"
	"github.com/DIMO-Network/trips-api/internal/kafka"
	"github.com/DIMO-Network/trips-api/internal/services"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
	_ "github.com/lib/pq"
	"github.com/lovoo/goka"
	"github.com/rs/zerolog"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

const userIDContextKey = "userID"

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
	userID := c.Locals(userIDContextKey).(string)
	p.logger.Info().Msg("User is " + userID)

	deviceID := c.Params("deviceID")

	// Ask devices-api whether userID has access to deviceID

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
		services.MigrateDatabase(logger, &settings, command, "trips_api")
	default:

		pgController := database.NewDatabaseConnection(settings, logger)
		tep := kafka.NewTripEventProcessor(&logger, pgController.Db)
		brokers := strings.Split(settings.KafkaBrokers, ",")
		go tep.RunProcessor(ctx, brokers) // press ctrl-c to stop

		keyRefreshInterval := time.Hour
		keyRefreshUnknownKID := true
		jwtAuth := jwtware.New(
			jwtware.Config{
				KeySetURL:            settings.JWTKeySetURL,
				KeyRefreshInterval:   &keyRefreshInterval,
				KeyRefreshUnknownKID: &keyRefreshUnknownKID,
				SuccessHandler: func(c *fiber.Ctx) error {
					token := c.Locals("user").(*jwt.Token)
					claims := token.Claims.(jwt.MapClaims)
					c.Locals(userIDContextKey, claims["sub"].(string))
					return c.Next()
				},
				ErrorHandler: func(c *fiber.Ctx, err error) error {
					return c.Status(fiber.StatusUnauthorized).JSON(
						map[string]any{
							"code":    401,
							"message": "Invalid or expired JWT.",
						},
					)
				},
			},
		)

		logger.Info().Interface("settings", settings).Msg("Settings")

		app := fiber.New()

		app.Use(cors.New(cors.Config{AllowOrigins: "*"}))

		app.Get("/ongoing/all", pgController.AllOngoingTrips)

		app.Group("/devices/:id", jwtAuth)
		app.Get("/ongoing", jwtAuth, pgController.DeviceTripOngoing)
		app.Get("/alltrips", jwtAuth, pgController.AllDeviceTrips)
		app.Get("/", func(c *fiber.Ctx) error {
			return c.SendString("Hello, World!")
		})
		go func() {
			app.Listen(":8000")
		}()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		logger.Info().Str("signal", sig.String()).Msg("Got signal, terminating.")
		cancel()
		<-tep.Done
		err = app.Shutdown()
		if err != nil {
			logger.Err(err).Msg("HTTP server did not shut down gracefully.")
		}
	}
}
