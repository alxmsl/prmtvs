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
	// AddReceiver increments a number of simultaneous receivers for the plexus
	AddReceiver(int)
	// AddSender increments a number of simultaneous senders for the plexus
	AddSender(int)
	// Close frees all waiting receivers, and it closes the plexus in the acquired general lock
	Close()
	// Recv returns value from the plexus. Recv checks the plexus is not closed. Recv gets value from senders
	// in the queue. If there are not enough senders, Recv blocks and enqueues itself.
	// See MsMr, MsSr, SsMr, SsSr constants for details
	Recv() (any, bool)
	// Send puts value into the plexus. Send checks the plexus is not closed. Send puts value to receivers from the
	// queue. If there are not enough receivers, Send blocks and enqueues itself.
	// See MsMr, MsSr, SsMr, SsSr constants for details
	Send(any)
	// State returns the current state of the plexus. See MsMr, MsSr, SsMr, SsSr constants
	State() int
}

// Mergeable declares value which can be merged together
type Mergeable interface {
	// Merge returns new Mergeable implementation using the given argument to merge value. The implementation
	// has to have a commutative property: a.Merge(b) must equal b.Merge(a)
	Merge(Mergeable) Mergeable
}

const (
	// MsMr plexus has multiple simultaneous senders and multiple simultaneous receivers. Value must implement
	// a Mergeable interface. All receivers take a merged value from all senders
	MsMr = iota
	// MsSr has multiple simultaneous senders and a single receiver. Value must implement a Mergeable interface.
	// Receiver takes a merged value from all senders
	MsSr
	// SsMr has a single sender and multiple simultaneous receivers. All receivers take a value from sender
	SsMr
	// SsSr has a single sender and a single receiver. Behaviour is similar to the unbuffered channel
	SsSr
)

type Plexus struct {
	closed bool
	lock   sync.RWMutex

	// recvn is a number of simultaneous receivers
	recvn int
	// recvr is ready-channel for the select statement on Plexus.Recv operations
	recvr chan struct{}
	// recvq is a queue of blocked receivers
	recvq *queue.Queue

	// sendn is a number of simultaneous senders
	sendn int
	// sendq is a queue of blocked senders
	sendq *queue.Queue
}

// NewPlexus creates a plexus with an initial number of simultaneous receivers and senders
func NewPlexus(recvn, sendn int) Plexus {
	return Plexus{
		closed: false,

		recvn: recvn,
		recvq: queue.New(),

		sendn: sendn,
		sendq: queue.New(),

		recvr: make(chan struct{}),
	}
}

func (m *Plexus) AddSender(n int) {
	m.lock.Lock()
	m.sendn += n
	m.lock.Unlock()
}

func (m *Plexus) AddReceiver(n int) {
	m.lock.Lock()
	m.recvn += n
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
	// If there are not enough waiting receiver(s) or sender(s), then go ahead with block and enqueue the receiver
	if (m.recvq.Length()+1) < m.recvn || m.sendq.Length() < m.sendn {
		// Enqueue a receiver
		var ch = make(chan any)
		m.recvq.Add(ch)
		m.lock.Unlock()
		// Block the execution till a sender
		v, ok := <-ch
		return v, ok
	}

	switch m.State() {
	case SsSr:
		// Dequeue a sender.
		var schs = dequeue(m.sendq, 1)
		m.lock.Unlock()
		// Return value from the sender to the current receiver. Close sender
		v, ok := <-schs[0]
		close(schs[0])
		return v, ok
	case SsMr:
		// Dequeue a sender and receivers
		var schs = dequeue(m.sendq, 1)
		var rchs = dequeue(m.recvq, m.recvn-1)
		m.lock.Unlock()
		// Pass value from sender to receivers. Close sender and receivers
		v, ok := <-schs[0]
		close(schs[0])
		for _, ch := range rchs {
			if ok {
				ch <- v
			}
			close(ch)
		}
		// Return value from the sender to the current receiver
		return v, ok
	case MsSr:
		// Dequeue senders
		var schs = dequeue(m.sendq, m.sendn)
		m.lock.Unlock()
		// Merge values from senders and return it to the current receiver
		var res Mergeable
		res = merge(schs, res)
		return res, true
	case MsMr:
		fallthrough
	default:
		// Dequeue receivers and senders
		var rchs = dequeue(m.recvq, m.recvn-1)
		var schs = dequeue(m.sendq, m.sendn)
		m.lock.Unlock()
		// Merge values from senders and pass it to receivers. Close receivers
		var res Mergeable
		res = merge(schs, res)
		for _, ch := range rchs {
			ch <- res
			close(ch)
		}
		// Return the merged value to the current receiver
		return res, true
	}
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
	// If there is not enough sender(s) or no waiting receiver(s), then go ahead with block and enqueue the sender
	if (m.sendq.Length()+1) < m.sendn || m.recvq.Length() < m.recvn {
		// Enqueue a sender
		var ch = make(chan any)
		m.sendq.Add(ch)
		m.lock.Unlock()
		// Block the execution till a receiver
		ch <- v
		return
	}

	switch m.State() {
	case SsSr:
		// Dequeue a receiver
		var rchs = dequeue(m.recvq, 1)
		m.lock.Unlock()
		// Pass value to receiver and close it
		rchs[0] <- v
		close(rchs[0])
	case SsMr:
		// Dequeue receivers
		var rchs = dequeue(m.recvq, m.recvn)
		m.lock.Unlock()
		// Pass value to receivers and close them
		for _, rch := range rchs {
			rch <- v
			close(rch)
		}
	case MsSr:
		// Value must implement the Mergeable interface to be passed
		if _, ok := v.(Mergeable); !ok {
			panic(ErrorValueIsNotMergeable)
		}
		// Dequeue receiver and senders
		var rchs = dequeue(m.recvq, 1)
		var schs = dequeue(m.sendq, m.sendn-1)
		m.lock.Unlock()
		// Merge values from senders and pass it to receiver. Close receiver
		var res = v.(Mergeable)
		res = merge(schs, res)
		rchs[0] <- res
		close(rchs[0])
	case MsMr:
		// Value must implement the Mergeable interface to be passed
		if _, ok := v.(Mergeable); !ok {
			panic(ErrorValueIsNotMergeable)
		}
		// Dequeue senders and receivers
		var schs = dequeue(m.sendq, m.sendn-1)
		var rchs = dequeue(m.recvq, m.recvn)
		m.lock.Unlock()
		// Merge values from senders and pass it to receivers. Close receivers
		var res = v.(Mergeable)
		res = merge(schs, res)
		for _, rch := range rchs {
			rch <- res
			close(rch)
		}
	}
}

func (m *Plexus) ReadyReceive() chan struct{} {
	return m.recvr
}

func (m *Plexus) State() int {
	switch {
	case m.sendn == 1 && m.recvn == 1:
		return SsSr
	case m.sendn == 1 && m.recvn > 1:
		return SsMr
	case m.sendn > 1 && m.recvn == 1:
		return MsSr
	case m.sendn > 1 && m.recvn > 1:
		return MsMr
	default:
		panic(ErrorUnknownState)
	}
}

// dequeue removes `l` elements from the given queue `q`. Function is not safe. Be sure queue has enough elements
// before the call. Otherwise, dequeue may panic
func dequeue(q *queue.Queue, l int) []chan any {
	var chs = make([]chan any, 0, l)
	for i := 0; i < l; i += 1 {
		var ch = q.Remove().(chan any)
		chs = append(chs, ch)
	}
	return chs
}

// merge returns merged result for the given slice of channels. Each element in the channels must implement Mergeable
// interface. Otherwise, function panics
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
