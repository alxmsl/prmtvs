package skm

import (
	"sync"
)

// SSKM struct represents a thread-safe map with a sorted set of keys.
type SSKM struct {
	sync.RWMutex
	skm *SKM
}

// NewSafeSortedKeyMap creates a SSKM object with a concurrent access.
func NewSafeSortedKeyMap() *SSKM {
	return &SSKM{
		skm: NewSortedKeyMap(),
	}
}

func (sm *SSKM) Add(key string, value interface{}) bool {
	sm.Lock()
	defer sm.Unlock()
	return sm.skm.Add(key, value)
}

func (sm *SSKM) ExistsIndex(idx int) bool {
	sm.RLock()
	defer sm.RUnlock()
	return sm.skm.ExistsIndex(idx)
}

func (sm *SSKM) ExistsKey(key string) bool {
	sm.RLock()
	defer sm.RUnlock()
	return sm.skm.ExistsKey(key)
}

func (sm *SSKM) GetByIndex(idx int) (interface{}, bool) {
	sm.RLock()
	defer sm.RUnlock()
	return sm.skm.GetByIndex(idx)
}

func (sm *SSKM) GetByKey(key string) (interface{}, bool) {
	sm.RLock()
	defer sm.RUnlock()
	return sm.skm.GetByKey(key)
}

func (sm *SSKM) Index(key string) (int, bool) {
	sm.RLock()
	defer sm.RUnlock()
	return sm.skm.Index(key)
}

func (sm *SSKM) Key(idx int) (string, bool) {
	sm.RLock()
	defer sm.RUnlock()
	return sm.skm.Key(idx)
}

func (sm *SSKM) Len() int {
	sm.RLock()
	defer sm.RUnlock()
	return sm.skm.Len()
}

func (sm *SSKM) Over(fn OverFunc) {
	sm.RLock()
	defer sm.RUnlock()
	sm.skm.Over(fn)
}

func (sm *SSKM) Reset() {
	sm.RLock()
	defer sm.RUnlock()
	sm.skm.Reset()
}
