package sportslot

import (
	"encoding/json"
	"errors"
	"time"
)

type Categoria struct {
	ID          int64     `json:"id"`
	Nombre      string    `json:"nombre"`
	Descripcion string    `json:"descripcion"`
	Color       *string   `json:"color,omitempty"`
	Activo      bool      `json:"activo"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Ambiente struct {
	ID         int64     `json:"id"`
	Nombre     string    `json:"nombre"`
	CategoriaID int64   `json:"categoria_id"`
	Activo     bool      `json:"activo"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type AmbienteInfo struct {
	AmbienteID int64           `json:"ambiente_id"`
	Data       json.RawMessage `json:"data"`
	Activo     bool            `json:"activo"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type Imagen struct {
	ID          int64     `json:"id"`
	AmbienteID int64     `json:"ambiente_id"`
	Archivo     string    `json:"archivo"`
	Descripcion *string   `json:"descripcion,omitempty"`
	Orden       int       `json:"orden"`
	MimeType    *string   `json:"mime_type,omitempty"`
	Bytes       *int64    `json:"bytes,omitempty"`
	Activo      bool      `json:"activo"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Ocupacion struct {
	ID         int64     `json:"id"`
	AmbienteID int64     `json:"ambiente_id"`
	Inicio     time.Time `json:"inicio"`
	Fin        time.Time `json:"fin"`
	Activo     bool      `json:"activo"`
	CreatedBy  *int64    `json:"created_by,omitempty"`
	UpdatedAt  time.Time `json:"updated_at"`
}

var (
	ErrNotFound          = errors.New("not found")
	ErrOverlap          = errors.New("overlapping occupation")
	ErrInvalidTimeRange = errors.New("invalid time range")
)
