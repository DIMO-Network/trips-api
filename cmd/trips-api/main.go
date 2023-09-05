package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/kafka"
	_ "github.com/DIMO-Network/trips-api/docs"
	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/golang-jwt/jwt"

	"github.com/DIMO-Network/trips-api/internal/database"
	"github.com/DIMO-Network/trips-api/internal/handlers/pg_handler"
	"github.com/DIMO-Network/trips-api/internal/services/bundlr"
	"github.com/DIMO-Network/trips-api/internal/services/consumer"
	es_store "github.com/DIMO-Network/trips-api/internal/services/es"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/gofiber/swagger"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

const userIDContextKey = "userID"

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

		controller := consumer.New(esStore, bundlrClient, pgStore, &logger)

		// start completed segment consumer
		consumer.Start(ctx, kafka.Config{
			Brokers: strings.Split(settings.KafkaBrokers, ","),
			Topic:   settings.TripEventTopic,
			Group:   "completed-segment",
		}, controller.CompletedSegment, &logger)

		// start vehicle event consumer
		consumer.Start(ctx, kafka.Config{
			Brokers: strings.Split(settings.KafkaBrokers, ","),
			Topic:   settings.EventTopic,
			Group:   "vehicle-event",
		}, controller.VehicleEvent, &logger)

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

		handler := pg_handler.New(pgStore, bundlrClient)

		app := fiber.New()
		app.Use(cors.New(cors.Config{AllowOrigins: "*"}))
		app.Get("/swagger/*", swagger.HandlerDefault)
		app.Get("/health", func(c *fiber.Ctx) error {
			return c.JSON(map[string]interface{}{
				"data": "Server is up and running",
			})
		})

		deviceGroup := app.Group("/devices/:id", jwtAuth)
		deviceGroup.Get("/segments", handler.Segments)

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
		_ = app.Shutdown()
	}
}
