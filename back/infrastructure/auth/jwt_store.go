package auth

import (
	"sync"
)

type JWTStore struct {
	mu         sync.RWMutex
	tokens     map[string]int64
	userTokens map[int64]map[string]struct{}
}

func NewJWTStore() *JWTStore {
	return &JWTStore{
		tokens:     make(map[string]int64),
		userTokens: make(map[int64]map[string]struct{}),
	}
}

func (s *JWTStore) Add(jti string, userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[jti] = userID
	if s.userTokens[userID] == nil {
		s.userTokens[userID] = make(map[string]struct{})
	}
	s.userTokens[userID][jti] = struct{}{}
}

func (s *JWTStore) Remove(jti string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	userID, ok := s.tokens[jti]
	if ok {
		delete(s.tokens, jti)
		if s.userTokens[userID] != nil {
			delete(s.userTokens[userID], jti)
		}
	}
}

func (s *JWTStore) RemoveUserTokens(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if tokens, ok := s.userTokens[userID]; ok {
		for jti := range tokens {
			delete(s.tokens, jti)
		}
		delete(s.userTokens, userID)
	}
}

func (s *JWTStore) Exists(jti string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.tokens[jti]
	return ok
}

func (s *JWTStore) CountUserTokens(userID int64) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.userTokens[userID])
}
