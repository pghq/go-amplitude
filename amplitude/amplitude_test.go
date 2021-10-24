package amplitude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/diagnostic/log"
	"github.com/stretchr/testify/assert"
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

func TestBatchEventsSuccessSummary_String(t *testing.T) {
	t.Run("no summary", func(t *testing.T) {
		s := BatchEventsSuccessSummary{}
		want := "0 events ingested (0 bytes) at 1970-01-01 00:00:00 +0000 UTC"

		if s.String() != want {
			t.Errorf("Summary = %s;\n Expected %s", s.String(), want)
		}
	})

	t.Run("summary with all fields", func(t *testing.T) {
		s := BatchEventsSuccessSummary{
			Code:           200,
			EventsIngested: 5,
			PayloadSize:    10,
			UploadTime:     86400 * 1000,
		}
		want := "5 events ingested (10 bytes) at 1970-01-02 00:00:00 +0000 UTC"

		if s.String() != want {
			t.Errorf("Summary = %s;\n Expected %s", s.String(), want)
		}
	})
}

func TestBatchEventUploadService_Send(t *testing.T) {
	t.Run("no events", func(t *testing.T) {
		client, _, teardown := setup()
		defer teardown()

		_, err := client.Events.Send(context.TODO())

		if err == nil {
			t.Fatal("Error is nil;\n Expected value")
		}

		if got, want := err.Error(), "no events to send"; got != want {
			t.Errorf("Error = %s;\n Expected %s", got, want)
		}
	})

	t.Run("no context", func(t *testing.T) {
		client, _, teardown := setup()
		defer teardown()

		_, err := client.Events.Send(nil, &Event{})

		if err == nil {
			t.Fatal("Error is nil;\n Expected value")
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		client, _, teardown := setup()
		defer teardown()

		ctx, cancel := context.WithTimeout(context.TODO(), 0)
		defer cancel()
		_, err := client.Events.Send(ctx, &Event{})

		if err == nil {
			t.Fatal("Error is nil;\n Expected value")
		}

		if got, want := err.Error(), ctx.Err().Error(); got != want {
			t.Errorf("Error = %v;\n Expected %v", got, want)
		}
	})

	t.Run("200 ok", func(t *testing.T) {
		client, mux, teardown := setup()
		defer teardown()

		var event Event
		data, _ := ioutil.ReadFile("testdata/event.json")
		_ = json.Unmarshal(data, &event)
		events := []*Event{&event}

		mux.HandleFunc(batchEventUploadEndpoint, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Method = %s;\n Expected %s", r.Method, http.MethodPost)
			}

			want := make(map[string]interface{})
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"events": [%s]}`, data)), &want)

			got := make(map[string]interface{})
			_ = json.NewDecoder(r.Body).Decode(&got)

			if !reflect.DeepEqual(got["events"], want["events"]) {
				t.Errorf("Body[\"events\"] = %v;\n Expected %v", got["events"], want["events"])
			}

			_, _ = fmt.Fprint(w, `{
			  "code": 200,
			  "events_ingested": 50,
			  "payload_size_bytes": 50,
			  "server_upload_time": 1396381378123
			}`)
		})

		resp, err := client.Events.Send(context.TODO(), events...)

		if err != nil {
			t.Fatalf("Error = %v;\n Expected no error", err)
		}

		if got, want := resp.Code, 200; got != want {
			t.Errorf("Code = %d;\n Expected %d", got, want)
		}

		if got, want := resp.EventsIngested, 50; got != want {
			t.Errorf("EventsIngested = %d;\n Expected %d", got, want)
		}

		if got, want := resp.PayloadSize, 50; got != want {
			t.Errorf("PayloadSize = %d;\n Expected %d", got, want)
		}

		if got, want := resp.UploadTime, int64(1396381378123); got != want {
			t.Errorf("UploadTime = %d;\n Expected %d", got, want)
		}
	})
}

func TestSendMiddleware(t *testing.T){
	c := New("")

	t.Run("can create instance", func(t *testing.T) {
		m := c.SendMiddleware().
			Environment("test").
			DeviceHeader("device").
			UserHeader("user").
			Version("v0")
		assert.NotNil(t, m)
		assert.Equal(t, c, m.client)
		assert.Equal(t, "test", m.environment)
		assert.Equal(t, "device", m.deviceHeader)
		assert.Equal(t, "user", m.userHeader)
		assert.Equal(t, "v0", m.version)
		assert.NotNil(t, m.events)
		assert.Equal(t, cap(m.events), MaxEvents)
		assert.Equal(t, cap(m.errors), MaxErrors)
	})

	t.Run("raises errors", func(t *testing.T) {
		m := c.SendMiddleware()
		m = m.SendError(errors.New("an error has occurred"))
		assert.NotNil(t, m)
		assert.NotNil(t, m.Error())
		assert.Nil(t, m.Error())

		for i := 0; i <= cap(m.errors); i++{
			m.SendError(errors.New("an error has occurred"))
		}

		assert.Equal(t, cap(m.errors), len(m.errors))
	})

	t.Run("sends events", func(t *testing.T) {
		m := c.SendMiddleware()
		m = m.Send(&Event{})
		assert.NotNil(t, m)
		assert.NotNil(t, m.Event())
		assert.Nil(t, m.Event())

		for i := 0; i <= cap(m.events); i++{
			m.Send(&Event{})
		}

		assert.Equal(t, cap(m.events), len(m.events))
	})

	t.Run("no device or user id", func(t *testing.T) {
		m := c.SendMiddleware()
		r := httptest.NewRequest("GET", "/tests", nil)
		w := httptest.NewRecorder()
		m.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(w, r)
		m.DeviceHeader("Device-Id")
		m.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(w, r)

		assert.Empty(t, m.events)
	})

	t.Run("handles http requests", func(t *testing.T) {
		m := c.SendMiddleware().Environment("test").
			DeviceHeader("Device-Id").
			UserHeader("User-Id").
			Version("v0")

		r := httptest.NewRequest("GET", "/tests?page=1234", nil)
		r.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/603.3.8 (KHTML, like Gecko) Version/10.1.2 Safari/603.3.8")
		r.Header.Set("Device-Id", "test")
		r.Header.Set("User-Id", "test")
		r.Header.Set("X-Forwarded-For", "1.2.3.4")
		r.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
		w := httptest.NewRecorder()
		m.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-time.After(1 * time.Millisecond)
		})).ServeHTTP(w, r)

		assert.NotEmpty(t, m.events)
		event := m.Event()
		assert.Equal(t, "GET /tests", event.Name)
		assert.Equal(t, "test", event.DeviceId)
		assert.Equal(t, "test", event.UserId)
		assert.Nil(t, event.UserProperties)
		assert.NotNil(t, event.Properties)
		assert.Equal(t, "test", event.Properties["environment"])
		assert.Equal(t, int64(1), event.Properties["latency"])
		assert.Equal(t, []string{"1234"}, event.Properties["page"])
		assert.Greater(t, event.Time, int64(0))
		assert.Equal(t, "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", event.Language)
		assert.Equal(t, "1.2.3.4", event.IP)
		assert.Equal(t, "Safari", event.OSName)
		assert.Equal(t, "10.1.2", event.OSVersion)
		assert.Equal(t, "AppleWebKit", event.DeviceManufacturer)
		assert.Equal(t, "603.3.8", event.DeviceModel)
		assert.Equal(t, "Macintosh", event.Platform)
	})

	t.Run("raises flush context errors", func(t *testing.T) {
		m := c.SendMiddleware()
		ctx, cancel := context.WithTimeout(context.Background(), 0)
		defer cancel()

		m.Send(&Event{})
		m.Flush(ctx)

		assert.NotEmpty(t, m.events)
	})

	t.Run("ignores no events", func(t *testing.T) {
		m := c.SendMiddleware()
		ctx := context.Background()

		m.Flush(ctx)
		assert.Empty(t, m.events)
	})

	t.Run("raises send errors", func(t *testing.T) {
		m := c.SendMiddleware()
		ctx := context.Background()

		m.Send(&Event{})
		m.Flush(ctx)

		assert.Empty(t, m.events)
		assert.NotEmpty(t, m.errors)
	})

	t.Run("handles events", func(t *testing.T) {
		log.Writer(io.Discard)
		defer log.Reset()
		c, mux, teardown := setup()
		defer teardown()

		mux.HandleFunc(batchEventUploadEndpoint, func(w http.ResponseWriter, r *http.Request) {})
		m := c.SendMiddleware()
		ctx := context.Background()

		m.Send(&Event{}).Send(&Event{})
		m.Flush(ctx)

		assert.Empty(t, m.events)
	})

}
