package sportslot

import (
	"mime/multipart"
	"time"

	"back/infrastructure/storage"
)

// Contratos de persistencia. Las implementaciones reales viven en
// back/infrastructure/postgres.

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
	UpdateOrden(ambienteID int64, ids []int64) error
}

type OcupacionStore interface {
	Create(o *Ocupacion) error
	Update(id int64, o *Ocupacion) error
	Delete(id int64) error
	FindByID(id int64) (*Ocupacion, error)
	ListByAmbienteIDRange(ambienteID int64, from, to time.Time) ([]Ocupacion, error)
	ListRange(from, to time.Time) ([]Ocupacion, error)
}
