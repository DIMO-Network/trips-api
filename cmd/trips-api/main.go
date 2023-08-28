package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb_devices "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	_ "github.com/DIMO-Network/trips-api/docs"
	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/DIMO-Network/trips-api/internal/handlers/pg_handler"
	"github.com/DIMO-Network/trips-api/internal/services/bundlr"
	"github.com/DIMO-Network/trips-api/internal/services/consumer"
	es_store "github.com/DIMO-Network/trips-api/internal/services/es"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/gofiber/swagger"
	"github.com/golang-jwt/jwt"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const userIDContextKey = "userID"

// var tripStatusCodec = &shared.JSONCodec[kafka.TripStatus]{}

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
		conn, err := grpc.Dial(settings.DevicesAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to establish devices grpc connection.")
		}
		defer conn.Close()

		deviceClient := pb_devices.NewUserDeviceServiceClient(conn)

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
			logger.Fatal().Err(err).Msg("Failed to initialize Bunldr client")
		}

		consumer, err := consumer.New(esStore, bundlrClient, pgStore, deviceClient, &settings, &logger)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create consumer.")
		}

		consumer.Start(ctx)

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

		handler := pg_handler.New(pgStore, deviceClient)

		app := fiber.New()
		app.Get("/swagger/*", swagger.HandlerDefault)
		app.Use(cors.New(cors.Config{AllowOrigins: "*"}))

		deviceGroup := app.Group("/devices/:id", jwtAuth)
		deviceGroup.Get("/segments", handler.Segments)

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

		c := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent with length of 1
		signal.Notify(c, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel
		<-c                                             // This blocks the main thread until an interrupt is received
		logger.Info().Msg("Gracefully shutting down and running cleanup tasks...")
		_ = app.Shutdown()
	}
}
