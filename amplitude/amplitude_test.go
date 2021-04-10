package amplitude

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// setup returns a mock client, test handler, and teardown function
func setup() (*Client, *http.ServeMux, func()) {
	handler := http.NewServeMux()
	server := httptest.NewServer(handler)

	client := New("")
	u, _ := url.Parse(server.URL)
	client.BaseURL = u

	return client, handler, server.Close
}

// testRequest checks that an http request meets the bare minimum requirements
func testRequest(t *testing.T, req *http.Request, expectedEndpoint string) {
	t.Helper()

	if got, want := req.URL.String(), expectedEndpoint; got != want {
		t.Errorf("URL = %v; expected %v", got, want)
	}

	if got, want := req.Header.Get("Accept"), defaultMediaType; got != want {
		t.Errorf("Accept header = %s; expected %s", got, want)
	}

	if got, want := req.Header.Get("User-Agent"), defaultUserAgent; got != want {
		t.Errorf("User-Agent header = %s; expected %s", got, want)
	}
}

// testError checks that an http response error meets the bare minimum requirements
func testError(t *testing.T, err *Error, expectedCode int, expectedMessage string) {
	t.Helper()

	if err.Response == nil {
		t.Fatal("Response is nil; expected value")
	}

	if err.Response.StatusCode != expectedCode {
		t.Errorf("StatusCode = %d; expected %d", err.Response.StatusCode, expectedCode)
	}

	if err.Code() != expectedCode {
		t.Errorf("Code = %d; expected %d", err.Code(), expectedCode)
	}

	if err.Message() != expectedMessage {
		t.Errorf("Message = %s; expected %s", err.Message(), expectedMessage)
	}

	expectedError := fmt.Sprintf("error code %d recieved with message %s", expectedCode, expectedMessage)
	if err.Error() != expectedError {
		t.Errorf("Error = %s; expected %s", err.Error(), expectedError)
	}
}

func TestNew(t *testing.T) {
	t.Run("basic client instantiation", func(t *testing.T) {
		c := New("your-amplitude-key")

		if got, want := c.BaseURL.String(), defaultBaseURL; got != want {
			t.Errorf("BaseURL = %v; expected %v", got, want)
		}

		if got, want := c.UserAgent, defaultUserAgent; got != want {
			t.Errorf("UserAgent = %v; expected %v", got, want)
		}
	})
}

func TestClient_NewRequestBody(t *testing.T) {
	t.Run("without API key", func(t *testing.T) {
		c := New("")
		body := c.NewRequestBody()

		if body == nil {
			t.Error("Expected body to be returned")
		}

		if len(body) != 0 {
			t.Error("Expected body to be empty")
		}
	})

	t.Run("with API key", func(t *testing.T) {
		c := New("12345")
		body := c.NewRequestBody()

		if got, want := len(body), 1; got != want {
			t.Errorf("Body length = %d; expected %d", got, want)
		}

		if got, want := body["api_key"], "12345"; got != want {
			t.Errorf("API key = %v; expected %s", want, got)
		}
	})
}

// TestRequestBody_WithValue
func TestRequestBody_WithValue(t *testing.T) {
	t.Run("without API key", func(t *testing.T) {
		c := New("")
		body := c.NewRequestBody().WithValue("api_key", "12345")

		if got, want := body["api_key"], "12345"; got != want {
			t.Errorf("API key = %v; expected %s", want, got)
		}
	})
}

func TestClient_WithHttpClient(t *testing.T) {
	t.Run("custom client with timeout", func(t *testing.T) {
		c := New("").WithHttpClient(&http.Client{
			Timeout: time.Second,
		})

		if got, want := c.client.Timeout, time.Second; got != want {
			t.Errorf("Timeout = %v; expected %v", want, got)
		}
	})
}

func TestClient_NewRequest(t *testing.T) {
	t.Run("without endpoint", func(t *testing.T) {
		c := New("")
		req, err := c.NewRequest(context.TODO(), "", "", nil)

		if err != nil {
			t.Fatalf("Error = %v; expected no error", err)
		}

		testRequest(t, req, defaultBaseURL)
	})

	t.Run("with endpoint", func(t *testing.T) {
		c := New("")
		req, err := c.NewRequest(context.TODO(), "", "/endpoint", nil)

		if err != nil {
			t.Fatalf("Error = %v; expected no error", err)
		}

		if got, want := req.URL.String(), defaultBaseURL+"/endpoint"; got != want {
			t.Errorf("URL = %v; expected %v", got, want)
		}

		testRequest(t, req, defaultBaseURL+"/endpoint")
	})

	t.Run("bad endpoint", func(t *testing.T) {
		c := New("")
		_, err := c.NewRequest(context.TODO(), "", ":", nil)

		if err == nil {
			t.Fatal("Error is nil; expected value")
		}
	})

	t.Run("with request body", func(t *testing.T) {
		c := New("")
		req, err := c.NewRequest(context.TODO(), "", "", map[string]interface{}{
			"key": "value",
		})

		if err != nil {
			t.Fatalf("Error = %v; expected no error", err)
		}

		testRequest(t, req, defaultBaseURL)

		body, _ := io.ReadAll(req.Body)

		if got, want := string(body), "{\"key\":\"value\"}\n"; got != want {
			t.Errorf("Body = %s; expected %s", got, want)
		}

		if got, want := req.Header.Get("Content-Type"), defaultContentType; got != want {
			t.Errorf("Content-Type header = %s; expected %s", got, want)
		}
	})

	t.Run("bad request body", func(t *testing.T) {
		c := New("")
		body := RequestBody{
			"errors": make(chan error),
		}
		_, err := c.NewRequest(context.TODO(), "", "", body)

		if err == nil {
			t.Fatal("Error is nil; expected value")
		}
	})

	t.Run("no context", func(t *testing.T) {
		c := New("")
		_, err := c.NewRequest(nil, "", "", nil)

		if err == nil {
			t.Fatal("Error is nil; expected value")
		}
	})
}

func TestClient_Do(t *testing.T) {
	t.Run("no request", func(t *testing.T) {
		c := New("")
		_, err := c.Do(nil, nil)

		if err == nil {
			t.Fatal("Error is nil; expected value")
		}

		if got, want := err.Error(), "no request passed"; got != want {
			t.Errorf("Error = %v; expected %v", got, want)
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		c := New("")
		ctx, cancel := context.WithTimeout(context.TODO(), 0)
		defer cancel()
		req, _ := c.NewRequest(ctx, "", "", nil)
		_, err := c.Do(req, nil)

		if err == nil {
			t.Fatal("Error is nil; expected value")
		}

		if got, want := err, ctx.Err(); got != want {
			t.Errorf("Error = %v; expected %v", got, want)
		}
	})

	t.Run("no request URL", func(t *testing.T) {
		c := New("")
		_, err := c.Do(&http.Request{}, nil)

		if err == nil {
			t.Fatal("Error is nil; expected value")
		}
	})

	t.Run("http error", func(t *testing.T) {
		client, mux, teardown := setup()
		defer teardown()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, `{"code": 400, "error": "bad request"}`, 400)
		})

		req, _ := client.NewRequest(context.TODO(), "GET", ".", nil)
		_, err := client.Do(req, nil)

		if err == nil {
			t.Fatal("Error is nil; expected value")
		}
	})

	t.Run("bad decode value", func(t *testing.T) {
		client, mux, teardown := setup()
		defer teardown()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, "bad response")
		})

		req, _ := client.NewRequest(context.TODO(), "GET", ".", nil)

		_, err := client.Do(req, "")

		if err == nil {
			t.Fatal("Error is nil; expected value")
		}
	})

	t.Run("no response", func(t *testing.T) {
		client, mux, teardown := setup()
		defer teardown()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

		req, _ := client.NewRequest(context.TODO(), "GET", ".", nil)

		resp, err := client.Do(req, nil)

		if err != nil {
			t.Fatalf("Error = %v; expected no error", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("StatusCode = %d; expected 200", resp.StatusCode)
		}
	})
}

func TestAsError(t *testing.T) {
	t.Run("with basic context", func(t *testing.T) {
		resp := &http.Response{StatusCode: 413}
		resp.Body = io.NopCloser(bytes.NewBufferString(`{
			"code": 413, 
			"error": "Payload too large"
		}`))

		err := AsError(resp)
		if err == nil {
			t.Fatal("Error is nil; expected value")
		}

		respError, ok := err.(*Error)
		if !ok {
			t.Fatalf("Error = %v; expected response error", err)
		}

		testError(t, respError, 413, "Payload too large")
	})

	t.Run("without basic context", func(t *testing.T) {
		resp := &http.Response{}
		resp.Body = io.NopCloser(bytes.NewBufferString("{}"))

		err := AsError(resp)
		if err == nil {
			t.Fatal("Error is nil; expected value")
		}

		respError, ok := err.(*Error)
		if !ok {
			t.Fatalf("Error = %v; expected response error", err)
		}

		testError(t, respError, 0, "")
	})

	t.Run("with additional context", func(t *testing.T) {
		resp := &http.Response{StatusCode: 400}
		resp.Body = io.NopCloser(bytes.NewBufferString(`{
			"code": 400, 
			"error": "Request missing required field",
			"missing_field": "api_key"
		}`))

		err := AsError(resp)
		if err == nil {
			t.Fatal("Error is nil; expected value")
		}

		respError, ok := err.(*Error)
		if !ok {
			t.Fatalf("Error = %v; expected response error", err)
		}

		testError(t, respError, 400, "Request missing required field")

		if got, want := respError.Context["missing_field"], "api_key"; got != want {
			t.Errorf("Context missing field = %v; expected %s", got, want)
		}
	})

	t.Run("with bad body", func(t *testing.T) {
		resp := &http.Response{StatusCode: 400}
		resp.Body = io.NopCloser(bytes.NewBufferString("bad response"))

		err := AsError(resp)
		if err == nil {
			t.Fatal("Error is nil; expected value")
		}
	})
}
