package postgres

import (
	"context"
	"time"

	"back/services/sportslot"
)

type CategoriaRepo struct {
	db *DB
}

func NewCategoriaRepo(db *DB) *CategoriaRepo { return &CategoriaRepo{db: db} }

const categoriaCols = `id, nombre, descripcion, color, activo, created_at, updated_at`

func scanCategoria(row interface {
	Scan(dest ...any) error
}) (*sportslot.Categoria, error) {
	var c sportslot.Categoria
	if err := row.Scan(&c.ID, &c.Nombre, &c.Descripcion, &c.Color, &c.Activo, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, mapErr(err)
	}
	return &c, nil
}

func (r *CategoriaRepo) Create(c *sportslot.Categoria) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.db.Pool.QueryRow(ctx,
		`INSERT INTO categorias (nombre, descripcion, color, activo)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+categoriaCols,
		c.Nombre, c.Descripcion, c.Color, c.Activo,
	).Scan(&c.ID, &c.Nombre, &c.Descripcion, &c.Color, &c.Activo, &c.CreatedAt, &c.UpdatedAt)
	return mapErr(err)
}

func (r *CategoriaRepo) FindByID(id int64) (*sportslot.Categoria, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := r.db.Pool.QueryRow(ctx, `SELECT `+categoriaCols+` FROM categorias WHERE id = $1`, id)
	return scanCategoria(row)
}

func (r *CategoriaRepo) FindAll() ([]sportslot.Categoria, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.db.Pool.Query(ctx, `SELECT `+categoriaCols+` FROM categorias ORDER BY id`)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()

	out := []sportslot.Categoria{}
	for rows.Next() {
		c, err := scanCategoria(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, mapErr(rows.Err())
}

func (r *CategoriaRepo) Update(id int64, c *sportslot.Categoria) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.db.Pool.QueryRow(ctx,
		`UPDATE categorias
		    SET nombre = $2, descripcion = $3, color = $4, activo = $5
		  WHERE id = $1
		  RETURNING `+categoriaCols,
		id, c.Nombre, c.Descripcion, c.Color, c.Activo,
	).Scan(&c.ID, &c.Nombre, &c.Descripcion, &c.Color, &c.Activo, &c.CreatedAt, &c.UpdatedAt)
	return mapErr(err)
}

func (r *CategoriaRepo) Delete(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tag, err := r.db.Pool.Exec(ctx, `DELETE FROM categorias WHERE id = $1`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return sportslot.ErrNotFound
	}
	return nil
}
