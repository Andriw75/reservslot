package sportslot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WriteSSE escribe un evento SSE completo (id + event + data) y hace flush.
func WriteSSE(w http.ResponseWriter, flusher http.Flusher, id int64, eventName string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if id > 0 {
		_, _ = fmt.Fprintf(w, "id: %d\n", id)
	}
	if eventName != "" {
		_, _ = fmt.Fprintf(w, "event: %s\n", eventName)
	}
	_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
	flusher.Flush()
	return nil
}

type Heartbeater struct {
	interval time.Duration
}

func NewHeartbeater(interval time.Duration) *Heartbeater { return &Heartbeater{interval: interval} }

func (h *Heartbeater) Run(w http.ResponseWriter, flusher http.Flusher, ctxDone <-chan struct{}) {
	t := time.NewTicker(h.interval)
	defer t.Stop()
	for {
		select {
		case <-ctxDone:
			return
		case <-t.C:
			_, _ = fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}
