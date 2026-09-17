package sportslot

import (
	"encoding/json"
	"mime/multipart"
	"sort"
	"strconv"
	"sync"
	"time"

	"back/infrastructure/storage"
)

type CategoriaStore interface {
	Create(c *Categoria) error
	FindByID(id int64) (*Categoria, error)
	FindAll() ([]Categoria, error)
	Update(id int64, c *Categoria) error
	Delete(id int64) error
}

type AmbienteStore interface {
	Create(a *Ambiente) error
	FindByID(id int64) (*Ambiente, error)
	FindAll() ([]Ambiente, error)
	Update(id int64, a *Ambiente) error
	Delete(id int64) error
}

type AmbienteInfoStore interface {
	Upsert(info *AmbienteInfo) error
	FindByAmbienteID(id int64) (*AmbienteInfo, error)
	DeleteByAmbienteID(id int64) error
}

type ImagenStore interface {
	AddFiles(ambienteID int64, archivos []*multipart.FileHeader, store *storage.Storage, allowedExts map[string]bool) ([]Imagen, error)
	ListByAmbienteID(ambienteID int64) ([]Imagen, error)
	DeleteByID(id int64) (*Imagen, error)
}

type OcupacionStore interface {
	Create(o *Ocupacion) error
	Update(id int64, o *Ocupacion) error
	Delete(id int64) error
	FindByID(id int64) (*Ocupacion, error)
	ListByAmbienteIDRange(ambienteID int64, from, to time.Time) ([]Ocupacion, error)
	ListRange(from, to time.Time) ([]Ocupacion, error)
}

type MockCategoriaStore struct {
	mu   sync.RWMutex
	seq  int64
	data map[int64]*Categoria
}

func NewMockCategoriaStore() *MockCategoriaStore {
	return &MockCategoriaStore{data: make(map[int64]*Categoria)}
}

func (s *MockCategoriaStore) Create(c *Categoria) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	c.ID = s.seq
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	s.data[c.ID] = cloneCategoria(c)
	return nil
}

func (s *MockCategoriaStore) FindByID(id int64) (*Categoria, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneCategoria(v), nil
}

func (s *MockCategoriaStore) FindAll() ([]Categoria, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Categoria, 0, len(s.data))
	for _, v := range s.data {
		out = append(out, *cloneCategoria(v))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (s *MockCategoriaStore) Update(id int64, c *Categoria) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return ErrNotFound
	}
	c.ID = id
	c.UpdatedAt = time.Now()
	s.data[id] = cloneCategoria(c)
	return nil
}

func (s *MockCategoriaStore) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return ErrNotFound
	}
	delete(s.data, id)
	return nil
}

func cloneCategoria(c *Categoria) *Categoria {
	if c == nil {
		return nil
	}
	cc := *c
	if c.Color != nil {
		v := *c.Color
		cc.Color = &v
	}
	return &cc
}

type MockAmbienteStore struct {
	mu   sync.RWMutex
	seq  int64
	data map[int64]*Ambiente
}

func NewMockAmbienteStore() *MockAmbienteStore {
	return &MockAmbienteStore{data: make(map[int64]*Ambiente)}
}

func (s *MockAmbienteStore) Create(a *Ambiente) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	a.ID = s.seq
	a.CreatedAt = time.Now()
	a.UpdatedAt = a.CreatedAt
	s.data[a.ID] = cloneAmbiente(a)
	return nil
}

func (s *MockAmbienteStore) FindByID(id int64) (*Ambiente, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneAmbiente(v), nil
}

func (s *MockAmbienteStore) FindAll() ([]Ambiente, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Ambiente, 0, len(s.data))
	for _, v := range s.data {
		out = append(out, *cloneAmbiente(v))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (s *MockAmbienteStore) Update(id int64, a *Ambiente) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return ErrNotFound
	}
	a.ID = id
	a.UpdatedAt = time.Now()
	s.data[id] = cloneAmbiente(a)
	return nil
}

func (s *MockAmbienteStore) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return ErrNotFound
	}
	delete(s.data, id)
	return nil
}

func cloneAmbiente(a *Ambiente) *Ambiente {
	if a == nil {
		return nil
	}
	aa := *a
	return &aa
}

type MockAmbienteInfoStore struct {
	mu   sync.RWMutex
	data map[int64]*AmbienteInfo
}

func NewMockAmbienteInfoStore() *MockAmbienteInfoStore {
	return &MockAmbienteInfoStore{data: make(map[int64]*AmbienteInfo)}
}

func (s *MockAmbienteInfoStore) Upsert(info *AmbienteInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if existing, ok := s.data[info.AmbienteID]; ok {
		info.CreatedAt = existing.CreatedAt
		info.CreatedAt = existing.CreatedAt
		info.UpdatedAt = now
	} else {
		if info.Data == nil {
			info.Data = json.RawMessage(`{}`)
		}
		info.UpdatedAt = now
		info.CreatedAt = now
	}
	s.data[info.AmbienteID] = cloneAmbienteInfo(info)
	return nil
}

func (s *MockAmbienteInfoStore) FindByAmbienteID(id int64) (*AmbienteInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneAmbienteInfo(v), nil
}

func (s *MockAmbienteInfoStore) DeleteByAmbienteID(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return ErrNotFound
	}
	delete(s.data, id)
	return nil
}

func cloneAmbienteInfo(i *AmbienteInfo) *AmbienteInfo {
	if i == nil {
		return nil
	}
	ci := *i
	if i.Data != nil {
		ci.Data = append([]byte(nil), i.Data...)
	}
	return &ci
}

type MockImagenStore struct {
	mu   sync.RWMutex
	seq  int64
	data map[int64]*Imagen
}

func NewMockImagenStore() *MockImagenStore {
	return &MockImagenStore{data: make(map[int64]*Imagen)}
}

func (s *MockImagenStore) AddFiles(ambienteID int64, archivos []*multipart.FileHeader, store *storage.Storage, allowedExts map[string]bool) ([]Imagen, error) {
	filenames, err := store.SaveFiles(archivos, fmtPrefix("ambiente", ambienteID), allowedExts)
	if err != nil {
		return nil, err
	}
	created := make([]Imagen, 0, len(filenames))
	s.mu.Lock()
	defer s.mu.Unlock()
	for idx, fn := range filenames {
		s.seq++
		img := &Imagen{ID: s.seq, AmbienteID: ambienteID, Archivo: fn, Orden: idx, CreatedAt: time.Now(), UpdatedAt: time.Now(), Activo: true}
		s.data[img.ID] = cloneImagen(img)
		created = append(created, *cloneImagen(img))
	}
	return created, nil
}

func fmtPrefix(tipo string, id int64) string { return tipo + "_" + strconv.FormatInt(id, 10) }

func (s *MockImagenStore) ListByAmbienteID(ambienteID int64) ([]Imagen, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Imagen
	for _, v := range s.data {
		if v.AmbienteID == ambienteID && v.Activo {
			out = append(out, *cloneImagen(v))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Orden != out[j].Orden {
			return out[i].Orden < out[j].Orden
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (s *MockImagenStore) DeleteByID(id int64) (*Imagen, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	cpy := cloneImagen(v)
	delete(s.data, id)
	return cpy, nil
}

func cloneImagen(i *Imagen) *Imagen {
	if i == nil {
		return nil
	}
	ci := *i
	if i.Descripcion != nil {
		v := *i.Descripcion
		ci.Descripcion = &v
	}
	if i.MimeType != nil {
		v := *i.MimeType
		ci.MimeType = &v
	}
	if i.Bytes != nil {
		v := *i.Bytes
		ci.Bytes = &v
	}
	return &ci
}

type MockOcupacionStore struct {
	mu   sync.RWMutex
	seq  int64
	data map[int64]*Ocupacion
}

func NewMockOcupacionStore() *MockOcupacionStore {
	return &MockOcupacionStore{data: make(map[int64]*Ocupacion)}
}

func (s *MockOcupacionStore) Create(o *Ocupacion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !o.Inicio.Before(o.Fin) {
		return ErrInvalidTimeRange
	}
	if err := s.checkOverlapLocked(o.AmbienteID, 0, o.Inicio, o.Fin, o.Activo); err != nil {
		return err
	}
	s.seq++
	o.ID = s.seq
	o.CreatedBy = nil
	now := time.Now()
	o.CreatedBy = nil
	o.UpdatedAt = now
	o2 := cloneOcupacion(o)
	o2.CreatedBy = o.CreatedBy
	s.data[o.ID] = o2
	return nil
}

func (s *MockOcupacionStore) Update(id int64, o *Ocupacion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.data[id]
	if !ok {
		return ErrNotFound
	}
	if !o.Inicio.Before(o.Fin) {
		return ErrInvalidTimeRange
	}
	if err := s.checkOverlapLocked(o.AmbienteID, id, o.Inicio, o.Fin, o.Activo); err != nil {
		return err
	}
	o.ID = id
	o.CreatedBy = existing.CreatedBy
	o.UpdatedAt = time.Now()
	s.data[id] = cloneOcupacion(o)
	return nil
}

func (s *MockOcupacionStore) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return ErrNotFound
	}
	delete(s.data, id)
	return nil
}

func (s *MockOcupacionStore) FindByID(id int64) (*Ocupacion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneOcupacion(v), nil
}

func (s *MockOcupacionStore) ListByAmbienteIDRange(ambienteID int64, from, to time.Time) ([]Ocupacion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Ocupacion
	for _, v := range s.data {
		if v.AmbienteID != ambienteID {
			continue
		}
		if v.Activo {
			// half-open [inicio,fin)
			if v.Inicio.Before(to) && v.Fin.After(from) {
				out = append(out, *cloneOcupacion(v))
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Inicio.Before(out[j].Inicio) })
	return out, nil
}

func (s *MockOcupacionStore) ListRange(from, to time.Time) ([]Ocupacion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Ocupacion
	for _, v := range s.data {
		if !v.Activo {
			continue
		}
		if v.Inicio.Before(to) && v.Fin.After(from) {
			out = append(out, *cloneOcupacion(v))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Inicio.Before(out[j].Inicio) })
	return out, nil
}

func (s *MockOcupacionStore) checkOverlapLocked(ambienteID int64, ignoreID int64, from, to time.Time, activo bool) error {
	if !activo {
		return nil
	}
	for id, v := range s.data {
		if id == ignoreID {
			continue
		}
		if v.AmbienteID != ambienteID || !v.Activo {
			continue
		}
		if v.Inicio.Before(to) && v.Fin.After(from) {
			return ErrOverlap
		}
	}
	return nil
}

func cloneOcupacion(o *Ocupacion) *Ocupacion {
	if o == nil {
		return nil
	}
	co := *o
	if o.CreatedBy != nil {
		v := *o.CreatedBy
		co.CreatedBy = &v
	}
	return &co
}
