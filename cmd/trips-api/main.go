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
	_ "github.com/DIMO-Network/trips-api/docs"
	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/DIMO-Network/trips-api/internal/database"
	"github.com/DIMO-Network/trips-api/internal/kafka"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/gofiber/swagger"
	"github.com/golang-jwt/jwt/v4"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

const userIDContextKey = "userID"

// var tripStatusCodec = &shared.JSONCodec[kafka.TripStatus]{}

// @title                      DIMO Segment API
// @version                    1.0
// @description segments

// @BasePath /

// @name Authorization
func main() {
	ctx, cancel := context.WithCancel(context.Background())
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

		tripQueryController := database.NewDatabaseConnection(settings, &logger)
		tep := kafka.NewTripEventProcessor(&logger, tripQueryController.Pg)
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
					fmt.Println(claims)
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
		app.Get("/swagger/*", swagger.HandlerDefault)
		app.Use(cors.New(cors.Config{AllowOrigins: "*"}))
		// app.Get("/ongoing/all", tripQueryController.AllOngoingTrips)
		// app.Get("/devices/all", tripQueryController.AllUsers)

		deviceGroup := app.Group("/devices/:id", jwtAuth)
		// deviceGroup.Get("/ongoing", tripQueryController.DeviceTripOngoing)
		deviceGroup.Get("/alltrips", tripQueryController.AllDeviceTrips)

		app.Get("/health", func(c *fiber.Ctx) error {
			return c.JSON(map[string]interface{}{
				"data": "Server is up and running",
			})
		})

		go func() {
			logger.Info().Msgf("Starting API server on port %s.", settings.Port)
			if err := app.Listen(fmt.Sprintf(":%s", settings.Port)); err != nil {
				logger.Fatal().Err(err)
			}
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
