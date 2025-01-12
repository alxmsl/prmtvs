package skm

import (
	. "gopkg.in/check.v1"

	"testing"
)

type SortedKeyMap interface {
	// Add adds new value into the SortedKeyMap, if it is not presented already.
	// If a given key is already exists, then returns FALSE.
	Add(key string, value interface{}) bool
	// ExistsIndex checks that a given index is presented in the SortedKeyMap.
	ExistsIndex(idx int) bool
	// ExistsKey checks that a given key is presented in the SortedKeyMap.
	ExistsKey(key string) bool
	// GetByIndex returns a value in the SortedKeyMap for a given index.
	// If there is no such index in the SortedKeyMap, then function returns FALSE as a second return value.
	GetByIndex(idx int) (interface{}, bool)
	// GetByKey returns a value in the SortedKeyMap for a given key.
	// If there is no such key in the SortedKeyMap, then function returns FALSE as a second return value.
	GetByKey(key string) (interface{}, bool)
	// Index returns an index in the SortedKeyMap for a given key.
	// If there is no such index in the SortedKeyMap, then function returns FALSE as a second return value.
	Index(key string) (int, bool)
	// Key returns a key in the SortedKeyMap for a given index.
	// If there is no such key in the SortedKeyMap, then function returns FALSE as a second return value.
	Key(idx int) (string, bool)
	// Len returns a length of the SortedKeyMap.
	Len() int
	// Over iterates oven the SortedKeyMap. It calls OverFunc on each item in the SortedKeyMap.
	// If OverFunc returns FALSE on some item, then function is broke.
	Over(fn OverFunc)
	// Reset resets the state of a SortedKeyMap.
	Reset()
}

func Test(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&SortedKeyMapSuite{
	builder: func() SortedKeyMap {
		return NewSortedKeyMap()
	},
})

type SortedKeyMapSuite struct {
	builder func() SortedKeyMap
	skm     SortedKeyMap
}

func (s *SortedKeyMapSuite) SetUpTest(c *C) {
	s.skm = s.builder()
}

func (s *SortedKeyMapSuite) TearDownTest(c *C) {
	s.skm.Reset()
}

var testsData = []struct {
	givenKey   string
	givenValue int
}{
	{givenKey: "a", givenValue: 50},
	{givenKey: "b", givenValue: 40},
	{givenKey: "c", givenValue: 30},
	{givenKey: "d", givenValue: 20},
	{givenKey: "e", givenValue: 10},
}

func (s *SortedKeyMapSuite) TestAdd(c *C) {
	// The first attempt is success.
	for _, testData := range testsData {
		var result = s.skm.Add(testData.givenKey, testData.givenValue)
		c.Assert(result, Equals, true)
	}
	// The second attempt is failed, because all keys are presented already.
	for _, testData := range testsData {
		var result = s.skm.Add(testData.givenKey, testData.givenValue)
		c.Assert(result, Equals, false)
	}
}

func (s *SortedKeyMapSuite) TestExistsIndex(c *C) {
	s.addTestData(c)
	for idx := 0; idx < len(testsData)+1; idx++ {
		// It is expected to have index from testsData in the sorted key map.
		// But if index is out of range, then an expected result is FALSE.
		var expectedResult = idx < len(testsData)
		c.Assert(s.skm.ExistsIndex(idx), Equals, expectedResult)
	}
}

func (s *SortedKeyMapSuite) TestExistsKey(c *C) {
	s.addTestData(c)
	// Check that all data is added.
	for _, testData := range testsData {
		c.Assert(s.skm.ExistsKey(testData.givenKey), Equals, true)
	}
	c.Assert(s.skm.ExistsKey("not found"), Equals, false)
}

func (s *SortedKeyMapSuite) TestGetByIndex(c *C) {
	s.addTestData(c)
	for idx, testData := range testsData {
		v, ok := s.skm.GetByIndex(idx)
		c.Assert(ok, Equals, true)
		c.Assert(v, Equals, testData.givenValue)
	}
	var outOfIndex = len(testsData)
	_, ok := s.skm.GetByIndex(outOfIndex)
	c.Assert(ok, Equals, false)
}

func (s *SortedKeyMapSuite) TestGetByKey(c *C) {
	s.addTestData(c)
	for _, testData := range testsData {
		v, ok := s.skm.GetByKey(testData.givenKey)
		c.Assert(ok, Equals, true)
		c.Assert(v, Equals, testData.givenValue)
	}
	_, ok := s.skm.GetByKey("not found")
	c.Assert(ok, Equals, false)
}

func (s *SortedKeyMapSuite) TestIndex(c *C) {
	s.addTestData(c)
	for actualIdx, testData := range testsData {
		idx, ok := s.skm.Index(testData.givenKey)
		c.Assert(ok, Equals, true)
		c.Assert(idx, Equals, actualIdx)
	}
	_, ok := s.skm.Index("not found")
	c.Assert(ok, Equals, false)
}

func (s *SortedKeyMapSuite) TestKey(c *C) {
	s.addTestData(c)
	for actualIdx, testData := range testsData {
		key, ok := s.skm.Key(actualIdx)
		c.Assert(ok, Equals, true)
		c.Assert(key, Equals, testData.givenKey)
	}
	var outOfIndex = len(testsData)
	_, ok := s.skm.Key(outOfIndex)
	c.Assert(ok, Equals, false)
}

func (s *SortedKeyMapSuite) TestLen(c *C) {
	s.addTestData(c)
	c.Assert(s.skm.Len(), Equals, len(testsData))
}

func (s *SortedKeyMapSuite) TestOver(c *C) {
	s.addTestData(c)

	var i = 0
	s.skm.Over(func(idx int, key string, value interface{}) bool {
		c.Assert(idx, Equals, i)
		c.Assert(key, Equals, testsData[idx].givenKey)
		c.Assert(value, Equals, testsData[idx].givenValue)
		i += 1
		return true
	})
}

func (s *SortedKeyMapSuite) addTestData(c *C) {
	for _, testData := range testsData {
		var ok = s.skm.Add(testData.givenKey, testData.givenValue)
		c.Assert(ok, Equals, true)
	}
}
