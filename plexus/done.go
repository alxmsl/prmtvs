package plexus

// doneMap represents a set of done-channels. Each done-channel has a unique name in a set.
type doneMap map[string]chan struct{}

// newDoneMap creates a doneMap object for a given capacity of a set.
func newDoneMap(cap int) doneMap {
	return make(map[string]chan struct{}, cap)
}

// add adds a done-channel with a given name.
func (rm doneMap) add(name string) {
	rm[name] = make(chan struct{})
}

// close closes all done-channels.
func (rm doneMap) close() {
	for _, ch := range rm {
		close(ch)
	}
}

// release unblocks done-channels with given names.
func (rm doneMap) release(names ...string) {
	for _, name := range names {
		rm[name] <- struct{}{}
	}
}
