package amplitude

import (
	"context"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/diagnostic/log"
)

// Flush will send all batched events to the amplitude batch upload API
func (m *SendMiddleware) Flush(ctx context.Context) {
	var events []*Event
	for {
		select {
		case <-ctx.Done():
			m.SendError(errors.Wrap(ctx.Err()))
			return
		default:
			event := m.Event()
			if event != nil{
				events = append(events, event)
				continue
			}

			if len(events) == 0 {
				return
			}

			resp, err := m.client.Events.Send(ctx, events...)
			if err != nil {
				m.SendError(errors.Wrap(err))
				return
			}

			log.Infof("amplitude: total=%d, size=%d, time=%d events were flushed",
				resp.EventsIngested, resp.PayloadSize, resp.UploadTime)

			return
		}
	}
}
