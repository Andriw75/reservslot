package database

import "context"

type Repository[T any, ID comparable] interface {
	Create(ctx context.Context, entity *T) error
	FindByID(ctx context.Context, id ID) (*T, error)
	FindAll(ctx context.Context) ([]T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id ID) error
}

type UserRepository interface {
	Repository[any, int64]
	FindByUsername(ctx context.Context, username string) (*any, error)
}

// TODO: When implementing database repositories:
//   1. Create a struct per entity (e.g., UserRepo, PermissionRepo)
//   2. Inject *pgxpool.Pool for PostgreSQL or *sql.DB for MySQL
//   3. Implement all methods of the Repository interface
//   4. Use the CrudRepository[T] generic helpers below

// CrudRepository provides generic CRUD helpers for database operations.
// Actual implementation depends on the chosen database driver.
type CrudRepository[T any] struct {
	table   string
	columns []string
	// pool *pgxpool.Pool  // Uncomment for PostgreSQL
	// db   *sql.DB        // Uncomment for MySQL
}

func NewCrudRepository[T any](table string, columns []string) *CrudRepository[T] {
	return &CrudRepository[T]{table: table, columns: columns}
}
