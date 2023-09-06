package helpers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

func GetUserID(c *fiber.Ctx) string {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)
	return userID
}

type SegmentEvent struct {
	Start    PointTime `json:"start"`
	End      PointTime `json:"end"`
	DeviceID string    `json:"deviceID"`
}

type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type PointTime struct {
	Point Point     `json:"point"`
	Time  time.Time `json:"time"`
}

type SegmentSummary struct {
	TripID   string    `json:"tripID"`
	DeviceID string    `json:"deviceID"`
	BundlrID string    `json:"bundlrID"`
	Start    PointTime `json:"start"`
	End      PointTime `json:"end"`
}
