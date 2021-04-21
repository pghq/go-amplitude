package amplitude

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

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

		_, err := client.Events.BatchUpload(context.TODO(), nil)

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

		_, err := client.Events.BatchUpload(nil, []*Event{{}})

		if err == nil {
			t.Fatal("Error is nil;\n Expected value")
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		client, _, teardown := setup()
		defer teardown()

		ctx, cancel := context.WithTimeout(context.TODO(), 0)
		defer cancel()
		_, err := client.Events.BatchUpload(ctx, []*Event{{}})

		if err == nil {
			t.Fatal("Error is nil;\n Expected value")
		}

		if got, want := err, ctx.Err(); got != want {
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

		resp, err := client.Events.BatchUpload(context.TODO(), events)

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
