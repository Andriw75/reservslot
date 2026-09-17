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
	"back/infrastructure/storage"
	"back/mock"
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

	// ---------------------------------------------------------------- Datos
	// Fase inicial: stores en memoria. La infraestructura PostgreSQL queda
	// lista (schema.sql + DSN en config) para sustituir estos stores.
	bus := sportslot.NewSSEBus()

	service := &sportslot.ServicioSportSlot{
		Categorias:   sportslot.NewMockCategoriaStore(),
		Ambientes:    sportslot.NewMockAmbienteStore(),
		AmbienteInfo: sportslot.NewMockAmbienteInfoStore(),
		Imagenes:     sportslot.NewMockImagenStore(),
		Ocupaciones:  sportslot.NewMockOcupacionStore(),
		Storage:      st,
		SSEBus:       bus,
	}
	seedSportSlot(service)

	deps := &handlers.Dependencies{
		Config:           cfg,
		JWTManager:       jwtManager,
		CryptManager:     cryptManager,
		JWTStore:         jwtStore,
		UserStore:        mock.NewUserStore(cryptManager),
		PermisoStore:     mock.NewPermisoStore(),
		Storage:          st,
		SportSlotService: service,
		SportSlotBus:     bus,
	}

	if cfg.UseMock {
		slog.Info("usando stores en memoria (USE_MOCK=true)")
	} else {
		slog.Warn("USE_MOCK=false pero los repositorios PostgreSQL aún no están implementados; se usan stores en memoria",
			"dsn_host", cfg.DBHost, "db", cfg.DBName)
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("apagado forzado", "error", err)
	}
	slog.Info("servidor detenido")
}

// seedSportSlot carga datos de ejemplo para poder probar el sistema.
func seedSportSlot(s *sportslot.ServicioSportSlot) {
	catCanchas := &sportslot.Categoria{Nombre: "Canchas", Descripcion: "Canchas deportivas", Color: strPtr("#2563EB"), Activo: true}
	catSalas := &sportslot.Categoria{Nombre: "Salas", Descripcion: "Salas de reunión", Color: strPtr("#16A34A"), Activo: true}
	_ = s.Categorias.Create(catCanchas)
	_ = s.Categorias.Create(catSalas)

	ambCancha1 := &sportslot.Ambiente{Nombre: "Cancha 1", CategoriaID: catCanchas.ID, Activo: true}
	ambCancha2 := &sportslot.Ambiente{Nombre: "Cancha 2", CategoriaID: catCanchas.ID, Activo: true}
	ambSalaA := &sportslot.Ambiente{Nombre: "Sala A", CategoriaID: catSalas.ID, Activo: true}
	_ = s.Ambientes.Create(ambCancha1)
	_ = s.Ambientes.Create(ambCancha2)
	_ = s.Ambientes.Create(ambSalaA)

	_, _ = s.UpsertAmbienteInfo(ambCancha1.ID, []byte(`{"ancho":20,"alto":10,"aforo":30,"descripcion":"Cancha techada con iluminación","opciones_adicionales":{"iluminacion":true,"vestidores":true,"proyector":false}}`), true)
	_, _ = s.UpsertAmbienteInfo(ambCancha2.ID, []byte(`{"ancho":18,"alto":9,"aforo":24,"descripcion":"Cancha descubierta","opciones_adicionales":{"iluminacion":false,"vestidores":true,"proyector":false}}`), true)
	_, _ = s.UpsertAmbienteInfo(ambSalaA.ID, []byte(`{"ancho":8,"alto":6,"aforo":12,"descripcion":"Sala con proyector","opciones_adicionales":{"iluminacion":true,"vestidores":false,"proyector":true}}`), true)

	tr := sportslot.WeekRangeAmericaLima(time.Now())
	base := tr.From.AddDate(0, 0, 1).Add(9 * time.Hour) // martes 09:00
	_ = s.Ocupaciones.Create(&sportslot.Ocupacion{AmbienteID: ambCancha1.ID, Inicio: base, Fin: base.Add(90 * time.Minute), Activo: true})
	_ = s.Ocupaciones.Create(&sportslot.Ocupacion{AmbienteID: ambCancha1.ID, Inicio: base.Add(2 * time.Hour), Fin: base.Add(3 * time.Hour), Activo: true})
	_ = s.Ocupaciones.Create(&sportslot.Ocupacion{AmbienteID: ambSalaA.ID, Inicio: base.Add(1 * time.Hour), Fin: base.Add(2 * time.Hour), Activo: true})
}

func strPtr(s string) *string { return &s }
