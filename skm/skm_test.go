package skm

import (
	. "gopkg.in/check.v1"

	"testing"
)

func Test(t *testing.T) {
	TestingT(t)
}

type SMSuite struct{}

var _ = Suite(&SMSuite{})

func (s *SMSuite) TestInterface(c *C) {
	ds := []struct {
		string
		int
	}{
		{"c", 0},
		{"bb", 1},
		{"aaa", 2},
		{"ccc", 3},
		{"bb", 4},
		{"a", 5},
	}

	sm := NewSKM()

	// Add test
	for _, d := range ds {
		sm.Add(d.string, d)
	}

	// ExistsIndex
	c.Assert(sm.kk, HasLen, 5)
	c.Assert(sm.ExistsIndex(-1), Equals, false)
	c.Assert(sm.ExistsIndex(0), Equals, true)
	c.Assert(sm.ExistsIndex(1), Equals, true)
	c.Assert(sm.ExistsIndex(2), Equals, true)
	c.Assert(sm.ExistsIndex(3), Equals, true)
	c.Assert(sm.ExistsIndex(4), Equals, true)
	c.Assert(sm.ExistsIndex(5), Equals, false)

	// ExistsKey
	c.Assert(sm.m, HasLen, 5)
	c.Assert(sm.ExistsKey(""), Equals, false)
	c.Assert(sm.ExistsKey("a"), Equals, true)
	c.Assert(sm.ExistsKey("aaa"), Equals, true)
	c.Assert(sm.ExistsKey("bb"), Equals, true)
	c.Assert(sm.ExistsKey("c"), Equals, true)
	c.Assert(sm.ExistsKey("ccc"), Equals, true)
	c.Assert(sm.ExistsKey("cccc"), Equals, false)

	// GetByKey
	c.Assert(sm.m, HasLen, 5)
	smp, ok := sm.GetByKey("c")
	c.Assert(ok, Equals, true)
	c.Assert(smp, Equals, ds[0])
	smp, ok = sm.GetByKey("bb")
	c.Assert(ok, Equals, true)
	c.Assert(smp, Equals, ds[1])
	smp, ok = sm.GetByKey("aaa")
	c.Assert(ok, Equals, true)
	c.Assert(smp, Equals, ds[2])
	smp, ok = sm.GetByKey("ccc")
	c.Assert(ok, Equals, true)
	c.Assert(smp, Equals, ds[3])
	smp, ok = sm.GetByKey("a")
	c.Assert(ok, Equals, true)
	c.Assert(smp, Equals, ds[5])

	// GetByIndex
	smp, ok = sm.GetByIndex(0)
	c.Assert(ok, Equals, true)
	c.Assert(smp, Equals, ds[5])
	smp, ok = sm.GetByIndex(1)
	c.Assert(ok, Equals, true)
	c.Assert(smp, Equals, ds[2])
	smp, ok = sm.GetByIndex(2)
	c.Assert(ok, Equals, true)
	c.Assert(smp, Equals, ds[1])
	smp, ok = sm.GetByIndex(3)
	c.Assert(ok, Equals, true)
	c.Assert(smp, Equals, ds[0])
	smp, ok = sm.GetByIndex(4)
	c.Assert(ok, Equals, true)
	c.Assert(smp, Equals, ds[3])

	// Key
	c.Assert(sm.Key(0), Equals, "a")
	c.Assert(sm.Key(1), Equals, "aaa")
	c.Assert(sm.Key(2), Equals, "bb")
	c.Assert(sm.Key(3), Equals, "c")
	c.Assert(sm.Key(4), Equals, "ccc")

	// Index
	c.Assert(sm.Index("a"), Equals, 0)
	c.Assert(sm.Index("aaa"), Equals, 1)
	c.Assert(sm.Index("bb"), Equals, 2)
	c.Assert(sm.Index("c"), Equals, 3)
	c.Assert(sm.Index("ccc"), Equals, 4)
	c.Assert(sm.Index(""), Equals, -1)
}
