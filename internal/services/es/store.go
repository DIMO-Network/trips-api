package es

import (
	"bytes"
	"context"
	"time"

	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/some"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Client struct {
	typedClient  *elasticsearch.TypedClient
	indexPattern string
}

func New(settings *config.Settings) (*Client, error) {
	es, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{settings.ElasticHost},
		Username:  settings.ElasticUsername,
		Password:  settings.ElasticPassword,
	})
	if err != nil {
		return nil, err
	}

	return &Client{typedClient: es, indexPattern: settings.ElasticIndex}, nil
}

const pageSize = 100

func (s *Client) FetchData(ctx context.Context, userDeviceID string, start, end time.Time) ([]byte, error) {
	elasticSearchQueryCount.Inc()
	esQueryTimer := time.Now()
	var buf bytes.Buffer

	buf.WriteByte('[')

	req := &search.Request{
		Query: &types.Query{
			Bool: &types.BoolQuery{
				Filter: []types.Query{
					{
						Term: map[string]types.TermQuery{
							"subject": {Value: userDeviceID},
						},
					},
					{
						Range: map[string]types.RangeQuery{
							"data.timestamp": types.DateRangeQuery{
								Gte: some.String(start.Format(time.RFC3339)),
								Lte: some.String(end.Format(time.RFC3339)),
							},
						},
					},
				},
			},
		},
		Size: some.Int(pageSize),
		Sort: []types.SortCombinations{
			types.SortOptions{SortOptions: map[string]types.FieldSort{
				"time": {Order: &sortorder.Asc},
			}},
		},
	}

	needComma := false

	for {
		resp, err := s.typedClient.Search().Request(req).Do(ctx)
		if err != nil {
			return nil, err
		}

		hitCount := len(resp.Hits.Hits)
		if hitCount == 0 {
			break
		}

		for _, h := range resp.Hits.Hits {
			if needComma {
				buf.WriteByte(',')
			} else {
				needComma = true
			}
			buf.Write(h.Source_)
		}

		req.SearchAfter = resp.Hits.Hits[hitCount-1].Sort
	}
	elasticSearchResponseTime.Observe(time.Since(esQueryTimer).Seconds())
	return buf.Bytes(), nil
}

var (
	elasticSearchQueryCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "trips_api_elasticsearch_query_count",
			Help: "The total number of queries made to ElasticSearch as a result of completed Segments.",
		},
	)

	elasticSearchResponseTime = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "trips_api_http_response_time",
			Help:    "The response time distribution of the ElasticSearch Server in seconds.",
			Buckets: []float64{1, 2.5, 5, 10, 25, 50, 100, 250},
		},
	)
)
