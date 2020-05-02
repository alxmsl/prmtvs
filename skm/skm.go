package skm

import (
	"strings"
)

type SKM struct {
	m  map[string]interface{}
	kk []string
}

func NewSKM() *SKM {
	return &SKM{
		m:  map[string]interface{}{},
		kk: []string{},
	}
}

func (sm *SKM) Add(k string, v interface{}) bool {
	if _, ok := sm.m[k]; ok {
		return !ok
	}

	i := 0
	for i = 0; i < len(sm.kk); i += 1 {
		if strings.Compare(k, sm.kk[i]) == -1 {
			break
		}
	}
	sm.kk = append(sm.kk[:i], append([]string{k}, sm.kk[i:]...)...)
	sm.m[k] = v
	return true
}

func (sm *SKM) ExistsIndex(i int) bool {
	return i >= 0 && i < len(sm.kk)
}

func (sm *SKM) ExistsKey(k string) bool {
	_, ok := sm.m[k]
	return ok
}

func (sm *SKM) GetByKey(k string) (interface{}, bool) {
	p, ok := sm.m[k]
	return p, ok
}

func (sm *SKM) GetByIndex(i int) (interface{}, bool) {
	v, ok := sm.m[sm.kk[i]]
	return v, ok
}

func (sm *SKM) Index(k string) int {
	for i, n := range sm.kk {
		if k == n {
			return i
		}
	}
	return -1
}

func (sm *SKM) Key(i int) string {
	return sm.kk[i]
}

func (sm *SKM) Len() int {
	return len(sm.kk)
}

func (sm *SKM) Over(fn func(int, string, interface{}) bool) {
	for i, n := range sm.kk {
		if !fn(i, n, sm.m[n]) {
			break
		}
	}
}
