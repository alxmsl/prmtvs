package plexus

import (
	"sync"
)

const (
	// MsMr plexus has multiple simultaneous senders and multiple simultaneous receivers. Value must implement
	// a Mergeable interface. All receivers take a merged value.
	MsMr = iota
	// MsSr has multiple simultaneous senders and a single receiver. Value must implement a Mergeable interface.
	// Receiver takes a merged value.
	MsSr
	// SsMr has a single sender and multiple simultaneous receivers. All receivers take a value from sender.
	SsMr
	// SsSr has a single sender and a single receiver. Behaviour is similar to the unbuffered channel.
	SsSr
)

// Plexus struct represents a multiplexed channel. Multiplexed channel awaits data on all senders and passes it to all
// receivers.
type Plexus struct {
	lock sync.RWMutex

	active bool
	closed bool

	recvn int     // recvn is a number of simultaneous receivers.
	recvq *queues // recvq is a named queues of blocked receivers.

	sendn int     // sendn is a number of simultaneous senders.
	sendq *queues // sendq is a named queues of blocked senders.
	sendr doneMap // sendr is a named set of ready-channels for the select statement on Plexus.Send operations.

	name              string // name is just a name of the Plexus object.
	selectableSenders bool   // selectableSenders defines that Plexus is used via select-statement.
}

// NewPlexus creates a Plexus object with a required set of Option.
func NewPlexus(options ...Option) *Plexus {
	var plx = &Plexus{
		active: false,
		closed: false,
	}
	for _, opt := range options {
		opt(plx)
	}
	plx.validate()
	return plx
}

func (plx *Plexus) validate() {
	if plx.sendq.cap != plx.sendn {
		panic(ErrorUnknownState)
	}
	if plx.selectableSenders && len(plx.sendr) != plx.sendn {
		panic(ErrorUnknownState)
	}
}

func (plx *Plexus) Close() {
	plx.lock.RLock()
	if plx.closed {
		defer plx.lock.RUnlock()
		panic(ErrorCloseClosedPlexus)
	}
	plx.lock.RUnlock()

	plx.lock.Lock()
	defer plx.lock.Unlock()
	if plx.closed {
		panic(ErrorCloseClosedPlexus)
	}

	plx.recvq.close()
	plx.sendq.close()
	plx.sendr.close()
	plx.closed = true
}

func (plx *Plexus) Recv(name string) (any, bool) {
	if plx.closed {
		return nil, false
	}
	plx.lock.Lock()
	if plx.closed {
		plx.lock.Unlock()
		return nil, false
	}
	if !plx.active {
		plx.active = true
	}
	// If there are not enough waiting receiver(s) or sender(s), then go ahead with block and enqueue the receiver.
	if plx.recvq.occupancyExcept(name)+1 < plx.recvn || plx.sendq.occupancy() < plx.sendn {
		// Enqueue a receiver.
		var ch = make(chan any)
		plx.recvq.enqueue(name, ch)

		// In case of selectable mode, release all receivers, if they are waiting.
		if plx.selectableSenders && plx.recvq.occupancy() == plx.recvn {
			for name := range plx.sendq.qm {
				plx.sendr[name] <- struct{}{}
			}
		}

		plx.lock.Unlock()
		// Block the execution till a sender.
		v, ok := <-ch
		return v, ok
	}

	// In case of selectable mode, release all receivers.
	if plx.selectableSenders {
		for name := range plx.sendq.qm {
			plx.sendr[name] <- struct{}{}
		}
	}

	switch plx.State() {
	case SsSr:
		// Dequeue a sender.
		var schs = plx.sendq.dequeue()
		plx.lock.Unlock()
		// Return value from the sender to the current receiver. Close sender.
		v, ok := <-schs[0]
		close(schs[0])
		return v, ok
	case SsMr:
		// Dequeue a sender and receivers.
		var schs = plx.sendq.dequeue()
		var rchs = plx.recvq.dequeueExcept(name)
		plx.lock.Unlock()
		// Pass value from sender to receivers. Close sender and receivers.
		v, ok := <-schs[0]
		close(schs[0])
		for _, ch := range rchs {
			if ok {
				ch <- v
			}
			close(ch)
		}
		// Return value from the sender to the current receiver.
		return v, ok
	case MsSr:
		// Dequeue senders.
		var schs = plx.sendq.dequeue()
		plx.lock.Unlock()
		// Merge values from senders and return it to the current receiver.
		var res Mergeable
		res = merge(schs, res)
		return res, true
	case MsMr:
		fallthrough
	default:
		// Dequeue receivers and senders.
		var rchs = plx.recvq.dequeueExcept(name)
		var schs = plx.sendq.dequeue()
		plx.lock.Unlock()
		// Merge values from senders and pass it to receivers. Close receivers.
		var res Mergeable
		res = merge(schs, res)
		for _, ch := range rchs {
			ch <- res
			close(ch)
		}
		// Return the merged value to the current receiver.
		return res, true
	}
}

func (plx *Plexus) Send(name string, value any) {
	if plx.closed {
		panic(ErrorSendToClosedPlexus)
	}
	plx.lock.Lock()
	if plx.closed {
		plx.lock.Unlock()
		panic(ErrorSendToClosedPlexus)
	}
	if !plx.active {
		plx.active = true
	}

	// If there is not enough sender(s) or no waiting receiver(s), then go ahead with block and enqueue the sender.
	if plx.sendq.occupancyExcept(name)+1 < plx.sendn || plx.recvq.occupancy() < plx.recvn {
		// Enqueue a sender.
		var ch = make(chan any)
		plx.sendq.enqueue(name, ch)

		plx.lock.Unlock()
		// Block the execution till a receiver.
		ch <- value
		return
	}

	switch plx.State() {
	case SsSr:
		// Dequeue a receiver.
		var rchs = plx.recvq.dequeue()
		plx.lock.Unlock()
		// Pass value to receiver and close it.
		rchs[0] <- value
		close(rchs[0])
	case SsMr:
		// Dequeue receivers.
		var rchs = plx.recvq.dequeue()
		plx.lock.Unlock()
		// Pass value to receivers and close them.
		for _, rch := range rchs {
			rch <- value
			close(rch)
		}
	case MsSr:
		// Value must implement the Mergeable interface to be passed.
		if _, ok := value.(Mergeable); !ok {
			panic(ErrorValueIsNotMergeable)
		}
		// Dequeue receiver and senders.
		var rchs = plx.recvq.dequeue()
		var schs = plx.sendq.dequeueExcept(name)
		plx.lock.Unlock()
		// Merge values from senders and pass it to receiver. Close receiver.
		var res = value.(Mergeable)
		res = merge(schs, res)
		rchs[0] <- res
		close(rchs[0])
	case MsMr:
		// Value must implement the Mergeable interface to be passed.
		if _, ok := value.(Mergeable); !ok {
			panic(ErrorValueIsNotMergeable)
		}
		// Dequeue senders and receivers.
		var schs = plx.sendq.dequeueExcept(name)
		var rchs = plx.recvq.dequeue()
		plx.lock.Unlock()
		// Merge values from senders and pass it to receivers. Close receivers.
		var res = value.(Mergeable)
		res = merge(schs, res)
		for _, rch := range rchs {
			rch <- res
			close(rch)
		}
	}
}

func (plx *Plexus) ReadySend(name string) <-chan struct{} {
	if !plx.selectableSenders {
		panic(ErrorNotSelectable)
	}
	return plx.sendr[name]
}

func (plx *Plexus) State() int {
	switch {
	case plx.sendn == 1 && plx.recvn == 1:
		return SsSr
	case plx.sendn == 1 && plx.recvn > 1:
		return SsMr
	case plx.sendn > 1 && plx.recvn == 1:
		return MsSr
	case plx.sendn > 1 && plx.recvn > 1:
		return MsMr
	default:
		panic(ErrorUnknownState)
	}
}
