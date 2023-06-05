package maps

import (
	"sync"
	"time"
)

type _timeData[D any] struct {
	tsNano int64
	data   D
}
type timeData[D any] struct {
	*_timeData[D]
}

func nowNano() int64 {
	return time.Now().UnixNano()
}

func newTimeData[D any](data D) timeData[D] {
	return timeData[D]{_timeData: &_timeData[D]{tsNano: nowNano(), data: data}}
}

type _CacheMap[K comparable, V any] struct {
	Concurrent    bool
	SizeThreshold int
	TTL           time.Duration
	m             map[K]timeData[V]
	rLock         sync.Locker
	wLock         sync.Locker
}
type CacheMap[K comparable, V any] struct {
	*_CacheMap[K, V]
}

func NewCacheMap[K comparable, V any](concurrent bool, sizeThreshold int, ttl time.Duration, init ...func(CacheMap[K, V])) CacheMap[K, V] {
	m := CacheMap[K, V]{
		_CacheMap: &_CacheMap[K, V]{
			Concurrent:    concurrent,
			SizeThreshold: sizeThreshold,
			TTL:           ttl,
		},
	}
	if len(init) > 0 {
		init[0](m)
	}
	m.m = make(map[K]timeData[V])
	if m.Concurrent {
		rw := new(sync.RWMutex)
		m.rLock = rw.RLocker()
		m.wLock = rw
	}
	return m
}

func (s *CacheMap[K, V]) Load(key *K) (value V, ok bool) {
	if s.Concurrent {
		s.rLock.Lock()
		defer s.rLock.Unlock()
	}
	x, ok := s.m[*key]
	if ok {
		if s.Concurrent {
			s.rLock.Unlock()
		}
		func() {
			if s.Concurrent {
				s.wLock.Lock()
				defer s.wLock.Unlock()
			}
			x, ok = s.m[*key]
			if !ok {
				return
			}
			x.tsNano = nowNano()
			value = x.data
		}()
		if s.Concurrent {
			s.rLock.Lock()
		}
	}
	return value, ok
}

func (s *CacheMap[K, V]) clear() (cleared []*K) {
	n := nowNano()
	for k, v := range s.m {
		if n-v.tsNano > s.TTL.Nanoseconds() {
			delete(s.m, k)
			kk := k
			cleared = append(cleared, &kk)
		}
	}
	return cleared
}

func (s *CacheMap[K, V]) Store(key *K, value V) (cleared []*K) {
	if s.Concurrent {
		s.wLock.Lock()
		defer s.wLock.Unlock()
	}
	s.m[*key] = newTimeData(value)
	if len(s.m) > s.SizeThreshold {
		cleared = s.clear()
	}
	return cleared
}

func (s *CacheMap[K, V]) LoadOrStore(key *K, value V) (actual V, loaded bool, cleared []*K) {
	if s.Concurrent {
		s.wLock.Lock()
		defer s.wLock.Unlock()
	}
	x, loaded := s.m[*key]
	if !loaded {
		x = newTimeData(value)
		s.m[*key] = x
		if len(s.m) > s.SizeThreshold {
			cleared = s.clear()
		}
	} else {
		x.tsNano = nowNano()
	}
	actual = x.data
	return actual, loaded, cleared
}

func (s *CacheMap[K, V]) LoadAndDelete(key *K) (value V, loaded bool) {
	if s.Concurrent {
		s.rLock.Lock()
		defer s.rLock.Unlock()
	}
	x, loaded := s.m[*key]
	if loaded {
		if s.Concurrent {
			s.rLock.Unlock()
		}
		func() {
			if s.Concurrent {
				s.wLock.Lock()
				defer s.wLock.Unlock()
			}
			x, loaded = s.m[*key]
			if !loaded {
				return
			}
			delete(s.m, *key)
			value = x.data
		}()
		if s.Concurrent {
			s.rLock.Lock()
		}
	}
	return value, loaded
}

func (s *CacheMap[K, V]) Delete(key *K) {
	if s.Concurrent {
		s.rLock.Lock()
		defer s.rLock.Unlock()
	}
	_, loaded := s.m[*key]
	if loaded {
		if s.Concurrent {
			s.rLock.Unlock()
		}
		func() {
			if s.Concurrent {
				s.wLock.Lock()
				defer s.wLock.Unlock()
			}
			_, loaded = s.m[*key]
			if !loaded {
				return
			}
			delete(s.m, *key)
		}()
		if s.Concurrent {
			s.rLock.Lock()
		}
	}
}

func (s *CacheMap[K, V]) Range(f func(key *K, value V) bool) {
	if f == nil {
		return
	}
	if s.Concurrent {
		s.rLock.Lock()
		defer s.rLock.Unlock()
	}
	for k, v := range s.m {
		kk := k
		if !f(&kk, v.data) {
			break
		}
	}
}
