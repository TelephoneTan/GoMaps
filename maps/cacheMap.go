package maps

import (
	"sync"
	"time"
)

type timeData[D any] struct {
	time int64
	data D
}

func newTimeData[D any](data D) timeData[D] {
	return timeData[D]{time: time.Now().UnixNano(), data: data}
}

type _CacheMap[K comparable, V any] struct {
	Concurrent bool
	m          map[K]timeData[V]
	rLock      sync.Locker
	wLock      sync.Locker
}
type CacheMap[K comparable, V any] struct {
	*_CacheMap[K, V]
}

func NewCacheMap[K comparable, V any](concurrent bool, init ...func(CacheMap[K, V])) CacheMap[K, V] {
	m := CacheMap[K, V]{
		_CacheMap: &_CacheMap[K, V]{
			Concurrent: concurrent,
		},
	}
	if len(init) > 0 {
		init[0](m)
	}
	m.m = make(map[K]timeData[V])
	if m.Concurrent {
		m.rLock = new(sync.RWMutex).RLocker()
		m.wLock = new(sync.Mutex)
	}
	return m
}

func (s *CacheMap[K, V]) Load(key K) (value V, ok bool) {
	if s.Concurrent {
		s.rLock.Lock()
		defer s.rLock.Unlock()
	}
	x, ok := s.m[key]
	if ok {
		value = x.data
	}
	return value, ok
}

func (s *CacheMap[K, V]) Store(key K, value V) {
	if s.Concurrent {
		s.rLock.Lock()
		defer s.rLock.Unlock()
		s.wLock.Lock()
		defer s.wLock.Unlock()
	}
	s.m[key] = newTimeData(value)
}

func (s *CacheMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	if s.Concurrent {
		s.rLock.Lock()
		defer s.rLock.Unlock()
	}
	x, loaded := s.m[key]
	if !loaded {
		if s.Concurrent {
			s.wLock.Lock()
			defer s.wLock.Unlock()
		}
		x = newTimeData(value)
		s.m[key] = x
	}
	actual = x.data
	return actual, loaded
}

func (s *CacheMap[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	if s.Concurrent {
		s.rLock.Lock()
		defer s.rLock.Unlock()
		s.wLock.Lock()
		defer s.wLock.Unlock()
	}
	x, loaded := s.m[key]
	if loaded {
		value = x.data
		delete(s.m, key)
	}
	return value, loaded
}

func (s *CacheMap[K, V]) Delete(key K) {
	if s.Concurrent {
		s.rLock.Lock()
		defer s.rLock.Unlock()
		s.wLock.Lock()
		defer s.wLock.Unlock()
	}
	delete(s.m, key)
}

func (s *CacheMap[K, V]) Range(f func(key K, value V) bool) {
	if f == nil {
		return
	}
	for k, v := range s.m {
		if !f(k, v.data) {
			break
		}
	}
}
