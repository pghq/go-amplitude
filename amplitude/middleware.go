package amplitude

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mssola/user_agent"
	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

const (
	// MaxEvents is the max number of events to track before dropping occurs
	MaxEvents = 1024

	// MaxErrors is the max number of errors to track before dropping occurs
	MaxErrors = 1024
)

// SendMiddleware creates a new http middleware instance from the client
func (c *Client) SendMiddleware() *SendMiddleware {
	return NewSendMiddleware(c)
}

// SendMiddleware is a http middleware for sending request metrics to amplitude
type SendMiddleware struct {
	version      string
	environment  string
	userHeader   string
	deviceHeader string
	events       chan *Event
	errors       chan error
	client       *Client
}

// Environment sets the environment the app is running
func (m *SendMiddleware) Environment(env string) *SendMiddleware {
	m.environment = env
	return m
}

// DeviceHeader sets the header to retrieve device id from
func (m *SendMiddleware) DeviceHeader(h string) *SendMiddleware {
	m.deviceHeader = h
	return m
}

// UserHeader sets the header to retrieve user id from
func (m *SendMiddleware) UserHeader(h string) *SendMiddleware {
	m.userHeader = h
	return m
}

// Version sets the version of your app
func (m *SendMiddleware) Version(v string) *SendMiddleware {
	m.version = v
	return m
}

// Error gets any errors that have occurred
func (m *SendMiddleware) Error() error {
	select {
	case err := <-m.errors:
		return err
	default:
		return nil
	}
}

// Event gets any events that have occurred
func (m *SendMiddleware) Event() *Event {
	select {
	case event := <-m.events:
		return event
	default:
		return nil
	}
}

// SendError buffers an error
func (m *SendMiddleware) SendError(err error) *SendMiddleware {
	select {
	case m.errors <- err:
	default:
	}

	return m
}

// Send buffers an event
func (m *SendMiddleware) Send(event *Event) *SendMiddleware {
	select {
	case m.events <- event:
	default:
		err := errors.Newf("amplitude: event=%s, user=%s, device=%s was dropped",
			event.Name, event.UserId, event.DeviceId)
		m.SendError(err)
	}

	return m
}

func (m *SendMiddleware) Handle(next http.Handler) http.Handler {
	if m.userHeader == "" && m.deviceHeader == "" {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if r.Header.Get(m.userHeader) == "" && r.Header.Get(m.deviceHeader) == "" {
			return
		}

		event := NewEventFromRequest(r, r.Header.Get(m.userHeader), r.Header.Get(m.deviceHeader)).
			Latency(time.Now().Sub(start).Milliseconds()).
			Environment(m.environment).
			Version(m.version)

		m.Send(event)
	})
}

// NewSendMiddleware create a new instance of the amplitude middleware
func NewSendMiddleware(c *Client) *SendMiddleware {
	m := SendMiddleware{
		events: make(chan *Event, MaxEvents),
		errors: make(chan error, MaxErrors),
		client: c,
	}

	return &m
}

// Latency sets the latency of the event
func (e *Event) Latency(latency int64) *Event {
	e.Properties["latency"] = latency

	return e
}

// Environment sets the environment of the event
func (e *Event) Environment(env string) *Event {
	if env != "" {
		e.Properties["environment"] = env
	}

	return e
}

// Version sets the app version of the event
func (e *Event) Version(v string) *Event {
	e.AppVersion = v

	return e
}

// NewEventFromRequest creates a new amplitude event from a http request
func NewEventFromRequest(r *http.Request, userId, deviceId string) *Event {
	event := Event{
		UserId:     userId,
		DeviceId:   deviceId,
		Name:       fmt.Sprintf("%s %s", r.Method, r.URL.Path),
		IP:         r.Header.Get("X-Forwarded-For"),
		Language:   r.Header.Get("Accept-Language"),
		Time:       time.Now().UnixMilli(),
		Properties: make(map[string]interface{}),
	}

	ua := user_agent.New(r.UserAgent())
	event.OSName, event.OSVersion = ua.Browser()
	event.DeviceManufacturer, event.DeviceModel = ua.Engine()
	event.Platform = ua.Platform()

	for k, v := range r.URL.Query() {
		event.Properties[k] = v
	}

	return &event
}
