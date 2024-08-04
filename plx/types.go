package plx

// Counter struct implements Mergeable interface for integer values
type Counter int

func (a Counter) Merge(b Mergeable) Mergeable {
	if _, ok := b.(Counter); !ok {
		panic("value does not implement Counter")
	}
	return a + b.(Counter)
}
