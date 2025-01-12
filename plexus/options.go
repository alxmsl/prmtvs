package plexus

import (
	"fmt"
)

// Option represents an abstract option with is allowed to be set for a Plexus.
type Option func(*Plexus)

// WithName defines a name for a Plexus.
func WithName(name string) Option {
	return func(plx *Plexus) {
		plx.lock.Lock()
		defer plx.lock.Unlock()

		plx.name = name
	}
}

// WithReceivers defines a set of names for receivers of a Plexus.
func WithReceivers(names ...string) Option {
	return func(plx *Plexus) {
		plx.lock.Lock()
		defer plx.lock.Unlock()

		plx.recvn = len(names)
		plx.recvq = newQueues(plx.recvn)
		for _, name := range names {
			plx.recvq.add(name)
		}
	}
}

// WithReceiversNumber defines a total number of receivers for a Plexus.
// Receiver name is assigned automatically in a sequential order with a prefix `receiver_`.
// E.g. `receiver_0`, `receiver_1` etc.
func WithReceiversNumber(n int) Option {
	return func(plx *Plexus) {
		var names = make([]string, 0, n)
		for i := 0; i < n; i += 1 {
			names = append(names, fmt.Sprintf("receiver_%d", i))
		}
		WithReceivers(names...)(plx)
	}
}

// WithSelectableSenders enabled a selectable senders functionality for a Plexus.
func WithSelectableSenders() Option {
	return func(plx *Plexus) {
		plx.lock.Lock()
		defer plx.lock.Unlock()

		plx.selectableSenders = true
		plx.sendr = newDoneMap(plx.sendn)
		for name := range plx.sendq.qm {
			plx.sendr.add(name)
		}
	}
}

// WithSenders defines a set of names for senders of a Plexus.
func WithSenders(names ...string) Option {
	return func(plx *Plexus) {
		plx.lock.Lock()
		defer plx.lock.Unlock()

		// Define senders.
		plx.sendn = len(names)
		plx.sendq = newQueues(plx.sendn)
		for _, name := range names {
			plx.sendq.add(name)
		}
	}
}

// WithSendersNumber defines a total number of senders for a Plexus.
// Sender name is assigned automatically in a sequential order with a prefix `sender_`.
// E.g. `sender_0`, `sender_1` etc.
func WithSendersNumber(n int) Option {
	return func(plx *Plexus) {
		var names = make([]string, 0, n)
		for i := 0; i < n; i += 1 {
			names = append(names, fmt.Sprintf("sender_%d", i))
		}
		WithSenders(names...)(plx)
	}
}
