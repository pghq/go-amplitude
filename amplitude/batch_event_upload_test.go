// XXX: Add comments
package amplitude

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestBatchEventsSuccessSummary_String(t *testing.T) {
	t.Run("no summary", func(t *testing.T) {
		s := BatchEventsSuccessSummary{}
		want := "0 events ingested (0 bytes) at 1970-01-01 00:00:00 +0000 UTC"

		if s.String() != want {
			t.Errorf("Summary = %s; expected %s", s.String(), want)
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
			t.Errorf("Summary = %s; expected %s", s.String(), want)
		}
	})
}

func TestBatchEventUploadService_Send(t *testing.T) {
	t.Run("no events", func(t *testing.T) {
		client, _, teardown := setup()
		defer teardown()

		_, err := client.BatchEventUpload.Send(context.TODO(), nil)

		if err == nil {
			t.Fatal("Error is nil; expected value")
		}

		if got, want := err.Error(), "no events to send"; got != want {
			t.Errorf("Error = %s; expected %s", got, want)
		}
	})

	t.Run("no context", func(t *testing.T) {
		client, _, teardown := setup()
		defer teardown()

		_, err := client.BatchEventUpload.Send(nil, []Event{{}})

		if err == nil {
			t.Fatal("Error is nil; expected value")
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		client, _, teardown := setup()
		defer teardown()

		ctx, cancel := context.WithTimeout(context.TODO(), 0)
		defer cancel()
		_, err := client.BatchEventUpload.Send(ctx, []Event{{}})

		if err == nil {
			t.Fatal("Error is nil; expected value")
		}

		if got, want := err, ctx.Err(); got != want {
			t.Errorf("Error = %v; expected %v", got, want)
		}
	})

	t.Run("200 ok", func(t *testing.T) {
		client, mux, teardown := setup()
		defer teardown()

		var event Event
		data, _ := ioutil.ReadFile("testdata/event.json")
		_ = json.Unmarshal(data, &event)
		events := []Event{event}

		mux.HandleFunc(batchEventUploadEndpoint, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Method = %s; expected %s", r.Method, http.MethodPost)
			}

			data, _ := json.Marshal(events)
			want := fmt.Sprintf("{\"events\":%s}\n", string(data))

			if got, _ := io.ReadAll(r.Body); string(got) != want {
				t.Errorf("Body = %s; expected %s", got, want)
			}

			_, _ = fmt.Fprint(w, `{
			  "code": 200,
			  "events_ingested": 50,
			  "payload_size_bytes": 50,
			  "server_upload_time": 1396381378123
			}`)
		})

		resp, err := client.BatchEventUpload.Send(context.TODO(), events)

		if err != nil {
			t.Fatalf("Error = %v; expected no error", err)
		}

		if got, want := resp.Code, 200; got != want {
			t.Errorf("Code = %d; expected %d", got, want)
		}

		if got, want := resp.EventsIngested, 50; got != want {
			t.Errorf("EventsIngested = %d; expected %d", got, want)
		}

		if got, want := resp.PayloadSize, 50; got != want {
			t.Errorf("PayloadSize = %d; expected %d", got, want)
		}

		if got, want := resp.UploadTime, int64(1396381378123); got != want {
			t.Errorf("UploadTime = %d; expected %d", got, want)
		}
	})
}
