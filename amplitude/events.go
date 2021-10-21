package amplitude

// EventsService provides access to events related functions
type EventsService service

// Event is the base analytic structure for capturing user activity
type Event struct {
	UserId              string                 `json:"user_id,omitempty"`
	DeviceId            string                 `json:"device_id,omitempty"`
	Name                string                 `json:"event_type"`
	Time                int64                  `json:"time,omitempty"`
	Properties          map[string]interface{} `json:"event_properties,omitempty"`
	UserProperties      map[string]interface{} `json:"user_properties,omitempty"`
	Groups              map[string]interface{} `json:"groups,omitempty"`
	GroupProperties     map[string]interface{} `json:"group_properties,omitempty"`
	AppVersion          string                 `json:"app_version,omitempty"`
	Platform            string                 `json:"platform,omitempty"`
	OSName              string                 `json:"os_name,omitempty"`
	OSVersion           string                 `json:"os_version,omitempty"`
	DeviceBrand         string                 `json:"device_brand,omitempty"`
	DeviceManufacturer  string                 `json:"device_manufacturer,omitempty"`
	DeviceModel         string                 `json:"device_model,omitempty"`
	Carrier             string                 `json:"carrier,omitempty"`
	Country             string                 `json:"country,omitempty"`
	Region              string                 `json:"region,omitempty"`
	City                string                 `json:"city,omitempty"`
	DMA                 string                 `json:"dma,omitempty"`
	Language            string                 `json:"language,omitempty"`
	Price               float64                `json:"price,omitempty"`
	Quantity            int                    `json:"quantity,omitempty"`
	Revenue             float64                `json:"revenue,omitempty"`
	ProductId           string                 `json:"productId,omitempty"`
	RevenueType         string                 `json:"revenueType,omitempty"`
	Latitude            float64                `json:"location_lat,omitempty"`
	Longitude           float64                `json:"location_lng,omitempty"`
	IP                  string                 `json:"ip,omitempty"`
	IOSAdvertiserId     string                 `json:"idfa,omitempty"`
	IOSVendorId         string                 `json:"idfv,omitempty"`
	AndroidAdvertiserId string                 `json:"adid,omitempty"`
	AndroidId           string                 `json:"android_id,omitempty"`
	Id                  int                    `json:"event_id,omitempty"`
	SessionId           int                    `json:"session_id,omitempty"`
	InsertId            string                 `json:"insert_id,omitempty"`
}
