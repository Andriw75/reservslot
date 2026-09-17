package sportslot

import (
	"back/infrastructure/storage"
	"encoding/json"
	"mime/multipart"
	"sort"
	"time"
)

type ServicioSportSlot struct {
	Categorias   CategoriaStore
	Ambientes    AmbienteStore
	AmbienteInfo AmbienteInfoStore
	Imagenes     ImagenStore
	Ocupaciones  OcupacionStore
	Storage      *storage.Storage
	SSEBus       *SSEBus
}

type SportslotSnapshot struct {
	Categorias []Categoria `json:"categorias"`
	Ambientes  []Ambiente  `json:"ambientes"`
}

func (s *ServicioSportSlot) GetCategoriasPublic() ([]Categoria, error) {
	return s.Categorias.FindAll()
}

func (s *ServicioSportSlot) GetAmbientesPublic() ([]Ambiente, error) {
	return s.Ambientes.FindAll()
}

func (s *ServicioSportSlot) GetAmbienteInfo(ambienteID int64) (*AmbienteInfo, error) {
	return s.AmbienteInfo.FindByAmbienteID(ambienteID)
}

func (s *ServicioSportSlot) GetAmbienteImagenes(ambienteID int64) ([]Imagen, error) {
	return s.Imagenes.ListByAmbienteID(ambienteID)
}

func (s *ServicioSportSlot) GetOcupaciones(ambienteID int64, from, to time.Time) ([]Ocupacion, error) {
	return s.Ocupaciones.ListByAmbienteIDRange(ambienteID, from, to)
}

func (s *ServicioSportSlot) CountOcupaciones(ambienteID int64) (int, error) {
	return s.Ocupaciones.CountByAmbienteID(ambienteID)
}

// DeleteAmbiente elimina el ambiente. Si tiene ocupaciones y cascade es false
// devuelve ErrHasOcupaciones para que la capa HTTP pida confirmación explícita.
func (s *ServicioSportSlot) DeleteAmbiente(ambienteID int64, cascade bool) error {
	total, err := s.Ocupaciones.CountByAmbienteID(ambienteID)
	if err != nil {
		return err
	}
	if total > 0 && !cascade {
		return ErrHasOcupaciones
	}
	if cascade {
		return s.Ambientes.DeleteCascade(ambienteID)
	}
	return s.Ambientes.Delete(ambienteID)
}

func (s *ServicioSportSlot) SemanaActual(now time.Time) TimeRange {
	return WeekRangeAmericaLima(now)
}

type ImagenInput struct {
	Archivo     *multipart.FileHeader
	Descripcion *string
	MimeType    *string
	Bytes       *int64
}

func (s *ServicioSportSlot) UpsertAmbienteInfo(ambienteID int64, data []byte, activo bool) (*AmbienteInfo, error) {
	info := &AmbienteInfo{AmbienteID: ambienteID, Data: dataToRaw(data), Activo: activo}
	if err := s.AmbienteInfo.Upsert(info); err != nil {
		return nil, err
	}
	return s.AmbienteInfo.FindByAmbienteID(ambienteID)
}

func dataToRaw(b []byte) json.RawMessage {
	if b == nil {
		return json.RawMessage(`{}`)
	}
	return json.RawMessage(b)
}

func (s *ServicioSportSlot) SaveImagenes(ambienteID int64, archivos []*multipart.FileHeader, allowedExts map[string]bool) ([]Imagen, error) {
	return s.Imagenes.AddFiles(ambienteID, archivos, s.Storage, allowedExts)
}

func (s *ServicioSportSlot) ReorderImagenes(ambienteID int64, ids []int64) ([]Imagen, error) {
	if len(ids) == 0 {
		return nil, ErrInvalidTimeRange
	}
	if err := s.Imagenes.UpdateOrden(ambienteID, ids); err != nil {
		return nil, err
	}
	return s.Imagenes.ListByAmbienteID(ambienteID)
}

func sortByOrden[T any](items []T, less func(a, b T) bool) {
	sort.Slice(items, func(i, j int) bool { return less(items[i], items[j]) })
}
