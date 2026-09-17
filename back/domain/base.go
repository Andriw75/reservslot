package domain

import "time"

// User es el usuario del sistema SportSlot.
// Los permisos se exponen en el JWT, pero de momento no restringen operaciones.
type User struct {
	ID            int64     `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	Password      string    `json:"-"`
	Activo        bool      `json:"activo"`
	Permisos      []string  `json:"permisos"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}

func (u *User) SetID(id int64) { u.ID = id }

// Permission representa un permiso del catálogo (tabla permisos).
// De momento solo se listan: no se enlazan a usuarios ni se validan.
type Permission struct {
	Codigo      string `json:"codigo"`
	Descripcion string `json:"descripcion"`
}

// UserPermission representa la relación usuario <-> permiso (tabla usuario_permisos).
type UserPermission struct {
	UserID        int64  `json:"usuario_id"`
	PermisoCodigo string `json:"permiso_codigo"`
	Permiso       string `json:"permiso"`
}
