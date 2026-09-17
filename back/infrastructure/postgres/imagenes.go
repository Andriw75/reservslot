package postgres

import (
	"context"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"back/infrastructure/storage"
	"back/services/sportslot"
)

type ImagenRepo struct {
	db *DB
}

func NewImagenRepo(db *DB) *ImagenRepo { return &ImagenRepo{db: db} }

const imagenCols = `id, ambiente_id, archivo, descripcion, orden, mime_type, bytes, activo, created_at, updated_at`

func scanImagen(row interface {
	Scan(dest ...any) error
}) (*sportslot.Imagen, error) {
	var i sportslot.Imagen
	if err := row.Scan(&i.ID, &i.AmbienteID, &i.Archivo, &i.Descripcion, &i.Orden, &i.MimeType, &i.Bytes, &i.Activo, &i.CreatedAt, &i.UpdatedAt); err != nil {
		return nil, mapErr(err)
	}
	return &i, nil
}

// AddFiles guarda los archivos en disco y luego registra los metadatos.
// Si falla la inserción en base de datos, elimina los archivos ya escritos.
func (r *ImagenRepo) AddFiles(ambienteID int64, archivos []*multipart.FileHeader, store *storage.Storage, allowedExts map[string]bool) ([]sportslot.Imagen, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filenames, err := store.SaveFiles(archivos, prefix("ambiente", ambienteID), allowedExts)
	if err != nil {
		return nil, err
	}

	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		store.RemoveImages(filenames...)
		return nil, mapErr(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var nextOrden int
	if err := tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(orden), -1) + 1 FROM imagenes WHERE ambiente_id = $1`,
		ambienteID,
	).Scan(&nextOrden); err != nil {
		store.RemoveImages(filenames...)
		return nil, mapErr(err)
	}

	created := make([]sportslot.Imagen, 0, len(filenames))
	for i, fn := range filenames {
		var mime *string
		if m := mimeFromExt(fn); m != "" {
			mime = &m
		}
		img, err := scanImagen(tx.QueryRow(ctx,
			`INSERT INTO imagenes (ambiente_id, archivo, orden, mime_type)
			 VALUES ($1, $2, $3, $4)
			 RETURNING `+imagenCols,
			ambienteID, fn, nextOrden+i, mime,
		))
		if err != nil {
			store.RemoveImages(filenames...)
			return nil, err
		}
		created = append(created, *img)
	}

	if err := tx.Commit(ctx); err != nil {
		store.RemoveImages(filenames...)
		return nil, mapErr(err)
	}
	return created, nil
}

func (r *ImagenRepo) ListByAmbienteID(ambienteID int64) ([]sportslot.Imagen, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.db.Pool.Query(ctx,
		`SELECT `+imagenCols+`
		   FROM imagenes
		  WHERE ambiente_id = $1 AND activo = true
		  ORDER BY orden, id`, ambienteID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()

	out := []sportslot.Imagen{}
	for rows.Next() {
		img, err := scanImagen(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *img)
	}
	return out, mapErr(rows.Err())
}

func (r *ImagenRepo) DeleteByID(id int64) (*sportslot.Imagen, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	img, err := scanImagen(r.db.Pool.QueryRow(ctx,
		`DELETE FROM imagenes WHERE id = $1 RETURNING `+imagenCols, id))
	if err != nil {
		return nil, err
	}
	return img, nil
}

// UpdateOrden reasigna el orden 0..n-1 según el orden de ids recibido.
// El índice único (ambiente_id, orden) obliga a mover los valores a un rango
// temporal antes de asignar los definitivos.
func (r *ImagenRepo) UpdateOrden(ambienteID int64, ids []int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return mapErr(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		`UPDATE imagenes SET orden = orden + 100000 WHERE ambiente_id = $1`, ambienteID,
	); err != nil {
		return mapErr(err)
	}

	for i, id := range ids {
		tag, err := tx.Exec(ctx,
			`UPDATE imagenes SET orden = $3 WHERE id = $1 AND ambiente_id = $2`, id, ambienteID, i)
		if err != nil {
			return mapErr(err)
		}
		if tag.RowsAffected() == 0 {
			return sportslot.ErrNotFound
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return mapErr(err)
	}
	return nil
}

func prefix(kind string, id int64) string {
	return kind + "_" + strconv.FormatInt(id, 10)
}

func mimeFromExt(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return ""
	}
}
