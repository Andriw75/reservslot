package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"back/infrastructure/auth"
	"back/infrastructure/postgres"
	"back/services"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	username := strings.TrimSpace(os.Getenv("ADMIN_USERNAME"))
	password := os.Getenv("ADMIN_PASSWORD")
	if username == "" || password == "" {
		slog.Error("define ADMIN_USERNAME y ADMIN_PASSWORD en el entorno")
		os.Exit(1)
	}
	if len(password) < 8 {
		slog.Error("ADMIN_PASSWORD debe tener al menos 8 caracteres")
		os.Exit(1)
	}

	cfg, err := services.LoadConfig()
	if err != nil {
		slog.Error("no se pudo cargar la configuracion", "error", err)
		os.Exit(1)
	}

	dsn := cfg.DBDSN()
	if dsn == "" {
		slog.Error("DATABASE_URL o las variables DB_* no estan configuradas")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := postgres.Connect(ctx, dsn)
	if err != nil {
		slog.Error("no se pudo conectar a PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	crypt := auth.NewCryptManager()
	hash, err := crypt.Hash(password)
	if err != nil {
		slog.Error("no se pudo generar el hash de la contrasena", "error", err)
		os.Exit(1)
	}

	var id int64
	err = db.Pool.QueryRow(ctx,
		`SELECT id FROM usuarios WHERE lower(username) = lower($1)`, username).Scan(&id)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		_, err = db.Pool.Exec(ctx,
			`INSERT INTO usuarios (username, password_hash, activo) VALUES ($1, $2, true)`,
			username, hash)
		if err != nil {
			slog.Error("no se pudo crear el usuario", "error", err)
			os.Exit(1)
		}
		slog.Info("usuario creado", "username", username)
	case err != nil:
		slog.Error("no se pudo consultar el usuario", "error", err)
		os.Exit(1)
	default:
		_, err = db.Pool.Exec(ctx,
			`UPDATE usuarios SET password_hash = $2, activo = true WHERE id = $1`,
			id, hash)
		if err != nil {
			slog.Error("no se pudo actualizar el usuario", "error", err)
			os.Exit(1)
		}
		slog.Info("usuario actualizado", "username", username)
	}
}
