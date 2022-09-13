package database

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"time"

	pb "github.com/DIMO-Network/shared/api/devices"
	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	geojson "github.com/paulmach/go.geojson"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TripDataQueryService connected to postgres db containing trip information and validates user
type TripDataQueryService struct {
	Pg                 *sql.DB
	es                 *elasticsearch.Client
	elasticIndex       string
	log                *zerolog.Logger
	devicesAPIGRPCAddr string
}

func NewDatabaseConnection(settings config.Settings, logger *zerolog.Logger) TripDataQueryService {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		settings.DBHost,
		settings.DBPort,
		settings.DBUser,
		settings.DBPassword,
		settings.DBName,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create postgres database handle.")
	}

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{settings.ElasticSearchAnalyticsHost},
		Username:  settings.ElasticSearchAnalyticsUsername,
		Password:  settings.ElasticSearchAnalyticsPassword,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create elasticsearch handle.")
	}

	return TripDataQueryService{Pg: db, devicesAPIGRPCAddr: settings.DevicesAPIGRPCAddr, log: logger, es: esClient, elasticIndex: settings.ElasticSearchIndex}
}

func (p *TripDataQueryService) AllOngoingTrips(c *fiber.Ctx) error {

	ongoingTrips := make([]string, 0)

	resp, err := models.Fulltrips(models.FulltripWhere.TripEnd.IsNull()).All(c.Context(), p.Pg)
	if err != nil {
		return err
	}

	for _, trp := range resp {
		ongoingTrips = append(ongoingTrips, trp.DeviceID.String)
	}

	return c.JSON(ongoingTrips)
}

func (p *TripDataQueryService) AllUsers(c *fiber.Ctx) error {

	ongoingTrips := make([]string, 0)

	resp, err := models.Fulltrips(qm.Distinct(models.FulltripColumns.DeviceID)).All(c.Context(), p.Pg)
	if err != nil {
		return err
	}

	for _, trp := range resp {
		ongoingTrips = append(ongoingTrips, trp.DeviceID.String)
	}

	return c.JSON(ongoingTrips)
}

func (p *TripDataQueryService) DeviceTripOngoing(c *fiber.Ctx) error {
	deviceID := c.Params("id")

	userID := getUserID(c)
	fmt.Println("uesrID: ", userID, " deviceID: ", deviceID)
	exists, err := p.UserDeviceBelongsToUserID(c.Context(), userID, deviceID)
	if err != nil {
		return err
	}
	if !exists {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	mods := []qm.QueryMod{
		models.FulltripWhere.DeviceID.EQ(null.StringFrom(deviceID)),
		models.FulltripWhere.TripEnd.IsNull(),
	}

	resp, err := models.Fulltrips(mods...).All(c.Context(), p.Pg)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

func (p *TripDataQueryService) AllDeviceTrips(c *fiber.Ctx) error {
	fmt.Println("all device trips start")
	deviceID := c.Params("id")

	userID := getUserID(c)
	fmt.Println("uesrID: ", userID, " deviceID: ", deviceID)
	exists, err := p.UserDeviceBelongsToUserID(c.Context(), userID, deviceID)
	if err != nil {
		fmt.Println("here")
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

	resp, err := models.Fulltrips(mods...).All(c.Context(), p.Pg)
	if err != nil {
		return err
	}

	allTrips := geojson.NewFeatureCollection()

	for _, trp := range resp {
		r := resp[0]
		ts := trp.TripStart.Time.UnixNano() / int64(time.Millisecond)
		te := trp.TripEnd.Time.UnixNano() / int64(time.Millisecond)
		trpData, err := p.TripGeoQuery(r.DeviceID.String, ts, te)
		if err != nil {
			return err
		}

		numResponses := gjson.Get(trpData, "hits.hits.#").Int()
		line := make([][]float64, 0)

		for i := int64(0); i < numResponses; i++ {
			base := fmt.Sprintf("hits.hits.%d.", i)
			lat := gjson.Get(trpData, base+"_source.data.latitude").Float()
			lon := gjson.Get(trpData, base+"_source.data.longitude").Float()
			line = append(line, []float64{lat, lon})
		}
		feature := geojson.NewLineStringFeature(line)
		feature.Geometry.Type = "LineString"
		feature.Properties = map[string]interface{}{"device_id": deviceID, "trip_id": r.TripID, "trip_start": r.TripStart, "trip_end": r.TripEnd}
		allTrips.AddFeature(feature)

	}

	return c.JSON(allTrips)
}

func (das *TripDataQueryService) UserDeviceBelongsToUserID(ctx context.Context, userID, userDeviceID string) (bool, error) {
	device, err := das.GetUserDevice(ctx, userDeviceID)
	if err != nil {
		return false, err
	}
	return device.UserId == userID, nil
}

func (das *TripDataQueryService) GetUserDevice(ctx context.Context, userDeviceID string) (*pb.UserDevice, error) {
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

func (tdqs *TripDataQueryService) executeESQuery(q interface{}) (string, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(q); err != nil {
		return "", err
	}

	res, err := tdqs.es.Search(
		tdqs.es.Search.WithContext(context.Background()),
		tdqs.es.Search.WithIndex(tdqs.elasticIndex),
		tdqs.es.Search.WithBody(&buf),
	)
	if err != nil {
		tdqs.log.Err(err).Msg("Could not query Elasticsearch")
		return "", err
	}
	defer res.Body.Close()

	responseBytes, err := io.ReadAll(res.Body)
	if err != nil {
		tdqs.log.Err(err).Msg("Could not parse Elasticsearch response body")
		return "", err
	}
	response := string(responseBytes)

	if res.StatusCode != 200 {
		tdqs.log.Info().RawJSON("elasticsearchResponseBody", responseBytes).Msg("Error from Elastic.")
		tdqs.log.Info().RawJSON("elasticRequest", buf.Bytes()).Msg("request sent to elastics")

		err := fmt.Errorf("invalid status code when querying elastic: %d", res.StatusCode)
		return response, err
	}

	return response, nil
}

func (tdqs *TripDataQueryService) TripGeoQuery(deviceID string, start, end int64) (string, error) {

	query := QueryTrip{}

	sort := SortByTimestamp{}
	sort.DataTimestamp = "asc"

	match := MatchTerm{}
	match.Match.Subject = deviceID

	date := DataInRange{}
	date.Range.DataTimestamp.Format = "epoch_millis"
	date.Range.DataTimestamp.Gte = start
	date.Range.DataTimestamp.Lte = end

	query.Sort = append(query.Sort, sort)
	query.Source = append(query.Source, "data.latitude")
	query.Source = append(query.Source, "data.longitude")
	query.Size = 10000
	query.Query.Bool.Must = append(query.Query.Bool.Must, match)
	query.Query.Bool.Filter = append(query.Query.Bool.Filter, date)

	response, err := tdqs.executeESQuery(query)
	if err != nil {
		return "", err
	}

	return response, nil
}

type QueryTrip struct {
	Sort   []SortByTimestamp `json:"sort"`
	Source []string          `json:"_source"`
	Size   int               `json:"size"`
	Query  struct {
		Bool struct {
			Must   []MatchTerm   `json:"must"`
			Filter []DataInRange `json:"filter"`
		} `json:"bool"`
	} `json:"query"`
}

type SortByTimestamp struct {
	DataTimestamp string `json:"data.timestamp"`
}

type MatchTerm struct {
	Match struct {
		Subject string `json:"subject"`
	} `json:"match"`
}

type DataInRange struct {
	Range struct {
		DataTimestamp struct {
			Format string `json:"format"`
			Lte    int64  `json:"lte"`
			Gte    int64  `json:"gte"`
		} `json:"data.timestamp"`
	} `json:"range"`
}
