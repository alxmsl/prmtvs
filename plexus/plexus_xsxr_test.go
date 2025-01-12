package plexus_test

import (
	. "github.com/alxmsl/prmtvs/plexus"
	. "gopkg.in/check.v1"
)

const (
	simultaneousReceivers = 10 // Should not be less than 0
	simultaneousSenders   = 2  // Should not be less than 0
)

type XsXrSuite struct{}

var (
	_ = Suite(&XsXrSuite{})
)

// TestMsMr checks Plexus with multiple simultaneous senders and multiple simultaneous receivers.
func (s *XsXrSuite) TestMsMr(c *C) {
	// Create MsMr plexus.
	var plx = NewPlexus(WithReceiversNumber(simultaneousReceivers), WithSendersNumber(simultaneousSenders))
	// Receive merged value from the plexus.
	var done = make(chan Counter)
	for i := 0; i < simultaneousReceivers; i++ {
		go func(i int) {
			v, _ := recvN(plx, i)
			done <- v.(Counter)
		}(i)
	}
	// Send sequential values to the plexus.
	for i := 0; i < simultaneousSenders; i += 1 {
		go func(i int) {
			sendN(plx, i, Counter(i+1))
		}(i)
	}
	// Check result. Should be sum of numbers up to `simultaneousSenders`: n * (n + 1) / 2.
	// Result is multiplied by `simultaneousReceivers` (number of receivers).
	var res Counter
	for i := 0; i < simultaneousReceivers; i += 1 {
		res += <-done
	}
	c.Assert(res, Equals, Counter(simultaneousReceivers*simultaneousSenders*(simultaneousSenders+1)/2))
}

// TestMsSr checks Plexus with multiple simultaneous senders and single receiver.
func (s *XsXrSuite) TestMsSr(c *C) {
	// Create MsSr plexus.
	var plx = NewPlexus(WithReceiversNumber(1), WithSendersNumber(simultaneousSenders))
	// Send sequential values to the plexus.
	for i := 0; i < simultaneousSenders; i += 1 {
		go func(i int) {
			sendN(plx, i, Counter(i+1))
		}(i)
	}
	// Check result. Should be equal the sum of numbers up to `simultaneousSenders`: n * (n + 1) / 2.
	v, ok := recv0(plx)
	c.Assert(ok, Equals, true)
	c.Assert(v, Equals, Counter(simultaneousSenders*(simultaneousSenders+1)/2))
}

// TestSsMr checks Plexus with a single sender and multiple simultaneous receivers.
func (s *XsXrSuite) TestSsMr(c *C) {
	// Create SsMr plexus.
	var plx = NewPlexus(WithReceiversNumber(simultaneousReceivers), WithSendersNumber(1))
	// Receive value from the plexus.
	var done = make(chan int)
	for i := 0; i < simultaneousReceivers; i++ {
		go func(i int) {
			v, _ := recvN(plx, i)
			done <- v.(int)
		}(i)
	}
	// Send value to the plexus.
	send0(plx, 1)
	// Check result. Should be equal `simultaneousReceivers`.
	var res int
	for i := 0; i < simultaneousReceivers; i += 1 {
		res += <-done
	}
	c.Assert(res, Equals, simultaneousReceivers)
}

// TestSsSr checks Plexus with a single sender and single receiver.
func (s *XsXrSuite) TestSsSr(c *C) {
	// Create SsSr plexus.
	var plx = NewPlexus(WithReceiversNumber(1), WithSendersNumber(1))
	// Send value to the plexus.
	go func() {
		send0(plx, testValue)
	}()
	// Receive value from the plexus.
	v, ok := recv0(plx)
	c.Assert(ok, Equals, true)
	c.Assert(v, Equals, testValue)
}
