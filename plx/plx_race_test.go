//go:build !race

package plx

import (
	. "gopkg.in/check.v1"

	"time"
)

// TestCloseUnblocksRecv checks that Close unblocks a waiting Recv.
// The test contains a race because it closes plexus on the receiving operation.
func (s *PlexSuite) TestCloseUnblocksRecv(c *C) {
	var (
		pl   = NewPlexus()
		done = make(chan bool)
	)
	go func() {
		v, ok := pl.Recv()
		done <- v == nil && !ok
	}()
	time.Sleep(time.Millisecond)
	pl.Close()
	c.Assert(<-done, Equals, true)
}

// TestSendPanicsOnClose checks that a blocked Send panics on Close.
// The test contains a race because it closes plexus on the sending operation.
func (s *PlexSuite) TestSendPanicsOnClose(c *C) {
	var (
		pl   = NewPlexus()
		done = make(chan interface{})
	)
	go func() {
		defer func() {
			done <- recover()
		}()
		pl.Send(testValue)
	}()
	time.Sleep(time.Millisecond)
	pl.Close()
	var v, ok = <-done
	c.Assert(ok, Equals, true)
	c.Assert(v.(error), NotNil)
	c.Assert(v.(error).Error(), Equals, "send on closed channel")
}
