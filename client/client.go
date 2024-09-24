package client

import (
	"encoding/json"
	"fmt"
	"github.com/kardolus/maps/http"
	"github.com/kardolus/maps/types"
	"strings"
	"time"
)

const (
	Endpoint         = "https://maps.googleapis.com/maps/api/place/textsearch/json?query=%s&key=%s"
	NextPageEndpoint = "https://maps.googleapis.com/maps/api/place/textsearch/json?pagetoken=%s&key=%s"
	ErrMissingEntity = "entity required"
)

type Client struct {
	caller  http.Caller
	timeout int
	apiKey  string
}

// WithTimeout configures the timeout in milliseconds for the pagination loop
func (c *Client) WithTimeout(timeout int) *Client {
	c.timeout = timeout
	return c
}

func New(caller http.Caller, apiKey string) *Client {
	return &Client{
		caller: caller,
		apiKey: apiKey,
	}
}

func (c *Client) FetchLocations(entity string) ([]types.Response, error) {
	var (
		result []types.Response
		record types.Response
	)

	if entity == "" {
		return nil, fmt.Errorf(ErrMissingEntity)
	}

	// First request
	bytes, err := c.caller.Get(c.constructURL(entity))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return nil, err
	}

	result = append(result, record)

	// Paginate through results using next_page_token
	for record.NextPageToken != "" {
		time.Sleep(time.Duration(c.timeout) * time.Millisecond)

		var nextRecord types.Response
		bytes, err := c.caller.Get(c.constructNextURL(record.NextPageToken))
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(bytes, &nextRecord); err != nil {
			return nil, err
		}

		result = append(result, nextRecord)
		record = nextRecord
	}

	return result, nil
}

func (c *Client) buildQuery(entity string) string {
	words := strings.Split(entity, " ")
	return strings.Join(words, "+")
}

func (c *Client) constructURL(entity string) string {
	query := c.buildQuery(entity)
	return fmt.Sprintf(Endpoint, query, c.apiKey)
}

func (c *Client) constructNextURL(token string) string {
	return fmt.Sprintf(NextPageEndpoint, token, c.apiKey)
}
