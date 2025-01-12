package plexus

import (
	"fmt"
	"sync"

	"gopkg.in/eapache/queue.v1"
)

// queues struct represents a named set of queue of the fixed capacity.
// Each item in the queue is a channel.
type queues struct {
	cap  int
	lock sync.Mutex
	qm   map[string]*queue.Queue
}

// newQueues creates a queues object with a given capacity.
func newQueues(cap int) *queues {
	return &queues{
		cap:  cap,
		lock: sync.Mutex{},
		qm:   make(map[string]*queue.Queue, cap),
	}
}

// add adds queue with a given name. If name is occupied already, then function panics.
// If capacity is full, then function panics.
func (qm *queues) add(name string) {
	qm.lock.Lock()
	defer qm.lock.Unlock()
	if _, ok := qm.qm[name]; ok {
		panic(fmt.Errorf("can not add queue '%s': %w", name, ErrorQueueAlreadyExists))
	}
	if len(qm.qm) >= qm.cap {
		panic(ErrorQueuesIsFull)
	}
	qm.qm[name] = queue.New()
}

// close closes all channels stored in queues.
func (qm *queues) close() {
	for _, q := range qm.qm {
		for q.Length() > 0 {
			var ch = q.Remove().(chan any)
			close(ch)
		}
	}
}

// dequeue returns a subset of channels. Subset contains one channel from each named queue.
func (qm *queues) dequeue() []chan any {
	if len(qm.qm) != qm.cap {
		panic(ErrorQueuesIsNotDefined)
	}
	var result = make([]chan any, 0, qm.cap)
	for _, q := range qm.qm {
		var ch = q.Remove().(chan any)
		result = append(result, ch)
	}
	return result
}

// dequeue returns a subset of channels.
// Subset contains one channel from each named queue except the given name.
func (qm *queues) dequeueExcept(name string) []chan any {
	if len(qm.qm) != qm.cap {
		panic(ErrorQueuesIsNotDefined)
	}
	var result = make([]chan any, 0, qm.cap-1)
	for k, q := range qm.qm {
		if name == k {
			continue
		}
		var ch = q.Remove().(chan any)
		result = append(result, ch)
	}
	return result
}

// enqueue adds a given channel into a queue with a given name.
func (qm *queues) enqueue(name string, ch chan any) {
	if _, ok := qm.qm[name]; !ok {
		panic(fmt.Errorf("can not add channel to '%s': %w", name, ErrorQueueDoesNotExist))
	}
	qm.qm[name].Add(ch)
}

// occupancy returns number of queue contains at least one channel.
func (qm *queues) occupancy() int {
	var result int
	for _, q := range qm.qm {
		if q.Length() > 0 {
			result += 1
		}
	}
	return result
}

// occupancy returns number of queue contains at least one channel. Queue with a given name is ignored.
func (qm *queues) occupancyExcept(name string) int {
	var result int
	for k, q := range qm.qm {
		if name == k {
			continue
		}
		if q.Length() > 0 {
			result += 1
		}
	}
	return result
}
