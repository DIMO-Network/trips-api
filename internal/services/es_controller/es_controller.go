package es_controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const elasticSearchMaxSize = 10000

type Client struct {
	Client *elasticsearch.Client
	Index  string
}

func New(settings *config.Settings) (*Client, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{settings.ElasticHost},
		Username:  settings.ElasticUsername,
		Password:  settings.ElasticPassword,
	})
	if err != nil {
		return nil, err
	}
	return &Client{Client: es, Index: settings.ElasticIndex}, nil
}

func (c *Client) FetchData(deviceID, start, end string) ([]byte, error) {
	var searchAfter int64
	query := QueryTrip{
		Sort: []map[string]string{
			{"data.timestamp": "asc"},
		},
		Query: map[string]any{
			"bool": map[string]any{
				"filter": []map[string]any{
					{
						"range": map[string]any{
							"data.timestamp": map[string]interface{}{
								"format": "strict_date_optional_time",
								"gte":    start,
								"lte":    end,
							},
						},
					},
				},
				"must": map[string]any{
					"match": map[string]string{
						"subject": deviceID,
					},
				},
			},
		},
		Size: elasticSearchMaxSize,
	}

	response, err := c.executeESQuery(query)
	if err != nil {
		return []byte{}, err
	}

	n := gjson.GetBytes(response, "hits.hits.#").Int()
	searchAfter = gjson.GetBytes(response, fmt.Sprintf("hits.hits.%d.sort.0", n-1)).Int()

	for searchAfter > 0 {
		query.SearchAfter = []int64{searchAfter}
		resp, err := c.executeESQuery(query)
		if err != nil {
			return []byte{}, err
		}

		n = gjson.GetBytes(resp, "hits.hits.#").Int()
		searchAfter = gjson.GetBytes(resp, fmt.Sprintf("hits.hits.%d.sort.0", n-1)).Int()

		for i := 0; int64(i) < n; i++ {
			response, err = sjson.SetBytes(response, "hits.hits.-1", gjson.GetBytes(resp, fmt.Sprintf("hits.hits.%d", i)).String())
			if err != nil {
				return []byte{}, err
			}
		}
	}

	return response, nil
}

func (c *Client) executeESQuery(query any) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return []byte{}, err
	}

	res, err := c.Client.Search(
		c.Client.Search.WithContext(context.Background()),
		c.Client.Search.WithIndex(c.Index),
		c.Client.Search.WithBody(&buf),
	)
	if err != nil {
		// c.logger.Err(err).Msg("Could not query Elasticsearch")
		return []byte{}, err
	}
	defer res.Body.Close()

	responseBytes, err := io.ReadAll(res.Body)
	if err != nil {
		// c.logger.Err(err).Msg("Could not parse Elasticsearch response body")
		return responseBytes, err
	}

	if res.StatusCode >= 400 {
		err := fmt.Errorf("invalid status code when querying elastic: %d", res.StatusCode)
		return responseBytes, err
	}

	return responseBytes, nil
}

type QueryTrip struct {
	Sort        []map[string]string `json:"sort"`
	Source      []string            `json:"_source,omitempty"`
	Size        int                 `json:"size"`
	Query       map[string]any      `json:"query"`
	SearchAfter []int64             `json:"search_after,omitempty"`
}
