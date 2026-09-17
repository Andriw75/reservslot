package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"back/services/sportslot"
)

type OcupacionRepo struct {
	db *DB
}

func NewOcupacionRepo(db *DB) *OcupacionRepo { return &OcupacionRepo{db: db} }

const ocupacionCols = `id, ambiente_id, inicio, fin, activo, created_by, updated_at`

func scanOcupacion(row interface {
	Scan(dest ...any) error
}) (*sportslot.Ocupacion, error) {
	var o sportslot.Ocupacion
	if err := row.Scan(&o.ID, &o.AmbienteID, &o.Inicio, &o.Fin, &o.Activo, &o.CreatedBy, &o.UpdatedAt); err != nil {
		return nil, mapErr(err)
	}
	return &o, nil
}

// Create inserta la ocupación dentro de una transacción y toma un lock por
// ambiente para evitar condiciones de carrera entre dos inserciones
// concurrentes. La restricción EXCLUDE de la tabla es la segunda barrera.
func (r *OcupacionRepo) Create(o *sportslot.Ocupacion) error {
	if !o.Inicio.Before(o.Fin) {
		return sportslot.ErrInvalidTimeRange
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return mapErr(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := lockAmbiente(ctx, tx, o.AmbienteID); err != nil {
		return err
	}

	if o.Activo {
		var solapa bool
		if err := tx.QueryRow(ctx,
			`SELECT ocupacion_solapa($1, $2, $3, NULL)`, o.AmbienteID, o.Inicio, o.Fin,
		).Scan(&solapa); err != nil {
			return mapErr(err)
		}
		if solapa {
			return sportslot.ErrOverlap
		}
	}

	created, err := scanOcupacion(tx.QueryRow(ctx,
		`INSERT INTO ocupaciones (ambiente_id, inicio, fin, activo, created_by)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING `+ocupacionCols,
		o.AmbienteID, o.Inicio, o.Fin, o.Activo, o.CreatedBy,
	))
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return mapErr(err)
	}

	*o = *created
	return nil
}

func (r *OcupacionRepo) Update(id int64, o *sportslot.Ocupacion) error {
	if !o.Inicio.Before(o.Fin) {
		return sportslot.ErrInvalidTimeRange
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return mapErr(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := lockAmbiente(ctx, tx, o.AmbienteID); err != nil {
		return err
	}

	if o.Activo {
		var solapa bool
		if err := tx.QueryRow(ctx,
			`SELECT ocupacion_solapa($1, $2, $3, $4)`, o.AmbienteID, o.Inicio, o.Fin, id,
		).Scan(&solapa); err != nil {
			return mapErr(err)
		}
		if solapa {
			return sportslot.ErrOverlap
		}
	}

	updated, err := scanOcupacion(tx.QueryRow(ctx,
		`UPDATE ocupaciones
		    SET ambiente_id = $2, inicio = $3, fin = $4, activo = $5
		  WHERE id = $1
		  RETURNING `+ocupacionCols,
		id, o.AmbienteID, o.Inicio, o.Fin, o.Activo,
	))
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return mapErr(err)
	}

	*o = *updated
	return nil
}

func (r *OcupacionRepo) Delete(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tag, err := r.db.Pool.Exec(ctx, `DELETE FROM ocupaciones WHERE id = $1`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return sportslot.ErrNotFound
	}
	return nil
}

func (r *OcupacionRepo) FindByID(id int64) (*sportslot.Ocupacion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return scanOcupacion(r.db.Pool.QueryRow(ctx,
		`SELECT `+ocupacionCols+` FROM ocupaciones WHERE id = $1`, id))
}

// ListByAmbienteIDRange devuelve las ocupaciones activas que intersecan
// [from, to). El rango de destino es semiabierto: inicio < to AND fin > from.
func (r *OcupacionRepo) ListByAmbienteIDRange(ambienteID int64, from, to time.Time) ([]sportslot.Ocupacion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.db.Pool.Query(ctx,
		`SELECT `+ocupacionCols+`
		   FROM ocupaciones
		  WHERE ambiente_id = $1
		    AND activo = true
		    AND inicio < $3
		    AND fin > $2
		  ORDER BY inicio`, ambienteID, from, to)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()

	return collectOcupaciones(rows)
}

func (r *OcupacionRepo) ListRange(from, to time.Time) ([]sportslot.Ocupacion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.db.Pool.Query(ctx,
		`SELECT `+ocupacionCols+`
		   FROM ocupaciones
		  WHERE activo = true
		    AND inicio < $2
		    AND fin > $1
		  ORDER BY inicio`, from, to)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()

	return collectOcupaciones(rows)
}

func collectOcupaciones(rows pgx.Rows) ([]sportslot.Ocupacion, error) {
	out := []sportslot.Ocupacion{}
	for rows.Next() {
		o, err := scanOcupacion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *o)
	}
	return out, mapErr(rows.Err())
}

// lockAmbiente serializa las escrituras de ocupaciones de un mismo ambiente.
func lockAmbiente(ctx context.Context, tx pgx.Tx, ambienteID int64) error {
	_, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, ambienteID)
	return mapErr(err)
}
