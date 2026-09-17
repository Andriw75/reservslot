package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"back/services/sportslot"
)

type SportSlotSSEHandler struct {
	Service *sportslot.ServicioSportSlot
	Bus     *sportslot.SSEBus
	counter atomic.Int64
}

func NewSportSlotSSEHandler(service *sportslot.ServicioSportSlot, bus *sportslot.SSEBus) *SportSlotSSEHandler {
	return &SportSlotSSEHandler{Service: service, Bus: bus}
}

// Events — GET /events?topics=categorias,ambientes,ambiente:1:info,ambiente:1:ocupaciones
//
// Un único stream SSE multiplexado. El cliente declara los topics al conectar
// y el servidor envía un snapshot inicial por topic + eventos incrementales.
func (h *SportSlotSSEHandler) Events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	topics := parseTopics(r.URL.Query().Get("topics"))
	if len(topics) == 0 {
		topics = []string{"categorias", "ambientes"}
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	// Snapshot inicial por cada topic autorizado.
	for _, t := range topics {
		data := h.snapshot(t)
		h.write(w, flusher, "snapshot", map[string]any{"topic": t, "data": data})
	}

	ch := h.Bus.Register()
	defer h.Bus.Unregister(ch)

	ctxDone := r.Context().Done()
	heart := sportslot.NewHeartbeater(20 * time.Second)
	go heart.Run(w, flusher, ctxDone)

	for {
		select {
		case <-ctxDone:
			return
		case evt, open := <-ch:
			if !open {
				return
			}
			if !containsTopic(topics, string(evt.Topic)) {
				continue
			}
			h.write(w, flusher, evt.Event, map[string]any{
				"topic": string(evt.Topic),
				"data":  evt.Data,
			})
		}
	}
}

func (h *SportSlotSSEHandler) write(w http.ResponseWriter, flusher http.Flusher, event string, data any) {
	id := h.counter.Add(1)
	_ = sportslot.WriteSSE(w, flusher, id, event, data)
}

func (h *SportSlotSSEHandler) snapshot(topic string) any {
	switch {
	case topic == "categorias":
		items, _ := h.Service.GetCategoriasPublic()
		return items
	case topic == "ambientes":
		items, _ := h.Service.GetAmbientesPublic()
		return items
	case strings.HasPrefix(topic, "ambiente:"):
		ambID, ok := ambienteIDFromTopic(topic)
		if !ok {
			return nil
		}
		if strings.HasSuffix(topic, ":info") {
			info, err := h.Service.GetAmbienteInfo(ambID)
			if err != nil || info == nil {
				info = &sportslot.AmbienteInfo{AmbienteID: ambID, Data: []byte(`{}`), Activo: true}
			}
			imgs, _ := h.Service.GetAmbienteImagenes(ambID)
			return map[string]any{"info": info, "imagenes": imgs}
		}
		if strings.HasSuffix(topic, ":ocupaciones") {
			tr := sportslot.WeekRangeAmericaLima(time.Now())
			items, _ := h.Service.GetOcupaciones(ambID, tr.From, tr.To)
			return map[string]any{"from": tr.From, "to": tr.To, "data": items}
		}
	}
	return nil
}

func parseTopics(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" || !validTopic(p) {
			continue
		}
		out = append(out, p)
	}
	return out
}

func validTopic(t string) bool {
	switch t {
	case "categorias", "ambientes":
		return true
	}
	if !strings.HasPrefix(t, "ambiente:") {
		return false
	}
	if !strings.HasSuffix(t, ":info") && !strings.HasSuffix(t, ":ocupaciones") {
		return false
	}
	_, ok := ambienteIDFromTopic(t)
	return ok
}

func ambienteIDFromTopic(t string) (int64, bool) {
	parts := strings.Split(t, ":")
	if len(parts) < 3 || parts[0] != "ambiente" {
		return 0, false
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func containsTopic(topics []string, t string) bool {
	for _, x := range topics {
		if x == t {
			return true
		}
	}
	return false
}
