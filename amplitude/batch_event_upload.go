package amplitude

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

const (
	batchEventUploadEndpoint = "/batch"
)

// BatchEventUploadService provides a service for sending analytics events in bulk to Amplitude
type BatchEventUploadService service

// Event is the base analytic structure for capturing user activity
type Event struct {
	UserID              string                 `json:"user_id,omitempty"`
	DeviceID            string                 `json:"device_id,omitempty"`
	Type                string                 `json:"event_type"`
	Time                uint                   `json:"time,omitempty"`
	EventProperties     map[string]interface{} `json:"event_properties,omitempty"`
	UserProperties      map[string]interface{} `json:"user_properties,omitempty"`
	Groups              map[string]interface{} `json:"groups,omitempty"`
	GroupProperties     map[string]interface{} `json:"group_properties,omitempty"`
	AppVersion          string                 `json:"app_version,omitempty"`
	Platform            string                 `json:"platform,omitempty"`
	OSName              string                 `json:"os_name,omitempty"`
	OSVersion           string                 `json:"os_version,omitempty"`
	DeviceBrand         string                 `json:"device_brand,omitempty"`
	DeviceManufacturer  string                 `json:"device_manufacturer,omitempty"`
	Carrier             string                 `json:"carrier,omitempty"`
	Country             string                 `json:"country,omitempty"`
	Region              string                 `json:"region,omitempty"`
	City                string                 `json:"city,omitempty"`
	DMA                 string                 `json:"dma,omitempty"`
	Language            string                 `json:"language,omitempty"`
	Price               float64                `json:"price,omitempty"`
	Quantity            int                    `json:"quantity,omitempty"`
	Revenue             float64                `json:"revenue,omitempty"`
	ProductID           string                 `json:"productId,omitempty"`
	RevenueType         string                 `json:"revenueType,omitempty"`
	Latitude            float64                `json:"location_lat,omitempty"`
	Longitude           float64                `json:"location_lng,omitempty"`
	IP                  net.Addr               `json:"ip,omitempty"`
	IOSAdvertiserID     string                 `json:"idfa,omitempty"`
	IOSVendorID         string                 `json:"idfv,omitempty"`
	AndroidAdvertiserID string                 `json:"adid,omitempty"`
	AndroidID           string                 `json:"android_id,omitempty"`
	ID                  int                    `json:"event_id,omitempty"`
	SessionID           int                    `json:"session_id,omitempty"`
	InsertID            string                 `json:"insert_id,omitempty"`
}

// BatchEventsSuccessSummary is expected to be returned for all successful requests
type BatchEventsSuccessSummary struct {
	Code           int   `json:"code"`
	EventsIngested int   `json:"events_ingested"`
	PayloadSize    int   `json:"payload_size_bytes"`
	UploadTime     int64 `json:"server_upload_time"`
}

// String converts the summary response to a pretty string format
func (s *BatchEventsSuccessSummary) String() string {
	timestamp := time.Unix(0, s.UploadTime*int64(time.Millisecond)).UTC()
	return fmt.Sprintf("%d events ingested (%d bytes) at %s", s.EventsIngested, s.PayloadSize, timestamp)
}

// Send analytics events to Amplitude using the BatchEventUpload API
func (s *BatchEventUploadService) Send(ctx context.Context, events []Event) (*BatchEventsSuccessSummary, error) {
	if len(events) == 0 {
		return nil, errors.New("no events to send")
	}

	body := s.client.NewRequestBody().WithValue("events", events)
	req, err := s.client.NewRequest(ctx, http.MethodPost, batchEventUploadEndpoint, body)
	if err != nil {
		return nil, err
	}

	var res BatchEventsSuccessSummary
	if _, err := s.client.Do(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
