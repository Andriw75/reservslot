package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"back/services/sportslot"
)

const maxImagenBytes = 10 << 20 // 10 MB por archivo
const maxImagenesPorRequest = 10

var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".gif":  true,
}

type SportSlotHandlers struct {
	Service *sportslot.ServicioSportSlot
	SSEBus  *sportslot.SSEBus
}

func NewSportSlotHandlers(service *sportslot.ServicioSportSlot, bus *sportslot.SSEBus) *SportSlotHandlers {
	return &SportSlotHandlers{Service: service, SSEBus: bus}
}

type CategoriaInput struct {
	Nombre      string  `json:"nombre"`
	Descripcion string  `json:"descripcion"`
	Color       *string `json:"color,omitempty"`
	Activo      *bool   `json:"activo,omitempty"`
}

type AmbienteInput struct {
	Nombre      string `json:"nombre"`
	CategoriaID int64  `json:"categoria_id"`
	Activo      *bool  `json:"activo,omitempty"`
}

type AmbienteInfoInput struct {
	Data   json.RawMessage `json:"data"`
	Activo *bool           `json:"activo,omitempty"`
}

type OcupacionInput struct {
	AmbienteID int64     `json:"ambiente_id"`
	Inicio     time.Time `json:"inicio"`
	Fin        time.Time `json:"fin"`
	Activo     *bool     `json:"activo,omitempty"`
}

func boolOr(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

func topicAmbienteInfo(id int64) sportslot.SportSlotTopic {
	return sportslot.SportSlotTopic("ambiente:" + strconv.FormatInt(id, 10) + ":info")
}

func topicAmbienteOcupaciones(id int64) sportslot.SportSlotTopic {
	return sportslot.SportSlotTopic("ambiente:" + strconv.FormatInt(id, 10) + ":ocupaciones")
}

// ---------------------------------------------------------------- Categorías

func (h *SportSlotHandlers) GetCategorias(w http.ResponseWriter, r *http.Request) {
	items, err := h.Service.GetCategoriasPublic()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *SportSlotHandlers) GetCategoria(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.Service.Categorias.FindByID(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *SportSlotHandlers) PostCategoria(w http.ResponseWriter, r *http.Request) {
	var in CategoriaInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if in.Nombre == "" {
		writeJSONError(w, http.StatusBadRequest, "nombre is required")
		return
	}
	item := &sportslot.Categoria{
		Nombre:      in.Nombre,
		Descripcion: in.Descripcion,
		Color:       in.Color,
		Activo:      boolOr(in.Activo, true),
	}
	if err := h.Service.Categorias.Create(item); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create")
		return
	}
	h.SSEBus.Broadcast(sportslot.SportSlotSSE{Topic: sportslot.TopicCategorias, Event: "categoria.upsert", Data: item})
	writeJSON(w, http.StatusCreated, item)
}

func (h *SportSlotHandlers) PutPatchCategoria(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in map[string]any
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	existing, err := h.Service.Categorias.FindByID(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	if v, ok := in["nombre"].(string); ok {
		existing.Nombre = v
	}
	if v, ok := in["descripcion"].(string); ok {
		existing.Descripcion = v
	}
	if v, ok := in["color"]; ok {
		if v == nil {
			existing.Color = nil
		} else if s, ok := v.(string); ok {
			if s == "" {
				existing.Color = nil
			} else {
				existing.Color = &s
			}
		}
	}
	if v, ok := in["activo"].(bool); ok {
		existing.Activo = v
	}
	if err := h.Service.Categorias.Update(id, existing); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to update")
		return
	}
	h.SSEBus.Broadcast(sportslot.SportSlotSSE{Topic: sportslot.TopicCategorias, Event: "categoria.upsert", Data: existing})
	writeJSON(w, http.StatusOK, existing)
}

func (h *SportSlotHandlers) DeleteCategoria(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	ambientes, _ := h.Service.Ambientes.FindAll()
	for _, a := range ambientes {
		if a.CategoriaID == id {
			writeJSONError(w, http.StatusConflict, "categoria con ambientes asociados")
			return
		}
	}
	if err := h.Service.Categorias.Delete(id); err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	h.SSEBus.Broadcast(sportslot.SportSlotSSE{Topic: sportslot.TopicCategorias, Event: "categoria.delete", Data: map[string]int64{"id": id}})
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// ---------------------------------------------------------------- Ambientes

func (h *SportSlotHandlers) GetAmbientes(w http.ResponseWriter, r *http.Request) {
	items, err := h.Service.GetAmbientesPublic()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *SportSlotHandlers) GetAmbiente(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.Service.Ambientes.FindByID(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *SportSlotHandlers) GetAmbienteInfo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	info, err := h.Service.GetAmbienteInfo(id)
	if err != nil {
		if errors.Is(err, sportslot.ErrNotFound) {
			info = &sportslot.AmbienteInfo{AmbienteID: id, Data: json.RawMessage(`{}`), Activo: true}
		} else {
			writeJSONError(w, http.StatusInternalServerError, "failed")
			return
		}
	}
	imagenes, err := h.Service.GetAmbienteImagenes(id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"info": info, "imagenes": imagenes})
}

func (h *SportSlotHandlers) PostAmbiente(w http.ResponseWriter, r *http.Request) {
	var in AmbienteInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if in.Nombre == "" {
		writeJSONError(w, http.StatusBadRequest, "nombre is required")
		return
	}
	if in.CategoriaID == 0 {
		writeJSONError(w, http.StatusBadRequest, "categoria_id is required")
		return
	}
	if _, err := h.Service.Categorias.FindByID(in.CategoriaID); err != nil {
		writeJSONError(w, http.StatusBadRequest, "categoria not found")
		return
	}
	item := &sportslot.Ambiente{Nombre: in.Nombre, CategoriaID: in.CategoriaID, Activo: boolOr(in.Activo, true)}
	if err := h.Service.Ambientes.Create(item); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed")
		return
	}
	h.SSEBus.Broadcast(sportslot.SportSlotSSE{Topic: sportslot.TopicAmbientes, Event: "ambiente.upsert", Data: item})
	writeJSON(w, http.StatusCreated, item)
}

func (h *SportSlotHandlers) UpsertAmbiente(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in AmbienteInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	existing, err := h.Service.Ambientes.FindByID(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	if in.Nombre != "" {
		existing.Nombre = in.Nombre
	}
	if in.CategoriaID != 0 {
		if _, err := h.Service.Categorias.FindByID(in.CategoriaID); err != nil {
			writeJSONError(w, http.StatusBadRequest, "categoria not found")
			return
		}
		existing.CategoriaID = in.CategoriaID
	}
	if in.Activo != nil {
		existing.Activo = *in.Activo
	}
	if err := h.Service.Ambientes.Update(id, existing); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed")
		return
	}
	h.SSEBus.Broadcast(sportslot.SportSlotSSE{Topic: sportslot.TopicAmbientes, Event: "ambiente.upsert", Data: existing})
	writeJSON(w, http.StatusOK, existing)
}

func (h *SportSlotHandlers) DeleteAmbiente(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if _, err := h.Service.Ambientes.FindByID(id); err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}

	cascade := r.URL.Query().Get("cascade") == "true"

	total, err := h.Service.CountOcupaciones(id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed")
		return
	}
	if total > 0 && !cascade {
		writeJSON(w, http.StatusConflict, map[string]any{
			"error":       "El ambiente tiene ocupaciones registradas",
			"ocupaciones": total,
		})
		return
	}

	imagenes, _ := h.Service.GetAmbienteImagenes(id)
	if err := h.Service.DeleteAmbiente(id, cascade); err != nil {
		if errors.Is(err, sportslot.ErrHasOcupaciones) {
			writeJSON(w, http.StatusConflict, map[string]any{
				"error":       "El ambiente tiene ocupaciones registradas",
				"ocupaciones": total,
			})
			return
		}
		if errors.Is(err, sportslot.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed")
		return
	}

	_ = h.Service.AmbienteInfo.DeleteByAmbienteID(id)
	names := make([]string, 0, len(imagenes))
	for _, img := range imagenes {
		names = append(names, img.Archivo)
		h.Service.Imagenes.DeleteByID(img.ID)
	}
	h.Service.Storage.RemoveImages(names...)

	h.SSEBus.Broadcast(sportslot.SportSlotSSE{Topic: sportslot.TopicAmbientes, Event: "ambiente.delete", Data: map[string]int64{"id": id}})
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *SportSlotHandlers) UpsertAmbienteInfo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if _, err := h.Service.Ambientes.FindByID(id); err != nil {
		writeJSONError(w, http.StatusNotFound, "ambiente not found")
		return
	}
	var in AmbienteInfoInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	info, err := h.Service.UpsertAmbienteInfo(id, in.Data, boolOr(in.Activo, true))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	imagenes, err := h.Service.GetAmbienteImagenes(id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed")
		return
	}
	h.broadcastAmbienteInfo(id)
	writeJSON(w, http.StatusOK, map[string]any{"info": info, "imagenes": imagenes})
}

// ---------------------------------------------------------------- Imágenes

func (h *SportSlotHandlers) UploadAmbienteImagenes(w http.ResponseWriter, r *http.Request) {
	ambienteID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if _, err := h.Service.Ambientes.FindByID(ambienteID); err != nil {
		writeJSONError(w, http.StatusNotFound, "ambiente not found")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxImagenBytes*maxImagenesPorRequest)
	if err := r.ParseMultipartForm(maxImagenBytes * maxImagenesPorRequest); err != nil {
		writeJSONError(w, http.StatusBadRequest, "formulario inválido o demasiado grande")
		return
	}

	files := r.MultipartForm.File["images"]
	if len(files) == 0 {
		files = r.MultipartForm.File["archivos"]
	}
	if len(files) == 0 {
		writeJSONError(w, http.StatusBadRequest, "se requiere al menos un archivo (campo 'images')")
		return
	}
	for _, fh := range files {
		if fh.Size > maxImagenBytes {
			writeJSONError(w, http.StatusRequestEntityTooLarge, "cada imagen debe pesar como máximo 10 MB")
			return
		}
	}

	created, err := h.Service.SaveImagenes(ambienteID, files, allowedImageExts)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.broadcastAmbienteInfo(ambienteID)
	writeJSON(w, http.StatusCreated, created)
}

func (h *SportSlotHandlers) DeleteAmbienteImagen(w http.ResponseWriter, r *http.Request) {
	ambienteID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	imgID, err := strconv.ParseInt(chi.URLParam(r, "imagenID"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid imagenID")
		return
	}
	img, err := h.Service.Imagenes.DeleteByID(imgID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	h.Service.Storage.RemoveImages(img.Archivo)
	h.broadcastAmbienteInfo(ambienteID)
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *SportSlotHandlers) ReorderAmbienteImagenes(w http.ResponseWriter, r *http.Request) {
	ambienteID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if _, err := h.Service.Ambientes.FindByID(ambienteID); err != nil {
		writeJSONError(w, http.StatusNotFound, "ambiente not found")
		return
	}

	var in struct {
		IDs []int64 `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(in.IDs) == 0 {
		writeJSONError(w, http.StatusBadRequest, "ids is required")
		return
	}

	imagenes, err := h.Service.ReorderImagenes(ambienteID, in.IDs)
	if err != nil {
		if errors.Is(err, sportslot.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "imagen not found for this ambiente")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed")
		return
	}
	h.broadcastAmbienteInfo(ambienteID)
	writeJSON(w, http.StatusOK, imagenes)
}

func (h *SportSlotHandlers) broadcastAmbienteInfo(ambienteID int64) {
	info, err := h.Service.GetAmbienteInfo(ambienteID)
	if err != nil || info == nil {
		return
	}
	imgs, _ := h.Service.GetAmbienteImagenes(ambienteID)
	h.SSEBus.Broadcast(sportslot.SportSlotSSE{
		Topic: topicAmbienteInfo(ambienteID),
		Event: "ambiente_info.upsert",
		Data:  map[string]any{"info": info, "imagenes": imgs},
	})
}

// ---------------------------------------------------------------- Ocupaciones

func (h *SportSlotHandlers) GetOcupaciones(w http.ResponseWriter, r *http.Request) {
	ambienteID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if _, err := h.Service.Ambientes.FindByID(ambienteID); err != nil {
		writeJSONError(w, http.StatusNotFound, "ambiente not found")
		return
	}

	q := r.URL.Query()
	fromStr := q.Get("from")
	toStr := q.Get("to")

	var from, to time.Time
	if fromStr == "" && toStr == "" {
		// Público: semana actual en America/Lima.
		tr := sportslot.WeekRangeAmericaLima(time.Now())
		from, to = tr.From, tr.To
	} else {
		if fromStr == "" || toStr == "" {
			writeJSONError(w, http.StatusBadRequest, "from y to son requeridos juntos")
			return
		}
		from, err = sportslot.ParseTimeInLima(fromStr)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "from inválido")
			return
		}
		to, err = sportslot.ParseTimeInLima(toStr)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "to inválido")
			return
		}
		if !from.Before(to) {
			writeJSONError(w, http.StatusBadRequest, "from debe ser menor que to")
			return
		}
	}

	items, err := h.Service.GetOcupaciones(ambienteID, from, to)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"from": from, "to": to, "data": items})
}

func (h *SportSlotHandlers) PostOcupacion(w http.ResponseWriter, r *http.Request) {
	var in OcupacionInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if in.AmbienteID == 0 {
		writeJSONError(w, http.StatusBadRequest, "ambiente_id required")
		return
	}
	if _, err := h.Service.Ambientes.FindByID(in.AmbienteID); err != nil {
		writeJSONError(w, http.StatusBadRequest, "ambiente not found")
		return
	}
	if !in.Inicio.Before(in.Fin) {
		writeJSONError(w, http.StatusBadRequest, "inicio debe ser menor que fin")
		return
	}
	o := &sportslot.Ocupacion{
		AmbienteID: in.AmbienteID,
		Inicio:     in.Inicio,
		Fin:        in.Fin,
		Activo:     boolOr(in.Activo, true),
	}
	if err := h.Service.Ocupaciones.Create(o); err != nil {
		h.writeOcupacionError(w, err)
		return
	}
	h.broadcastOcupacion(o, "ocupacion.upsert")
	writeJSON(w, http.StatusCreated, o)
}

func (h *SportSlotHandlers) PutOcupacion(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	before, err := h.Service.Ocupaciones.FindByID(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	var in OcupacionInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if in.AmbienteID == 0 {
		in.AmbienteID = before.AmbienteID
	}
	if in.Inicio.IsZero() {
		in.Inicio = before.Inicio
	}
	if in.Fin.IsZero() {
		in.Fin = before.Fin
	}
	if _, err := h.Service.Ambientes.FindByID(in.AmbienteID); err != nil {
		writeJSONError(w, http.StatusBadRequest, "ambiente not found")
		return
	}
	if !in.Inicio.Before(in.Fin) {
		writeJSONError(w, http.StatusBadRequest, "inicio debe ser menor que fin")
		return
	}
	o := &sportslot.Ocupacion{
		AmbienteID: in.AmbienteID,
		Inicio:     in.Inicio,
		Fin:        in.Fin,
		Activo:     boolOr(in.Activo, before.Activo),
	}
	if err := h.Service.Ocupaciones.Update(id, o); err != nil {
		h.writeOcupacionError(w, err)
		return
	}
	if before.AmbienteID != o.AmbienteID {
		h.broadcastOcupacion(before, "ocupacion.delete")
	}
	h.broadcastOcupacion(o, "ocupacion.upsert")
	writeJSON(w, http.StatusOK, o)
}

func (h *SportSlotHandlers) DeleteOcupacion(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	before, err := h.Service.Ocupaciones.FindByID(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	if err := h.Service.Ocupaciones.Delete(id); err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	h.broadcastOcupacion(before, "ocupacion.delete")
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *SportSlotHandlers) writeOcupacionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, sportslot.ErrOverlap):
		writeJSONError(w, http.StatusConflict, "el horario se solapa con otra ocupación")
	case errors.Is(err, sportslot.ErrInvalidTimeRange):
		writeJSONError(w, http.StatusBadRequest, "rango de fechas inválido")
	case errors.Is(err, sportslot.ErrNotFound):
		writeJSONError(w, http.StatusNotFound, "not found")
	default:
		writeJSONError(w, http.StatusInternalServerError, "failed")
	}
}

// broadcastOcupacion emite el evento solo si la ocupación interseca la semana
// actual (el frontend público únicamente muestra la semana en curso).
func (h *SportSlotHandlers) broadcastOcupacion(o *sportslot.Ocupacion, event string) {
	if o == nil {
		return
	}
	tr := sportslot.WeekRangeAmericaLima(time.Now())
	if !o.Inicio.Before(tr.To) || !o.Fin.After(tr.From) {
		return
	}
	h.SSEBus.Broadcast(sportslot.SportSlotSSE{
		Topic: topicAmbienteOcupaciones(o.AmbienteID),
		Event: event,
		Data:  o,
	})
}
