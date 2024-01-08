package plx

import (
	. "gopkg.in/check.v1"
)

const (
	factorI = 2  // Should not be less than 0.
	factorO = 10 // Should not be less than 0.
)

type mergeableInt int

func (a mergeableInt) Merge(b Mergeable) Mergeable {
	if _, ok := b.(mergeableInt); !ok {
		panic("value does not implement plexus.mergeableInt")
	}
	return a + b.(mergeableInt)
}

type XiXoSuite struct{}

var (
	_ = Suite(&XiXoSuite{})
)

// TestMiMo checks multiple-input and multiple-output plexus.
func (s *XiXoSuite) TestMiMo(c *C) {
	// Create MiMo plexus.
	var pl = NewPlexus()
	pl.AddO(factorO)
	pl.AddI(factorI)
	// Receive merged value from the plexus.
	var done = make(chan mergeableInt)
	for i := 0; i < factorO; i++ {
		go func() {
			v, _ := pl.Recv()
			done <- v.(mergeableInt)
		}()
	}
	// Send sequential values to the plexus.
	for i := 0; i < factorI; i += 1 {
		go func(i int) { pl.Send(mergeableInt(i + 1)) }(i)
	}
	// Check result. Should be sum of numbers up to `factorI`: n * (n + 1) / 2.
	// Result is multiplied by `factorO` (number of receivers).
	var res mergeableInt
	for i := 0; i < factorO; i += 1 {
		res += <-done
	}
	c.Assert(res, Equals, mergeableInt(factorO*factorI*(factorI+1)/2))
}

// TestMiSo checks multiple-input and single-output plexus.
func (s *XiXoSuite) TestMiSo(c *C) {
	// Create MiSo plexus.
	var pl = NewPlexus()
	pl.AddI(factorI)
	// Send sequential values to the plexus.
	for i := 0; i < factorI; i += 1 {
		go func(i int) { pl.Send(mergeableInt(i + 1)) }(i)
	}
	// Check result. Should be equal the sum of numbers up to `factorI`: n * (n + 1) / 2.
	v, ok := pl.Recv()
	c.Assert(ok, Equals, true)
	c.Assert(v, Equals, mergeableInt(factorI*(factorI+1)/2))
}

// TestSiMo checks single-input and multiple-output plexus.
func (s *XiXoSuite) TestSiMo(c *C) {
	// Create SiMo plexus.
	var pl = NewPlexus()
	pl.AddO(factorO)
	// Receive value from the plexus.
	var done = make(chan int)
	for i := 0; i < factorO; i++ {
		go func() {
			v, _ := pl.Recv()
			done <- v.(int)
		}()
	}
	// Send value to the plexus
	pl.Send(1)
	// Check result. Should be equal `factorO`.
	var res int
	for i := 0; i < factorO; i += 1 {
		res += <-done
	}
	c.Assert(res, Equals, factorO)
}

// TestSiSo checks single-input and single-output plexus.
func (s *XiXoSuite) TestSiSo(c *C) {
	// Create SiSo plexus.
	var pl = NewPlexus()
	// Send value to the plexus
	go func() { pl.Send(testValue) }()
	// Receive value from the plexus.
	v, ok := pl.Recv()
	c.Assert(ok, Equals, true)
	c.Assert(v, Equals, testValue)
}
