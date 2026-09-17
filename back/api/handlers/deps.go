package handlers

import (
	"back/infrastructure/auth"
	"back/infrastructure/storage"
	"back/mock"
	"back/services"
	"back/services/sportslot"
)

type Dependencies struct {
	Config       *services.Config
	JWTManager   *auth.JWTManager
	CryptManager *auth.CryptManager
	JWTStore     *auth.JWTStore

	UserStore    *mock.UserStore
	PermisoStore *mock.PermisoStore

	Storage *storage.Storage

	SportSlotService *sportslot.ServicioSportSlot
	SportSlotBus     *sportslot.SSEBus
}
