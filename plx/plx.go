package plx

import (
	"errors"
	"sync"

	"gopkg.in/eapache/queue.v1"
)

var (
	ErrorCloseClosedPlexus   = errors.New("closed the closed plexus")
	ErrorSendToClosedPlexus  = errors.New("send to the closed plexus")
	ErrorValueIsNotMergeable = errors.New("value does not implement plexus.Mergeable")
	ErrorUnknownState        = errors.New("plexus is in unknown state")
)

type Interface interface {
	// AddI increments a number of inputs for the plexus.
	AddI(int)
	// AddO increments a number of outputs for the plexus.
	AddO(int)
	// Close frees all waiting receivers, and it closes the plexus in the acquired general lock.
	Close()
	// Recv returns value from the plexus. Recv checks the plexus is not closed. Recv gets value from senders
	// in the queue. If there are not enough senders, Recv blocks and enqueues itself.
	// See MiMo, MiSo, SiMo, SiSo constants for details.
	Recv() (any, bool)
	// Send puts value into the plexus. Send checks the plexus is not closed. Send puts value to receivers from the
	// queue. If there are not enough receivers, Send blocks and enqueues itself.
	// See MiMo, MiSo, SiMo, SiSo constants for details.
	Send(any)
	// State returns the current state of the plexus. See MiMo, MiSo, SiMo, SiSo constants.
	State() int
}

// Mergeable declares value which can be merged together
type Mergeable interface {
	// Merge returns new Mergeable implementation using the given argument to merge value. The implementation
	// has to have a commutative property: a.Merge(b) must equal b.Merge(a).
	Merge(Mergeable) Mergeable
}

const (
	// MiMo has multiple inputs and multiple outputs. An input value must implement a Mergeable interface.
	// All outputs return a merged value from all inputs.
	MiMo = iota
	// MiSo has multiple inputs and a single output. An input value must implement a Mergeable interface.
	// Output returns a merged value from all inputs.
	MiSo
	// SiMo has a single input and multiple outputs. All outputs return a value from input.
	SiMo
	// SiSo has a single input and a single output. Behaviour is similar to the unbuffered channel.
	SiSo
)

type Plexus struct {
	closed bool
	lock   sync.RWMutex

	// xo is a number of outputs.
	xo int
	// xi is a number of inputs.
	xi int

	// recvq is a queue of blocked receivers.
	recvq *queue.Queue
	// sendq is a queue of blocked senders.
	sendq *queue.Queue
}

func NewPlexus() *Plexus {
	return &Plexus{
		closed: false,

		recvq: queue.New(),
		sendq: queue.New(),
	}
}

func (m *Plexus) AddI(n int) {
	m.lock.Lock()
	m.xi += n
	m.lock.Unlock()
}

func (m *Plexus) AddO(n int) {
	m.lock.Lock()
	m.xo += n
	m.lock.Unlock()
}

func (m *Plexus) Close() {
	m.lock.RLock()
	if m.closed {
		defer m.lock.RUnlock()
		panic(ErrorCloseClosedPlexus)
	}
	m.lock.RUnlock()

	m.lock.Lock()
	defer m.lock.Unlock()
	if m.closed {
		panic(ErrorCloseClosedPlexus)
	}

	m.closed = true
	if m.recvq.Length() > 0 {
		for m.recvq.Length() > 0 {
			var ch = m.recvq.Remove().(chan any)
			close(ch)
		}
	}
	if m.sendq.Length() > 0 {
		for m.sendq.Length() > 0 {
			var ch = m.sendq.Remove().(chan any)
			close(ch)
		}
	}
}

func (m *Plexus) Recv() (any, bool) {
	if m.closed {
		return nil, false
	}
	m.lock.Lock()
	if m.closed {
		m.lock.Unlock()
		return nil, false
	}

	switch m.State() {
	case SiSo:
		// If there is no waiting sender, then go ahead with block and enqueue the receiver.
		if m.sendq.Length() <= 0 {
			break
		}
		// Dequeue a sender.
		var schs = dequeue(m.sendq, 1)
		m.lock.Unlock()
		// Return value from the sender to the current receiver. Close sender.
		v, ok := <-schs[0]
		close(schs[0])
		return v, ok
	case SiMo:
		// If there are not enough waiting receivers, then go ahead with block and enqueue the receiver.
		if (m.recvq.Length() + 1) < m.xo {
			break
		}
		// If there is xo waiting sender, then go ahead with block and enqueue the receiver.
		if m.sendq.Length() <= 0 {
			break
		}
		// Dequeue a sender and receivers.
		var schs = dequeue(m.sendq, 1)
		var rchs = dequeue(m.recvq, m.xo-1)
		m.lock.Unlock()
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
	case MiSo:
		// If there is not enough senders, then go ahead with block and enqueue the receiver.
		if m.sendq.Length() < m.xi {
			break
		}
		// Dequeue senders.
		var schs = dequeue(m.sendq, m.xi)
		m.lock.Unlock()
		// Merge values from senders and return it to the current receiver.
		var res Mergeable
		res = merge(schs, res)
		return res, true
	case MiMo:
		// If there are not enough waiting receivers, then go ahead with block and enqueue the receiver.
		if (m.recvq.Length() + 1) < m.xo {
			break
		}
		// If there is not enough senders, then go ahead with block and enqueue the receiver.
		if m.sendq.Length() < m.xi {
			break
		}
		// Dequeue receivers and senders.
		var rchs = dequeue(m.recvq, m.xo-1)
		var schs = dequeue(m.sendq, m.xi)
		m.lock.Unlock()
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

	// Block and enqueue a receiver.
	var ch = make(chan any)
	m.recvq.Add(ch)
	m.lock.Unlock()
	v, ok := <-ch
	return v, ok
}

func (m *Plexus) Send(v any) {
	if m.closed {
		panic(ErrorSendToClosedPlexus)
	}
	m.lock.Lock()
	if m.closed {
		m.lock.Unlock()
		panic(ErrorSendToClosedPlexus)
	}

	switch m.State() {
	case SiSo:
		// If there is no waiting receiver, then go ahead with block and enqueue the sender.
		if m.recvq.Length() == 0 {
			break
		}
		// Dequeue a receiver.
		var rchs = dequeue(m.recvq, 1)
		m.lock.Unlock()
		// Pass value to receiver and close it.
		rchs[0] <- v
		close(rchs[0])
		return
	case SiMo:
		// If there are not enough waiting receivers, then go ahead with block and enqueue the sender.
		if m.recvq.Length() < m.xo {
			break
		}
		// Dequeue receivers.
		var rchs = dequeue(m.recvq, m.xo)
		m.lock.Unlock()
		// Pass value to receivers and close them.
		for _, rch := range rchs {
			rch <- v
			close(rch)
		}
		return
	case MiSo:
		// If there is not enough senders, then go ahead with block and enqueue the receiver.
		if (m.sendq.Length() + 1) < m.xi {
			break
		}
		// If there is no waiting receiver, then go ahead with block and enqueue the sender.
		if m.recvq.Length() <= 0 {
			break
		}
		// Value must implement the Mergeable interface to be passed.
		if _, ok := v.(Mergeable); !ok {
			panic(ErrorValueIsNotMergeable)
		}
		// Dequeue receiver and senders.
		var rchs = dequeue(m.recvq, 1)
		var schs = dequeue(m.sendq, m.xi-1)
		m.lock.Unlock()
		// Merge values from senders and pass it to receiver. Close receiver.
		var res = v.(Mergeable)
		res = merge(schs, res)
		rchs[0] <- res
		close(rchs[0])
		return
	case MiMo:
		// If there is not enough senders, then go ahead with block and enqueue the receiver.
		if (m.sendq.Length() + 1) < m.xi {
			break
		}
		// If there are not enough waiting receivers, then go ahead with block and enqueue the sender.
		if m.recvq.Length() < m.xo {
			break
		}
		// Value must implement the Mergeable interface to be passed.
		if _, ok := v.(Mergeable); !ok {
			panic(ErrorValueIsNotMergeable)
		}
		// Dequeue senders and receivers.
		var schs = dequeue(m.sendq, m.xi-1)
		var rchs = dequeue(m.recvq, m.xo)
		m.lock.Unlock()
		// Merge values from senders and pass it to receivers. Close receivers.
		var res = v.(Mergeable)
		res = merge(schs, res)
		for _, rch := range rchs {
			rch <- res
			close(rch)
		}
		return
	}

	// Block and enqueue a sender.
	var ch = make(chan any)
	m.sendq.Add(ch)
	m.lock.Unlock()
	ch <- v
}

func (m *Plexus) State() int {
	switch {
	case m.xi == 0 && m.xo == 0:
		return SiSo
	case m.xi == 0 && m.xo > 0:
		return SiMo
	case m.xi > 0 && m.xo == 0:
		return MiSo
	case m.xi > 0 && m.xo > 0:
		return MiMo
	default:
		panic(ErrorUnknownState)
	}
}

// dequeue removes `l` elements from the given queue `q`. Function is not safe. Be sure queue has enough elements
// before the call. Otherwise, dequeue may panic.
func dequeue(q *queue.Queue, l int) []chan any {
	var chs = make([]chan any, 0, l)
	for i := 0; i < l; i += 1 {
		var ch = q.Remove().(chan any)
		chs = append(chs, ch)
	}
	return chs
}

// merge returns merged result for the given slice of channels. Each element in the channels must implement Mergeable
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
