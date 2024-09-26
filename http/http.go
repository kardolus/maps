package http

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

const (
	contentType              = "application/json"
	errFailedToRead          = "failed to read response: %w"
	errFailedToCreateRequest = "failed to create request: %w"
	errFailedToMakeRequest   = "failed to make request: %w"
	errHTTP                  = "http status %d: %s"
	headerContentType        = "Content-Type"
)

type Caller interface {
	Get(url string) ([]byte, error)
}

type RestCaller struct {
	client  *http.Client
	retries int
}

// Ensure RestCaller implements Caller interface
var _ Caller = &RestCaller{}

// New creates a new RestCaller with a default http.Client
func New() *RestCaller {
	return &RestCaller{
		client: &http.Client{},
	}
}

// WithRetries configures the number of retries
func (r *RestCaller) WithRetries(retries int) *RestCaller {
	r.retries = retries
	return r
}

// Get performs a GET request with retry logic
func (r *RestCaller) Get(url string) ([]byte, error) {
	fmt.Printf("%s\n\n\n", url) // TODO create a debug setting (maybe create a "config" struct)

	var result []byte
	var err error

	for attempt := 0; attempt <= r.retries; attempt++ {
		result, err = r.doRequest(http.MethodGet, url, nil)
		if err == nil {
			return result, nil // successful request, return response
		}

		// Apply exponential backoff with randomness for retries
		if attempt < r.retries {
			backoff := r.calculateBackoff(attempt)
			time.Sleep(backoff)
		}
	}

	// Return last error after all retries have failed
	return nil, fmt.Errorf("request failed after %d attempts: %w", r.retries+1, err)
}

// Internal method to perform HTTP request
func (r *RestCaller) doRequest(method, url string, body []byte) ([]byte, error) {
	req, err := r.newRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf(errFailedToCreateRequest, err)
	}

	response, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(errFailedToMakeRequest, err)
	}
	defer response.Body.Close()

	result, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf(errFailedToRead, err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf(errHTTP, response.StatusCode, string(result))
	}

	return result, nil
}

// newRequest creates a new HTTP request with headers
func (r *RestCaller) newRequest(method, url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set(headerContentType, contentType)
	return req, nil
}

// calculateBackoff calculates an exponential backoff with some jitter
func (r *RestCaller) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff with randomness (jitter)
	base := time.Duration(100 * time.Millisecond)              // base delay
	jitter := time.Duration(rand.Intn(100)) * time.Millisecond // random jitter up to 100ms
	backoff := (1 << attempt) * base                           // exponential backoff

	return backoff + jitter
}
