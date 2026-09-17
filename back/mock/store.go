package mock

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type MockStore[T any] struct {
	mu       sync.RWMutex
	data     map[int64]T
	nextID   atomic.Int64
	validate func(*T) error
}

func NewMockStore[T any](validate func(*T) error) *MockStore[T] {
	return &MockStore[T]{
		data:     make(map[int64]T),
		validate: validate,
	}
}

type Identifiable interface {
	SetID(int64)
}

func (s *MockStore[T]) Create(entity T) (T, error) {
	if s.validate != nil {
		if err := s.validate(&entity); err != nil {
			return entity, err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextID.Add(1)
	if setter, ok := any(&entity).(Identifiable); ok {
		setter.SetID(id)
	}
	s.data[id] = entity
	return entity, nil
}

func (s *MockStore[T]) FindByID(id int64) (T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entity, ok := s.data[id]
	if !ok {
		var zero T
		return zero, fmt.Errorf("not found")
	}
	return entity, nil
}

func (s *MockStore[T]) FindAll() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]T, 0, len(s.data))
	for _, entity := range s.data {
		result = append(result, entity)
	}
	return result
}

func (s *MockStore[T]) Update(id int64, entity T) error {
	if s.validate != nil {
		if err := s.validate(&entity); err != nil {
			return err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return fmt.Errorf("not found")
	}
	s.data[id] = entity
	return nil
}

func (s *MockStore[T]) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return fmt.Errorf("not found")
	}
	delete(s.data, id)
	return nil
}

func (s *MockStore[T]) Insert(id int64, entity T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = entity
	if id > s.nextID.Load() {
		s.nextID.Store(id)
	}
}
