package pg_handler

import (
	pb_devices "github.com/DIMO-Network/devices-api/pkg/grpc"
	pg_store "github.com/DIMO-Network/trips-api/internal/services/pg"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/ericlagergren/decimal"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type Handler struct {
	*pg_store.Store
	grpc pb_devices.UserDeviceServiceClient
}

func New(pgStore *pg_store.Store, grpcDevices pb_devices.UserDeviceServiceClient) (*Handler, error) {
	return &Handler{pgStore, grpcDevices}, nil
}

// Segments godoc
// @Description details for all segments associated with device
// @Tags        user-segments
// @Produce     json
// @Security    BearerAuth
// @Router      /devices/:id/segments [get]
func (h *Handler) Segments(c *fiber.Ctx) error {
	deviceID := c.Params("id")
	userDevice, err := h.grpc.GetUserDevice(c.Context(), &pb_devices.GetUserDeviceRequest{
		Id: deviceID,
	})
	if err != nil {
		return err
	}

	allSegments, err := models.Trips(
		models.TripWhere.VehicleTokenID.EQ(types.NewDecimal(decimal.New(int64(*userDevice.TokenId), 0))),
		qm.Select(
			models.TripColumns.VehicleTokenID,
			models.TripColumns.Start,
			models.TripColumns.StartHex,
			models.TripColumns.StartPosition,
			models.TripColumns.End,
			models.TripColumns.EndHex,
			models.TripColumns.EndPosition),
	).All(c.Context(), h.Db)
	if err != nil {
		return err
	}

	return c.JSON(allSegments)
}

func getUserID(c *fiber.Ctx) string {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)
	return userID
}
