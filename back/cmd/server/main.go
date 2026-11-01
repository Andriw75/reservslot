package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"back/api"
	"back/api/handlers"
	"back/infrastructure/auth"
	"back/infrastructure/postgres"
	"back/infrastructure/storage"
	"back/services"
	"back/services/sportslot"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := services.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if cfg.JWTSecret == "" || cfg.JWTSecret == "cambia-este-secreto-en-produccion" {
		slog.Error("JWT_SECRET no configurado: genera uno con 'go run ./cmd/generate_secrets.go' y defínelo en el entorno")
		os.Exit(1)
	}

	dsn := cfg.DBDSN()
	if dsn == "" {
		slog.Error("DATABASE_URL no está configurado")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	db, err := postgres.Connect(ctx, dsn)
	cancel()
	if err != nil {
		slog.Error("no se pudo conectar a PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("conectado a PostgreSQL", "host", cfg.DBHost, "db", cfg.DBName)

	jwtStore := auth.NewJWTStore()
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpirationMin, cfg.JWTRenewMin)
	cryptManager := auth.NewCryptManager()
	if cfg.JWTStoreEnabled {
		jwtManager.Store = jwtStore
	}

	st := storage.NewStorage(cfg.ImagesDir())
	if err := st.EnsureDirs(); err != nil {
		slog.Error("failed to create images dir", "error", err, "dir", cfg.ImagesDir())
		os.Exit(1)
	}

	bus := sportslot.NewSSEBus()
	service := &sportslot.ServicioSportSlot{
		Categorias:   postgres.NewCategoriaRepo(db),
		Ambientes:    postgres.NewAmbienteRepo(db),
		AmbienteInfo: postgres.NewAmbienteInfoRepo(db),
		Imagenes:     postgres.NewImagenRepo(db),
		Ocupaciones:  postgres.NewOcupacionRepo(db),
		Storage:      st,
		SSEBus:       bus,
	}

	deps := &handlers.Dependencies{
		Config:           cfg,
		JWTManager:       jwtManager,
		CryptManager:     cryptManager,
		JWTStore:         jwtStore,
		UserStore:        postgres.NewUsuarioRepo(db, cryptManager),
		Storage:          st,
		SportSlotService: service,
		SportSlotBus:     bus,
	}

	router := api.NewRouter(deps)

	srv := &http.Server{
		Addr:              cfg.ServerPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("servidor iniciado", "addr", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("apagando servidor...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("apagado forzado", "error", err)
	}
	slog.Info("servidor detenido")
}
