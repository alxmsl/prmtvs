package skm

// OverFunc represents an item processor to go through the SKM.
// Function accepts index, key and value.
// OverFunc returns bool value to continue or stop the loop.
type OverFunc func(idx int, key string, value interface{}) bool

// SKM struct represents a map with a sorted set of keys.
type SKM struct {
	mm map[string]interface{}
	kk []string
}

// NewSortedKeyMap creates a SKM object.
func NewSortedKeyMap() *SKM {
	return &SKM{
		mm: map[string]interface{}{},
		kk: []string{},
	}
}

func (sm *SKM) Add(key string, value interface{}) bool {
	if _, ok := sm.mm[key]; ok {
		return false
	}

	var idx = 0
	for idx = 0; idx < len(sm.kk); idx += 1 {
		if key < sm.kk[idx] {
			break
		}
	}
	sm.kk = append(sm.kk[:idx], append([]string{key}, sm.kk[idx:]...)...)
	sm.mm[key] = value
	return true
}

func (sm *SKM) ExistsIndex(idx int) bool {
	return idx >= 0 && idx < len(sm.kk)
}

func (sm *SKM) ExistsKey(key string) bool {
	_, ok := sm.mm[key]
	return ok
}

func (sm *SKM) GetByIndex(idx int) (interface{}, bool) {
	if idx < 0 || idx >= len(sm.kk) {
		return nil, false
	}
	v, ok := sm.mm[sm.kk[idx]]
	return v, ok
}

func (sm *SKM) GetByKey(key string) (interface{}, bool) {
	p, ok := sm.mm[key]
	return p, ok
}

func (sm *SKM) Index(key string) (int, bool) {
	for idx, k := range sm.kk {
		if key == k {
			return idx, true
		}
	}
	return 0, false
}

func (sm *SKM) Key(idx int) (string, bool) {
	if idx < 0 || idx >= len(sm.kk) {
		return "", false
	}
	return sm.kk[idx], true
}

func (sm *SKM) Len() int {
	return len(sm.kk)
}

func (sm *SKM) Over(fn OverFunc) {
	for idx, key := range sm.kk {
		if !fn(idx, key, sm.mm[key]) {
			break
		}
	}
}

func (sm *SKM) Reset() {
	sm.mm = map[string]interface{}{}
	sm.kk = []string{}
}
