// Copyright 2021 PGHQ. All Rights Reserved.
//
// Licensed under the GNU General Public License, Version 3 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package amplitude provides a http client for working with various Amplitude APIs.
package amplitude

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	defaultBaseURL     = "https://api2.amplitude.com"
	defaultUserAgent   = "go-amplitude/v0"
	defaultMediaType   = "application/json"
	defaultContentType = "application/json"
)

// Client allows interaction with amplitude services
type Client struct {
	// BaseURL for all API requests. Exposed services should
	// use relative paths for making requests.
	BaseURL   *url.URL
	UserAgent string
	APIKey    string

	client *http.Client

	// common service is shared between all exposed services
	common service

	Events *EventsService
}

type service struct {
	client *Client
}

// New creates a new Amplitude client
func New(apiKey string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	c := Client{
		client:    http.DefaultClient,
		BaseURL:   baseURL,
		UserAgent: defaultUserAgent,
		APIKey:    apiKey,
	}

	c.common.client = &c
	c.Events = (*EventsService)(&c.common)

	return &c
}

// WithHttpClient sets the underlining http client if default is not enough
func (c *Client) WithHttpClient(client *http.Client) *Client {
	if client != nil {
		c.client = client
	}

	return c
}

// RequestBody is sent as the body of all requests
type RequestBody map[string]interface{}

// NewRequestBody creates a request body with authentication
func (c *Client) NewRequestBody() RequestBody {
	body := make(RequestBody)

	if c.APIKey != "" {
		body.WithValue("api_key", c.APIKey)
	}

	return body
}

// WithValue adds a value to the request body
func (b RequestBody) WithValue(key string, v interface{}) RequestBody {
	b[key] = v

	return b
}

// NewRequest provides a http request to be sent to Amplitude
func (c *Client) NewRequest(ctx context.Context, method, endpoint string, body RequestBody) (*http.Request, error) {
	// if no endpoint is specified, then the base URL is used
	var u *url.URL
	if endpoint == "" {
		u = c.BaseURL
	} else {
		endpointURL, err := url.Parse(endpoint)
		if err != nil {
			return nil, err
		}
		u = c.BaseURL.ResolveReference(endpointURL)
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)

		if err := enc.Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", defaultContentType)
	}

	req.Header.Set("Accept", defaultMediaType)
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	return req, nil
}

// Do a http request to Amplitude and handles the response it receives
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("no request passed")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		if ctx := req.Context(); ctx != nil && ctx.Err() != nil {
			return nil, ctx.Err()
		}

		return nil, err
	}

	// if the response is an API error that we know about, this propagates it up the stack,
	// if its one we don't know about check the error context for additional information
	if err := AsError(resp); err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// we only support json responses (and so does the amplitude API)
	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil && err != io.EOF {
			return nil, err
		}
	}

	return resp, err
}

// Error houses any potential http error responses from the API
// The body is not well-defined so more information can be obtained
// by access the context
type Error struct {
	Response *http.Response
	Context  map[string]interface{}
}

// Code is an integer denoting what error has been received
func (e *Error) Code() int {
	raw := e.Context["code"]
	// The JSON decoder will unmarshal all numbers as floating point values,
	// but the code is actually expected to be an integer.
	if code, ok := raw.(float64); ok {
		return int(code)
	}

	return 0
}

// Message is a human-readable error denoting (hopefully) what has gone wrong
func (e *Error) Message() string {
	raw := e.Context["error"]
	if message, ok := raw.(string); ok {
		return message
	}

	return ""
}

// Error implements the error interface
func (e *Error) Error() string {
	return fmt.Sprintf("error code %d recieved with message %s", e.Code(), e.Message())
}

// AsError reads the response for errors and returns them if present
func AsError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return nil
	}

	err := Error{
		Response: resp,
	}

	if data, _ := io.ReadAll(resp.Body); data != nil {
		ctx := make(map[string]interface{})
		if err := json.Unmarshal(data, &ctx); err != nil {
			return err
		}
		err.Context = ctx
	}

	return &err
}
