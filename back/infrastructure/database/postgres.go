package database

/*
PostgreSQL implementation using pgx v5.

To activate:
  1. Uncomment the code below
  2. Add to go.mod:
       go get github.com/jackc/pgx/v5

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPool struct {
	Pool *pgxpool.Pool
}

func NewPostgresPool(ctx context.Context, dsn string) (*PostgresPool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &PostgresPool{Pool: pool}, nil
}

func (p *PostgresPool) Close() {
	p.Pool.Close()
}

// Example usage in a UserRepository:
//
// type UserRepoPostgres struct {
//     pool *pgxpool.Pool
// }
//
// func NewUserRepoPostgres(pool *pgxpool.Pool) *UserRepoPostgres {
//     return &UserRepoPostgres{pool: pool}
// }
//
// func (r *UserRepoPostgres) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
//     query := `SELECT id, username, email, password, activo FROM usuarios WHERE username = $1`
//     var user domain.User
//     err := r.pool.QueryRow(ctx, query, username).Scan(
//         &user.ID, &user.Username, &user.Email, &user.Password, &user.Activo,
//     )
//     if err != nil {
//         return nil, err
//     }
//     return &user, nil
// }
*/
