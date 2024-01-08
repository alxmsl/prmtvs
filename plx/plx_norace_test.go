package plx

import (
	. "gopkg.in/check.v1"

	"testing"
	"time"
)

const (
	testValue = "test value"
)

func Test(t *testing.T) {
	TestingT(t)
}

type PlexSuite struct{}

var (
	_ = Suite(&PlexSuite{})
)

// TestRecvOnClosedPlexus checks that Recv returns zero-value on reading closed plexus.
func (s *PlexSuite) TestRecvOnClosedPlexus(c *C) {
	var pl = NewPlexus()
	pl.Close()
	var v, ok = pl.Recv()
	c.Assert(v, IsNil)
	c.Assert(ok, Equals, false)
}

// TestRecvOnEmptyPlexus checks that Recv blocks on reading empty plexus.
func (s *PlexSuite) TestRecvOnEmptyPlexus(c *C) {
	var (
		pl  = NewPlexus()
		res bool
	)
	go func() {
		pl.Recv()
		res = true
	}()
	time.Sleep(time.Millisecond)
	c.Assert(res, Equals, false)
}

// TestSendOnClosedPlexus checks that Send panics on writing to a closed plexus.
func (s *PlexSuite) TestSendOnClosedPlexus(c *C) {
	defer func() {
		var v = recover()
		c.Assert(v, Equals, ErrorSendToClosedPlexus)
	}()
	var pl = NewPlexus()
	pl.Close()
	pl.Send(testValue)
}

// TestSendOnEmptyPlexus checks that Send blocks on writing to a plexus without readers.
func (s *PlexSuite) TestSendOnEmptyPlexus(c *C) {
	var (
		pl  = NewPlexus()
		res bool
	)
	go func() {
		pl.Send(testValue)
		res = true
	}()
	time.Sleep(time.Millisecond)
	c.Assert(res, Equals, false)
}

// TestSendOrder checks FIFO order for sender-receiver pair.
func (s *PlexSuite) TestSendOrder(c *C) {
	const count = 100
	var pl = NewPlexus()
	go func() {
		for i := 0; i < count; i += 1 {
			pl.Send(i)
		}
	}()

	for i := 0; i < count; i += 1 {
		v, ok := pl.Recv()
		c.Assert(ok, Equals, true)
		c.Assert(v, Equals, i)
	}
}

// TestRecvSend checks a plexus works in many goroutines.
func (s *PlexSuite) TestRecvSend(c *C) {
	const (
		concurrency = 5
		count       = 1000
	)
	var pl = NewPlexus()
	for i := 0; i < concurrency; i += 1 {
		go func() {
			for j := 0; j < count; j += 1 {
				pl.Send(j)
			}
		}()
	}

	var done = make(chan map[int]int)
	for i := 0; i < concurrency; i += 1 {
		go func() {
			var res = make(map[int]int)
			for j := 0; j < count; j += 1 {
				v, _ := pl.Recv()
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
