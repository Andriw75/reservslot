package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"back/api/handlers"
)

func NewRouter(deps *handlers.Dependencies) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(corsMiddleware(deps.Config.CORSOrigin))

	authHandler := handlers.NewAuthHandler(deps)
	sportSlot := handlers.NewSportSlotHandlers(deps.SportSlotService, deps.SportSlotBus)
	sportSlotSSE := handlers.NewSportSlotSSEHandler(deps.SportSlotService, deps.SportSlotBus)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// ---------------------------------------------------------------- Estáticos
	// En desarrollo Go sirve las imágenes directamente; en producción nginx
	// puede servirlas desde la misma carpeta.
	r.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.Dir(deps.Config.PublicPath()))))

	// ---------------------------------------------------------------- Auth
	r.Route("/auth", func(r chi.Router) {
		r.Post("/token", authHandler.Token)
		r.Group(func(r chi.Router) {
			r.Use(authHandler.AuthMiddleware())
			r.Post("/logout", authHandler.Logout)
			r.Get("/me", authHandler.Me)
		})
	})

	// ---------------------------------------------------------------- SSE
	// Un único stream multiplexado. Público (no exige sesión).
	r.Get("/events", sportSlotSSE.Events)

	// ------------------------------------------------------------- Lectura pública
	r.Get("/categorias", sportSlot.GetCategorias)
	r.Get("/categorias/{id}", sportSlot.GetCategoria)
	r.Get("/ambientes", sportSlot.GetAmbientes)
	r.Get("/ambientes/{id}", sportSlot.GetAmbiente)
	r.Get("/ambientes/{id}/info", sportSlot.GetAmbienteInfo)
	r.Get("/ambientes/{id}/ocupaciones", sportSlot.GetOcupaciones)

	// ------------------------------------------------------------- Escritura (requiere sesión)
	r.Group(func(r chi.Router) {
		r.Use(authHandler.AuthMiddleware())

		r.Post("/categorias", sportSlot.PostCategoria)
		r.Put("/categorias/{id}", sportSlot.PutPatchCategoria)
		r.Patch("/categorias/{id}", sportSlot.PutPatchCategoria)
		r.Delete("/categorias/{id}", sportSlot.DeleteCategoria)

		r.Post("/ambientes", sportSlot.PostAmbiente)
		r.Put("/ambientes/{id}", sportSlot.UpsertAmbiente)
		r.Patch("/ambientes/{id}", sportSlot.UpsertAmbiente)
		r.Delete("/ambientes/{id}", sportSlot.DeleteAmbiente)
		r.Put("/ambientes/{id}/info", sportSlot.UpsertAmbienteInfo)
		r.Patch("/ambientes/{id}/info", sportSlot.UpsertAmbienteInfo)
		r.Post("/ambientes/{id}/imagenes", sportSlot.UploadAmbienteImagenes)
		r.Put("/ambientes/{id}/imagenes/orden", sportSlot.ReorderAmbienteImagenes)
		r.Delete("/ambientes/{id}/imagenes/{imagenID}", sportSlot.DeleteAmbienteImagen)

		r.Post("/ocupaciones", sportSlot.PostOcupacion)
		r.Put("/ocupaciones/{id}", sportSlot.PutOcupacion)
		r.Patch("/ocupaciones/{id}", sportSlot.PutOcupacion)
		r.Delete("/ocupaciones/{id}", sportSlot.DeleteOcupacion)
	})

	return r
}

func corsMiddleware(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
