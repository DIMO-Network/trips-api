package escontroller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/rs/zerolog"
)

type ElasticSearchQueryService struct {
	es           *elasticsearch.Client
	log          *zerolog.Logger
	elasticIndex string
}

func NewESQueryService(settings config.Settings) *ElasticSearchQueryService {

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{settings.ElasticSearchAnalyticsHost},
		Username:  settings.ElasticSearchAnalyticsUsername,
		Password:  settings.ElasticSearchAnalyticsPassword,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(settings.ElasticSearchAnalyticsHost, settings.ElasticSearchAnalyticsUsername, settings.ElasticSearchAnalyticsPassword)

	return &ElasticSearchQueryService{es: esClient, elasticIndex: settings.ElasticSearchIndex}
}

func (eqs *ElasticSearchQueryService) executeESQuery(q interface{}) (string, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(q); err != nil {
		return "", err
	}

	res, err := eqs.es.Search(
		eqs.es.Search.WithContext(context.Background()),
		eqs.es.Search.WithIndex(eqs.elasticIndex),
		eqs.es.Search.WithBody(&buf),
	)
	if err != nil {
		eqs.log.Err(err).Msg("Could not query Elasticsearch")
		return "", err
	}
	defer res.Body.Close()

	responseBytes, err := io.ReadAll(res.Body)
	if err != nil {
		eqs.log.Err(err).Msg("Could not parse Elasticsearch response body")
		return "", err
	}
	response := string(responseBytes)

	if res.StatusCode != 200 {
		eqs.log.Info().RawJSON("elasticsearchResponseBody", responseBytes).Msg("Error from Elastic.")
		eqs.log.Info().RawJSON("elasticRequest", buf.Bytes()).Msg("request sent to elastics")

		err := fmt.Errorf("invalid status code when querying elastic: %d", res.StatusCode)
		return response, err
	}

	return response, nil
}

func (eqs *ElasticSearchQueryService) TripGeoQuery(deviceID, start, end string) (string, error) {

	query := QueryTrip{}

	sort := SortByTimestamp{}
	sort.DataTimestamp = "asc"

	match := MatchTerm{}
	match.Match.Subject = deviceID

	date := DataInRange{}
	date.Range.DataTimestamp.Format = "strict_date_optional_time"
	date.Range.DataTimestamp.Gte = start
	date.Range.DataTimestamp.Lte = end

	query.Sort = append(query.Sort, sort)
	query.Source = append(query.Source, "data.latitude")
	query.Source = append(query.Source, "data.longitude")
	query.Size = 10000
	query.Query.Bool.Must = append(query.Query.Bool.Must, match)
	query.Query.Bool.Filter = append(query.Query.Bool.Filter, date)

	response, err := eqs.executeESQuery(query)
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
			Lte    string `json:"lte"`
			Gte    string `json:"gte"`
		} `json:"data.timestamp"`
	} `json:"range"`
}
