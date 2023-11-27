package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/kafka"
	"github.com/DIMO-Network/shared/middleware/privilegetoken"
	_ "github.com/DIMO-Network/trips-api/docs"
	"github.com/DIMO-Network/trips-api/internal/api"
	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/ethereum/go-ethereum/common"

	jwtware "github.com/gofiber/contrib/jwt"

	"github.com/DIMO-Network/trips-api/internal/database"
	"github.com/DIMO-Network/trips-api/internal/services/bundlr"
	"github.com/DIMO-Network/trips-api/internal/services/consumer"
	es_store "github.com/DIMO-Network/trips-api/internal/services/es"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/swagger"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// const userIDContextKey = "userID"

// @title			DIMO Segment API
// @version		1.0
// @description	segments
// @BasePath		/v1
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
func main() {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "trips-api").Logger()

	settings, err := shared.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed loading settings.")
	}

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
		database.MigrateDatabase(logger, &settings, command, "trips_api")
	default:
		esStore, err := es_store.New(&settings)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to establish connection to elasticsearch.")
		}

		pgStore, err := pg_store.New(&settings)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to establish connection to postgres.")
		}

		bundlrClient, err := bundlr.New(&settings)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to initialize Bundlr client")
		}

		controller := consumer.New(esStore, bundlrClient, pgStore, &logger, settings.DataFetchEnabled, settings.WorkerCount, settings.BundlrEnabled)
		segmentChannel := make(chan *shared.CloudEvent[consumer.SegmentEvent])
		vehicleEventChannel := make(chan *shared.CloudEvent[consumer.UserDeviceMintEvent])
		var wg sync.WaitGroup

		if err := kafka.Consume(ctx, kafka.Config{
			Brokers: strings.Split(settings.KafkaBrokers, ","),
			Topic:   settings.TripEventTopic,
			Group:   "completed-segment",
		}, controller.ProcessSegmentEvent, &logger); err != nil {
			logger.Fatal().Err(err).Msg("Couldn't start completed segment consumer.")
		}

		if err := kafka.Consume(ctx, kafka.Config{
			Brokers: strings.Split(settings.KafkaBrokers, ","),
			Topic:   settings.EventTopic,
			Group:   "vehicle-event",
		}, controller.VehicleEvent, &logger); err != nil {
			logger.Fatal().Err(err).Msg("Couldn't start vehicle event consumer.")
		}

		logger.Info().Interface("settings", settings.PrivilegeJWKURL).Msg("Settings")

		app := fiber.New()
		v1 := app.Group("/v1")
		v1.Get("/swagger/*", swagger.HandlerDefault)

		go serveMonitoring(settings.MonPort, &logger) //nolint

		privilegeJWT := jwtware.New(jwtware.Config{
			JWKSetURLs: []string{settings.PrivilegeJWKURL},
		})

		privilege := privilegetoken.New(privilegetoken.Config{
			Log: &logger,
		})
		vehicleAddr := common.HexToAddress(settings.VehicleNFTAddr)

		handler := api.NewHandler(pgStore)
		v1.Get("/vehicle/:tokenID/trips", privilegeJWT, privilege.OneOf(vehicleAddr, []int64{4}), handler.GetVehicleTrips)

		go func() {
			logger.Info().Msgf("Starting API server on port %s.", settings.Port)
			if err := app.Listen(fmt.Sprintf(":%s", settings.Port)); err != nil {
				logger.Fatal().Err(err)
			}
		}()

		c := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent with length of 1
		signal.Notify(c, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel
		<-c                                             // This blocks the main thread until an interrupt is received
		logger.Info().Msg("Gracefully shutting down and running cleanup tasks...")
		close(segmentChannel)
		close(vehicleEventChannel)
		wg.Wait()
		_ = app.Shutdown()
	}
}

func serveMonitoring(port string, logger *zerolog.Logger) (*fiber.App, error) {
	logger.Info().Str("port", port).Msg("Starting monitoring web server.")

	monApp := fiber.New(fiber.Config{DisableStartupMessage: true})

	monApp.Get("/", func(c *fiber.Ctx) error { return nil })
	monApp.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	go func() {
		if err := monApp.Listen(":" + port); err != nil {
			logger.Fatal().Err(err).Str("port", port).Msg("Failed to start monitoring web server.")
		}
	}()

	return monApp, nil
}
