package postgres

import (
	"context"
	"time"

	"back/services/sportslot"
)

type AmbienteRepo struct {
	db *DB
}

func NewAmbienteRepo(db *DB) *AmbienteRepo { return &AmbienteRepo{db: db} }

const ambienteBaseCols = `id, nombre, categoria_id, activo, created_at, updated_at`
const ambienteJoinCols = `a.id, a.nombre, a.categoria_id, a.activo, a.created_at, a.updated_at, ip.archivo`

// ambienteSelect trae además la primera imagen activa (menor `orden`), que la
// portada pública usa como miniatura de cada ambiente.
const ambienteSelect = `
	SELECT ` + ambienteJoinCols + `
	  FROM ambientes a
	  LEFT JOIN LATERAL (
	       SELECT i.archivo
	         FROM imagenes i
	        WHERE i.ambiente_id = a.id AND i.activo = true
	        ORDER BY i.orden, i.id
	        LIMIT 1
	  ) ip ON true`

func scanAmbiente(row interface {
	Scan(dest ...any) error
}) (*sportslot.Ambiente, error) {
	var a sportslot.Ambiente
	if err := row.Scan(&a.ID, &a.Nombre, &a.CategoriaID, &a.Activo, &a.CreatedAt, &a.UpdatedAt, &a.ImagenPrincipal); err != nil {
		return nil, mapErr(err)
	}
	return &a, nil
}

func scanAmbienteBase(row interface {
	Scan(dest ...any) error
}) (*sportslot.Ambiente, error) {
	var a sportslot.Ambiente
	if err := row.Scan(&a.ID, &a.Nombre, &a.CategoriaID, &a.Activo, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, mapErr(err)
	}
	return &a, nil
}

func (r *AmbienteRepo) Create(a *sportslot.Ambiente) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	created, err := scanAmbienteBase(r.db.Pool.QueryRow(ctx,
		`INSERT INTO ambientes (nombre, categoria_id, activo)
		 VALUES ($1, $2, $3)
		 RETURNING `+ambienteBaseCols,
		a.Nombre, a.CategoriaID, a.Activo,
	))
	if err != nil {
		return err
	}
	*a = *created
	return nil
}

func (r *AmbienteRepo) FindByID(id int64) (*sportslot.Ambiente, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return scanAmbiente(r.db.Pool.QueryRow(ctx, ambienteSelect+` WHERE a.id = $1`, id))
}

func (r *AmbienteRepo) FindAll() ([]sportslot.Ambiente, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.db.Pool.Query(ctx, ambienteSelect+` ORDER BY a.categoria_id, a.nombre, a.id`)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()

	out := []sportslot.Ambiente{}
	for rows.Next() {
		a, err := scanAmbiente(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, mapErr(rows.Err())
}

func (r *AmbienteRepo) Update(id int64, a *sportslot.Ambiente) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updated, err := scanAmbienteBase(r.db.Pool.QueryRow(ctx,
		`UPDATE ambientes
		    SET nombre = $2, categoria_id = $3, activo = $4
		  WHERE id = $1
		  RETURNING `+ambienteBaseCols,
		id, a.Nombre, a.CategoriaID, a.Activo,
	))
	if err != nil {
		return err
	}
	updated.ImagenPrincipal = a.ImagenPrincipal
	*a = *updated
	return nil
}

func (r *AmbienteRepo) Delete(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tag, err := r.db.Pool.Exec(ctx, `DELETE FROM ambientes WHERE id = $1`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return sportslot.ErrNotFound
	}
	return nil
}

// DeleteCascade elimina el ambiente y sus ocupaciones en una sola transacción.
func (r *AmbienteRepo) DeleteCascade(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return mapErr(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM ocupaciones WHERE ambiente_id = $1`, id); err != nil {
		return mapErr(err)
	}

	tag, err := tx.Exec(ctx, `DELETE FROM ambientes WHERE id = $1`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return sportslot.ErrNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return mapErr(err)
	}
	return nil
}
