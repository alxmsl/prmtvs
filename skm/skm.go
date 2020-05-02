package skm

import (
	"sync"
)

type SKM struct {
	sync.RWMutex
	ca bool

	m  map[string]interface{}
	kk []string
}

func NewSafeSKM() *SKM {
	skm := NewSKM()
	skm.ca = true
	return skm
}

func NewSKM() *SKM {
	return &SKM{
		m:  map[string]interface{}{},
		kk: []string{},
	}
}

func (sm *SKM) Add(k string, v interface{}) bool {
	if sm.ca {
		sm.Lock()
		defer sm.Unlock()
	}
	if _, ok := sm.m[k]; ok {
		return !ok
	}

	i := 0
	for i = 0; i < len(sm.kk); i += 1 {
		if k < sm.kk[i] {
			break
		}
	}
	sm.kk = append(sm.kk[:i], append([]string{k}, sm.kk[i:]...)...)
	sm.m[k] = v
	return true
}

func (sm *SKM) ExistsIndex(i int) bool {
	if sm.ca {
		sm.RLock()
		defer sm.RUnlock()
	}
	return i >= 0 && i < len(sm.kk)
}

func (sm *SKM) ExistsKey(k string) bool {
	if sm.ca {
		sm.RLock()
		defer sm.RUnlock()
	}
	_, ok := sm.m[k]
	return ok
}

func (sm *SKM) GetByKey(k string) (interface{}, bool) {
	if sm.ca {
		sm.RLock()
		defer sm.RUnlock()
	}
	p, ok := sm.m[k]
	return p, ok
}

func (sm *SKM) GetByIndex(i int) (interface{}, bool) {
	if sm.ca {
		sm.RLock()
		defer sm.RUnlock()
	}
	if i < 0 || i >= len(sm.kk) {
		return nil, false
	}
	v, ok := sm.m[sm.kk[i]]
	return v, ok
}

func (sm *SKM) Index(k string) int {
	if sm.ca {
		sm.RLock()
		defer sm.RUnlock()
	}
	for i, n := range sm.kk {
		if k == n {
			return i
		}
	}
	return -1
}

func (sm *SKM) Key(i int) string {
	if sm.ca {
		sm.RLock()
		defer sm.RUnlock()
	}
	return sm.kk[i]
}

func (sm *SKM) Len() int {
	if sm.ca {
		sm.RLock()
		defer sm.RUnlock()
	}
	return len(sm.kk)
}

type OverFunc func(int, string, interface{}) bool

func (sm *SKM) Over(fn OverFunc) {
	if sm.ca {
		sm.RLock()
		defer sm.RUnlock()
	}
	for i, n := range sm.kk {
		if !fn(i, n, sm.m[n]) {
			break
		}
	}
}
