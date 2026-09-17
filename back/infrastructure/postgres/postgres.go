package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"back/services/sportslot"
)

// DB es el pool de conexiones a PostgreSQL.
type DB struct {
	Pool *pgxpool.Pool
}

// Connect abre el pool y verifica la conexión.
func Connect(ctx context.Context, dsn string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("dsn inválido: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear el pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("no se pudo conectar a PostgreSQL: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func (d *DB) Close() {
	if d.Pool != nil {
		d.Pool.Close()
	}
}

// mapErr traduce errores de PostgreSQL a errores de dominio.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return sportslot.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23P01": // exclusion_violation (ocupaciones_no_solapadas)
			return sportslot.ErrOverlap
		case "23503": // foreign_key_violation (p. ej. ambiente con ocupaciones)
			return fmt.Errorf("%w: %s", ErrConflict, pgErr.ConstraintName)
		case "23505": // unique_violation
			return fmt.Errorf("%w: %s", ErrConflict, pgErr.ConstraintName)
		case "23514": // check_violation
			return fmt.Errorf("%w: %s", ErrInvalid, pgErr.ConstraintName)
		}
	}
	return err
}

var (
	ErrConflict = errors.New("conflicto de datos")
	ErrInvalid  = errors.New("datos inválidos")
)
