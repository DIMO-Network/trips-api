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

const pageSize = 1000

func (s *Client) FetchData(ctx context.Context, userDeviceID string, start, end time.Time) ([]byte, error) {
	ElasticSearchRequestTotal.Inc()
	timer := prometheus.NewTimer(ElasticSearchRequestDuration)
	defer timer.ObserveDuration()

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

	return buf.Bytes(), nil
}

var (
	ElasticSearchRequestTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "trips_api",
			Subsystem: "elasticsearch",
			Name:      "requests_total",
			Help:      "The total number of queries made to ElasticSearch as a result of completed Segments.",
		},
	)

	ElasticSearchRequestDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "trips_api",
			Subsystem: "elasticsearch",
			Name:      "request_duration_seconds",
			Help:      "The distribution of request duration to ElasticSearch server in seconds.",
			Buckets:   []float64{1, 2.5, 5, 10, 25, 50, 100, 250},
		},
	)
)
