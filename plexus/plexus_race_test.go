//go:build !race

package plexus_test

import (
	. "github.com/alxmsl/prmtvs/plexus"
	. "gopkg.in/check.v1"

	"time"
)

// TestCloseUnblocksRecv checks that Plexus.Close unblocks a waiting Plexus.Recv.
// The test contains a race because it closes Plexus on the receiving operation.
func (s *PlexSuite) TestCloseUnblocksRecv(c *C) {
	var (
		plx  = NewPlexus(WithReceiversNumber(1), WithSendersNumber(1))
		done = make(chan bool)
	)
	go func() {
		v, ok := recv0(plx)
		done <- v == nil && !ok
	}()
	time.Sleep(time.Millisecond)
	plx.Close()
	c.Assert(<-done, Equals, true)
}

// TestSendPanicsOnClose checks that a blocked Plexus.Send panics on Plexus.Close.
// The test contains a race because it closes Plexus on the sending operation.
func (s *PlexSuite) TestSendPanicsOnClose(c *C) {
	var (
		plx  = NewPlexus(WithReceiversNumber(1), WithSendersNumber(1))
		done = make(chan interface{})
	)
	go func() {
		defer func() {
			done <- recover()
		}()
		send0(plx, testValue)
	}()
	time.Sleep(time.Millisecond)
	plx.Close()
	var v, ok = <-done
	c.Assert(ok, Equals, true)
	c.Assert(v.(error), NotNil)
	c.Assert(v.(error).Error(), Equals, "send on closed channel")
}
