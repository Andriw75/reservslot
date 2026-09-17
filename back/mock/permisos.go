package mock

import "back/domain"

// Catálogo de permisos del sistema.
// Coincide con la tabla `permisos` de schema.sql.
// De momento NO se enlazan a usuarios ni se validan en el backend:
// cualquier usuario autenticado puede operar sobre todos los módulos.
var permissionCodes = []domain.Permission{
	{Codigo: "usuarios:read", Descripcion: "Consultar usuarios"},
	{Codigo: "usuarios:write", Descripcion: "Crear, editar y eliminar usuarios"},
	{Codigo: "catalogo:write", Descripcion: "Administrar categorías, ambientes e información"},
	{Codigo: "archivos:write", Descripcion: "Subir y eliminar imágenes"},
	{Codigo: "ocupaciones:read", Descripcion: "Consultar ocupaciones"},
	{Codigo: "ocupaciones:write", Descripcion: "Crear, editar y borrar ocupaciones"},
}

// PermisoStore expone el catálogo de permisos (sin enlaces a usuarios).
type PermisoStore struct{}

func NewPermisoStore() *PermisoStore { return &PermisoStore{} }

func (s *PermisoStore) FindAll() []domain.Permission {
	out := make([]domain.Permission, len(permissionCodes))
	copy(out, permissionCodes)
	return out
}

// AllPermissionCodes devuelve todos los códigos de permiso existentes.
func AllPermissionCodes() []string {
	out := make([]string, 0, len(permissionCodes))
	for _, p := range permissionCodes {
		out = append(out, p.Codigo)
	}
	return out
}
