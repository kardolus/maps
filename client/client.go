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

func (c *Client) FetchAllLocations(queries, contains, matches []string) ([]types.Location, error) {
	fmt.Printf("queries %v\n", queries)                          // TODO move to debug
	fmt.Printf("contains: %v\nmatches: %v\n", contains, matches) // TODO move to debug

	var result []types.Location

	found := make(map[string]struct{})

	for _, query := range queries {
		locations, err := c.FetchLocations(query, contains, matches)

		if err != nil {
			return nil, err
		}

		for _, location := range locations {
			if _, ok := found[location.PlaceId]; !ok {
				result = append(result, location)
				found[location.PlaceId] = struct{}{}
			}
		}
	}

	return result, nil
}

func (c *Client) FetchLocations(entity string, contains, matches []string) ([]types.Location, error) {
	var (
		result []types.Location
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

	// Filter results based on contains and matches
	for _, location := range record.Results {
		if containsAny(location.Name, contains) || matchesAny(location.Name, matches) {
			result = append(result, location)
		}
	}

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

		for _, location := range nextRecord.Results {
			if containsAny(location.Name, contains) || matchesAny(location.Name, matches) {
				result = append(result, location)
			}
		}

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

func containsAny(name string, list []string) bool {
	for _, item := range list {
		item = strings.TrimSpace(item)
		if strings.Contains(strings.ToLower(name), strings.ToLower(item)) {
			return true
		}
	}
	return false
}

func matchesAny(name string, list []string) bool {
	for _, item := range list {
		item = strings.TrimSpace(item)
		if strings.ToLower(name) == strings.ToLower(item) {
			return true
		}
	}
	return false
}
