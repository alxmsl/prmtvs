package plexus

// Mergeable declares value which can be merged together.
type Mergeable interface {
	// Merge returns new Mergeable implementation using the given argument to merge value.
	// The implementation has to have a commutative property: a.Merge(b) must equal b.Merge(a).
	// Details: https://en.wikipedia.org/wiki/Commutative_property
	Merge(Mergeable) Mergeable
}

// merge returns merged result for the given slice of channels. Element in each channel must implement Mergeable
// interface. Otherwise, function panics.
func merge(chs []chan any, res Mergeable) Mergeable {
	for _, ch := range chs {
		if res == nil {
			v := <-ch
			if _, ok := v.(Mergeable); !ok {
				panic(ErrorValueIsNotMergeable)
			}
			res = v.(Mergeable)
		} else {
			res = res.Merge((<-ch).(Mergeable))
		}
		close(ch)
	}
	return res
}
