package postgres

import (
	"context"
	"time"

	"back/domain"
	"back/infrastructure/auth"
)

// UsuarioRepo implementa el acceso a usuarios.
//
// Los permisos se leen de la tabla `permisos` y se adjuntan al usuario: de
// momento no se enlazan por `usuario_permisos` ni restringen operaciones, pero
// el JWT ya viaja con el catálogo completo para el frontend.
type UsuarioRepo struct {
	db    *DB
	crypt *auth.CryptManager
}

func NewUsuarioRepo(db *DB, crypt *auth.CryptManager) *UsuarioRepo {
	return &UsuarioRepo{db: db, crypt: crypt}
}

const usuarioCols = `id, username, password_hash, activo, created_at, updated_at`

func (r *UsuarioRepo) scan(row interface {
	Scan(dest ...any) error
}) (*domain.User, error) {
	var u domain.User
	if err := row.Scan(&u.ID, &u.Username, &u.Password, &u.Activo, &u.CreadoEn, &u.ActualizadoEn); err != nil {
		return nil, mapErr(err)
	}
	return &u, nil
}

func (r *UsuarioRepo) FindByID(id int64) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u, err := r.scan(r.db.Pool.QueryRow(ctx, `SELECT `+usuarioCols+` FROM usuarios WHERE id = $1`, id))
	if err != nil {
		return nil, err
	}
	u.Permisos = r.allPermissionCodes(ctx)
	return u, nil
}

func (r *UsuarioRepo) FindByUsername(username string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u, err := r.scan(r.db.Pool.QueryRow(ctx,
		`SELECT `+usuarioCols+` FROM usuarios WHERE lower(username) = lower($1)`, username))
	if err != nil {
		return nil, err
	}
	u.Permisos = r.allPermissionCodes(ctx)
	return u, nil
}

func (r *UsuarioRepo) FindAll() []domain.User {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.db.Pool.Query(ctx, `SELECT `+usuarioCols+` FROM usuarios ORDER BY id`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	permisos := r.allPermissionCodes(ctx)
	out := []domain.User{}
	for rows.Next() {
		u, err := r.scan(rows)
		if err != nil {
			continue
		}
		u.Permisos = permisos
		out = append(out, *u)
	}
	return out
}

func (r *UsuarioRepo) IsValidPassword(username, password string) (*domain.User, bool) {
	user, err := r.FindByUsername(username)
	if err != nil || user == nil || !user.Activo {
		return nil, false
	}
	if !r.crypt.Verify(user.Password, password) {
		return nil, false
	}
	return user, true
}

func (r *UsuarioRepo) allPermissionCodes(ctx context.Context) []string {
	rows, err := r.db.Pool.Query(ctx, `SELECT codigo FROM permisos ORDER BY codigo`)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err == nil {
			out = append(out, code)
		}
	}
	return out
}
