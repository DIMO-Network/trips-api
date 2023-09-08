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
	_ "github.com/DIMO-Network/trips-api/docs"
	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/DIMO-Network/trips-api/internal/services/bundlr"
	"github.com/DIMO-Network/trips-api/internal/services/consumer"
	es_store "github.com/DIMO-Network/trips-api/internal/services/es"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// @title                      DIMO Segment API
// @version                    1.0
// @description segments

// @BasePath /

// @name Authorization
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
		MigrateDatabase(logger, &settings, command, "trips_api")
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
			logger.Fatal().Err(err).Msg("Failed to start Bunldr uploader")
		}

		controller := consumer.New(esStore, bundlrClient, pgStore, &logger)
		segmentChannel := make(chan shared.CloudEvent[consumer.SegmentEvent])
		vehicleEventChannel := make(chan shared.CloudEvent[consumer.UserDeviceMintEvent])
		var wg sync.WaitGroup

		// start completed segment consumer
		consumer.Start(ctx, kafka.Config{
			Brokers: strings.Split(settings.KafkaBrokers, ","),
			Topic:   settings.TripEventTopic,
			Group:   "completed-segment",
		}, controller.CompletedSegment, segmentChannel, &wg, &logger)

		// start vehicle event consumer
		consumer.Start(ctx, kafka.Config{
			Brokers: strings.Split(settings.KafkaBrokers, ","),
			Topic:   settings.EventTopic,
			Group:   "vehicle-event",
		}, controller.VehicleEvent, vehicleEventChannel, &wg, &logger)

		app := fiber.New()
		app.Get("/health", func(c *fiber.Ctx) error {
			return c.JSON(map[string]interface{}{
				"data": "Server is up and running",
			})
		})

		go serveMonitoring(settings.MonPort, &logger) //nolint

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
