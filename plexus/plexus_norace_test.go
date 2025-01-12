package plexus_test

import (
	. "github.com/alxmsl/prmtvs/plexus"
	. "gopkg.in/check.v1"

	"time"
)

const (
	testValue = "test value"
)

type PlexSuite struct{}

var (
	_ = Suite(&PlexSuite{})
)

// TestRecvOnClosedPlexus checks that Plexus.Recv returns zero-value on reading closed plexus.
func (s *PlexSuite) TestRecvOnClosedPlexus(c *C) {
	var plx = NewPlexus(WithReceiversNumber(1), WithSendersNumber(1))
	plx.Close()
	var v, ok = recv0(plx)
	c.Assert(v, IsNil)
	c.Assert(ok, Equals, false)
}

// TestRecvOnEmptyPlexus checks that Plexus.Recv blocks on reading empty plexus.
func (s *PlexSuite) TestRecvOnEmptyPlexus(c *C) {
	var (
		plx = NewPlexus(WithReceiversNumber(1), WithSendersNumber(1))
		res bool
	)
	go func() {
		recv0(plx)
		res = true
	}()
	time.Sleep(time.Millisecond)
	c.Assert(res, Equals, false)
}

// TestSendOnClosedPlexus checks that Plexus.Send panics on writing to a closed plexus.
func (s *PlexSuite) TestSendOnClosedPlexus(c *C) {
	defer func() {
		var v = recover()
		c.Assert(v, Equals, ErrorSendToClosedPlexus)
	}()
	var plx = NewPlexus(WithReceiversNumber(1), WithSendersNumber(1))
	plx.Close()
	send0(plx, testValue)
}

// TestSendOnEmptyPlexus checks that Plexus.Send blocks on writing to a plexus without readers.
func (s *PlexSuite) TestSendOnEmptyPlexus(c *C) {
	var (
		plx = NewPlexus(WithReceiversNumber(1), WithSendersNumber(1))
		res bool
	)
	go func() {
		send0(plx, testValue)
		res = true
	}()
	time.Sleep(time.Millisecond)
	c.Assert(res, Equals, false)
}

// TestSendOrder checks a FIFO order for sender-receiver pair.
func (s *PlexSuite) TestSendOrder(c *C) {
	const count = 100
	var plx = NewPlexus(WithReceiversNumber(1), WithSendersNumber(1))
	go func() {
		for i := 0; i < count; i += 1 {
			send0(plx, i)
		}
	}()

	for i := 0; i < count; i += 1 {
		v, ok := recv0(plx)
		c.Assert(ok, Equals, true)
		c.Assert(v, Equals, i)
	}
}

// TestRecvSend checks a Plexus works in many goroutines.
func (s *PlexSuite) TestRecvSend(c *C) {
	const (
		concurrency = 5
		count       = 1000
	)
	var plx = NewPlexus(WithReceiversNumber(1), WithSendersNumber(1))
	for i := 0; i < concurrency; i += 1 {
		go func() {
			for j := 0; j < count; j += 1 {
				send0(plx, j)
			}
		}()
	}

	var done = make(chan map[int]int)
	for i := 0; i < concurrency; i += 1 {
		go func() {
			var res = make(map[int]int)
			for j := 0; j < count; j += 1 {
				v, _ := recv0(plx)
				res[v.(int)] += 1
			}
			done <- res
		}()
	}

	var res = make(map[int]int)
	for i := 0; i < concurrency; i += 1 {
		for k, v := range <-done {
			res[k] += v
		}
	}

	c.Assert(res, HasLen, count)
	for _, v := range res {
		c.Assert(v, Equals, concurrency)
	}
}
