package amplitude

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

const (
	batchEventUploadEndpoint = "/batch"
)

// BatchEventsSuccessSummary is expected to be returned for all successful requests.
type BatchEventsSuccessSummary struct {
	Code           int   `json:"code"`
	EventsIngested int   `json:"events_ingested"`
	PayloadSize    int   `json:"payload_size_bytes"`
	UploadTime     int64 `json:"server_upload_time"`
}

// String converts the summary response to a pretty string format.
func (s *BatchEventsSuccessSummary) String() string {
	timestamp := time.Unix(0, s.UploadTime*int64(time.Millisecond)).UTC()
	return fmt.Sprintf("%d events ingested (%d bytes) at %s", s.EventsIngested, s.PayloadSize, timestamp)
}

// Send batches of events via the Batch Event Upload API
// This endpoint is recommended for Customers that want to send large batches of data at a time,
// for example through scheduled jobs, rather than in a continuous realtime stream.
// Due to the higher rate of data that is permitted to this endpoint, data sent to this endpoint
// may be delayed based on load.
func (s *EventsService) Send(ctx context.Context, events ...*Event) (*BatchEventsSuccessSummary, error) {
	if len(events) == 0 {
		return nil, errors.New("no events to send")
	}

	body := s.client.NewRequestBody().WithValue("events", events)
	req, err := s.client.NewRequest(ctx, http.MethodPost, batchEventUploadEndpoint, body)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	var res BatchEventsSuccessSummary
	if _, err := s.client.Do(req, &res); err != nil {
		return nil, errors.Wrap(err)
	}

	return &res, nil
}
