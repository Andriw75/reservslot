package postgres

import (
	"context"
	"encoding/json"
	"time"

	"back/services/sportslot"
)

type AmbienteInfoRepo struct {
	db *DB
}

func NewAmbienteInfoRepo(db *DB) *AmbienteInfoRepo { return &AmbienteInfoRepo{db: db} }

const ambienteInfoCols = `ambiente_id, data, activo, created_at, updated_at`

func scanAmbienteInfo(row interface {
	Scan(dest ...any) error
}) (*sportslot.AmbienteInfo, error) {
	var i sportslot.AmbienteInfo
	if err := row.Scan(&i.AmbienteID, &i.Data, &i.Activo, &i.CreatedAt, &i.UpdatedAt); err != nil {
		return nil, mapErr(err)
	}
	return &i, nil
}

func (r *AmbienteInfoRepo) Upsert(info *sportslot.AmbienteInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data := info.Data
	if len(data) == 0 {
		data = json.RawMessage(`{}`)
	}

	err := r.db.Pool.QueryRow(ctx,
		`INSERT INTO ambiente_info (ambiente_id, data, activo)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (ambiente_id)
		 DO UPDATE SET data = EXCLUDED.data, activo = EXCLUDED.activo
		 RETURNING `+ambienteInfoCols,
		info.AmbienteID, data, info.Activo,
	).Scan(&info.AmbienteID, &info.Data, &info.Activo, &info.CreatedAt, &info.UpdatedAt)
	return mapErr(err)
}

func (r *AmbienteInfoRepo) FindByAmbienteID(id int64) (*sportslot.AmbienteInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return scanAmbienteInfo(r.db.Pool.QueryRow(ctx,
		`SELECT `+ambienteInfoCols+` FROM ambiente_info WHERE ambiente_id = $1`, id))
}

func (r *AmbienteInfoRepo) DeleteByAmbienteID(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tag, err := r.db.Pool.Exec(ctx, `DELETE FROM ambiente_info WHERE ambiente_id = $1`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return sportslot.ErrNotFound
	}
	return nil
}
