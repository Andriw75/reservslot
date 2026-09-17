package mock

import (
	"errors"
	"sync"
	"time"

	"back/domain"
	"back/infrastructure/auth"
)

// UserStore es el almacén en memoria de usuarios de SportSlot.
// Se usa mientras no exista la conexión real a PostgreSQL.
type UserStore struct {
	mu    sync.RWMutex
	seq   int64
	data  map[int64]*domain.User
	crypt *auth.CryptManager
}

func NewUserStore(crypt *auth.CryptManager) *UserStore {
	s := &UserStore{data: make(map[int64]*domain.User), crypt: crypt}

	hash, err := crypt.Hash("placeholder-password")
	if err != nil {
		panic(err)
	}

	now := time.Now()
	s.Insert(&domain.User{
		ID:            1,
		Username:      "usuario",
		Email:         "admin@example.com",
		Password:      hash,
		Activo:        true,
		Permisos:      AllPermissionCodes(),
		CreadoEn:      now,
		ActualizadoEn: now,
	})

	return s
}

func (s *UserStore) Insert(u *domain.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u.ID > s.seq {
		s.seq = u.ID
	}
	cpy := *u
	s.data[u.ID] = &cpy
}

func (s *UserStore) FindByID(id int64) (*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.data[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	cpy := *u
	return &cpy, nil
}

func (s *UserStore) FindByUsername(username string) (*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.data {
		if u.Username == username {
			cpy := *u
			return &cpy, nil
		}
	}
	return nil, errors.New("user not found")
}

func (s *UserStore) FindAll() []domain.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.User, 0, len(s.data))
	for _, u := range s.data {
		out = append(out, *u)
	}
	return out
}

// IsValidPassword valida credenciales contra el hash bcrypt almacenado.
func (s *UserStore) IsValidPassword(crypt *auth.CryptManager, username, password string) (*domain.User, bool) {
	user, err := s.FindByUsername(username)
	if err != nil {
		return nil, false
	}
	if !user.Activo {
		return nil, false
	}
	if !crypt.Verify(user.Password, password) {
		return nil, false
	}
	return user, true
}
