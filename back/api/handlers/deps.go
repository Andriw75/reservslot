package handlers

import (
	"back/domain"
	"back/infrastructure/auth"
	"back/infrastructure/storage"
	"back/services"
	"back/services/sportslot"
)

// UserStore es el contrato mínimo que el handler de auth necesita.
type UserStore interface {
	FindByID(id int64) (*domain.User, error)
	IsValidPassword(username, password string) (*domain.User, bool)
}

type Dependencies struct {
	Config       *services.Config
	JWTManager   *auth.JWTManager
	CryptManager *auth.CryptManager
	JWTStore     *auth.JWTStore

	UserStore UserStore

	Storage *storage.Storage

	SportSlotService *sportslot.ServicioSportSlot
	SportSlotBus     *sportslot.SSEBus
}
