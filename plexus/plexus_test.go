package plexus_test

import (
	. "gopkg.in/check.v1"

	"fmt"
	"testing"

	"github.com/alxmsl/prmtvs/plexus"
)

func Test(t *testing.T) {
	TestingT(t)
}

// Plexer describes the plexus interface.
type Plexer interface {
	// Close frees all waiting receivers, and it closes the plexus in the acquired general lock.
	Close()
	// Recv returns value from the plexus for a given receiver (by name). Recv checks the plexus is not closed.
	// Recv gets value from senders (from a queues). If there are not enough senders, Recv blocks and enqueues itself.
	// See MsMr, MsSr, SsMr, SsSr constants for details.
	Recv(name string) (any, bool)
	// Send puts value into the plexus from a given sender (by name). Send checks the plexus is not closed.
	// Send puts value to receivers (into the queues). If there are not enough receivers, it blocks and enqueues itself.
	// See MsMr, MsSr, SsMr, SsSr constants for details.
	Send(name string, value any)
	// State returns the current state of the plexus. See MsMr, MsSr, SsMr, SsSr constants.
	State() int
}

var _ Plexer = (*plexus.Plexus)(nil)

func recv0(plx *plexus.Plexus) (any, bool) {
	return plx.Recv("receiver_0")
}

func recvN(plx *plexus.Plexus, n int) (any, bool) {
	return plx.Recv(fmt.Sprintf("receiver_%d", n))
}

func send0(plx *plexus.Plexus, value any) {
	plx.Send("sender_0", value)
}

func sendN(plx *plexus.Plexus, n int, value any) {
	plx.Send(fmt.Sprintf("sender_%d", n), value)
}
