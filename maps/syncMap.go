package maps

import "sync"

type _SyncMap[K comparable, V any] struct {
	m sync.Map
}
type SyncMap[K comparable, V any] struct {
	*_SyncMap[K, V]
}

func NewSyncMap[K comparable, V any]() SyncMap[K, V] {
	return SyncMap[K, V]{
		_SyncMap: &_SyncMap[K, V]{},
	}
}

func (s *SyncMap[K, V]) Load(key *K) (value V, ok bool) {
	x, ok := s.m.Load(*key)
	value, _ = x.(V)
	return value, ok
}

func (s *SyncMap[K, V]) Store(key *K, value V) {
	s.m.Store(*key, value)
}

func (s *SyncMap[K, V]) LoadOrStore(key *K, value V) (actual V, loaded bool) {
	x, loaded := s.m.LoadOrStore(*key, value)
	actual, _ = x.(V)
	return actual, loaded
}

func (s *SyncMap[K, V]) LoadAndDelete(key *K) (value V, loaded bool) {
	x, loaded := s.m.LoadAndDelete(*key)
	value, _ = x.(V)
	return value, loaded
}

func (s *SyncMap[K, V]) Delete(key *K) {
	s.m.Delete(*key)
}

func (s *SyncMap[K, V]) Range(f func(key *K, value V) bool) {
	if f == nil {
		return
	}
	s.m.Range(func(key, value any) bool {
		k, _ := key.(K)
		v, _ := value.(V)
		return f(&k, v)
	})
}
