package database

import (
	"context"
	"database/sql"
	"fmt"

	pb "github.com/DIMO-Network/shared/api/devices"
	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PostgresController struct {
	Db                 *sql.DB
	devicesAPIGRPCAddr string
}

func NewDatabaseConnection(settings config.Settings, logger zerolog.Logger) PostgresController {
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

	return PostgresController{Db: db, devicesAPIGRPCAddr: settings.JwtKeySetURL}
}

func (p *PostgresController) AllOngoingTrips(c *fiber.Ctx) error {

	ongoingTrips := make([]string, 0)

	resp, err := models.Fulltrips(models.FulltripWhere.TripEnd.IsNull()).All(c.Context(), p.Db)
	if err != nil {
		return err
	}

	for _, trp := range resp {
		ongoingTrips = append(ongoingTrips, trp.DeviceID.String)
	}

	return c.JSON(ongoingTrips)
}

func (p *PostgresController) DeviceTripOngoing(c *fiber.Ctx) error {
	deviceID := c.Params("id")

	mods := []qm.QueryMod{
		models.FulltripWhere.DeviceID.EQ(null.StringFrom(deviceID)),
		models.FulltripWhere.TripEnd.IsNull(),
	}

	resp, err := models.Fulltrips(mods...).All(c.Context(), p.Db)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

func (p *PostgresController) AllDeviceTrips(c *fiber.Ctx) error {
	userID := getUserID(c)
	deviceID := c.Params("id")

	exists, err := p.UserDeviceBelongsToUserID(c.Context(), userID, deviceID)
	if err != nil {
		return err
	}
	if !exists {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	mods := []qm.QueryMod{
		models.FulltripWhere.DeviceID.EQ(null.StringFrom(deviceID)),
		models.FulltripWhere.TripStart.IsNotNull(),
		models.FulltripWhere.TripEnd.IsNotNull(),
		qm.OrderBy(models.FulltripColumns.TripEnd + " DESC"),
	}

	resp, err := models.Fulltrips(mods...).All(c.Context(), p.Db)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

func (das *PostgresController) UserDeviceBelongsToUserID(ctx context.Context, userID, userDeviceID string) (bool, error) {
	device, err := das.GetUserDevice(ctx, userDeviceID)
	if err != nil {
		return false, err
	}
	return device.UserId == userID, nil
}

func (das *PostgresController) GetUserDevice(ctx context.Context, userDeviceID string) (*pb.UserDevice, error) {
	conn, err := grpc.Dial(das.devicesAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	deviceClient := pb.NewUserDeviceServiceClient(conn)

	userDevice, err := deviceClient.GetUserDevice(ctx, &pb.GetUserDeviceRequest{
		Id: userDeviceID,
	})
	if err != nil {
		return nil, err
	}

	return userDevice, nil
}

func getUserID(c *fiber.Ctx) string {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)
	return userID
}
