package skm

import (
	. "gopkg.in/check.v1"
)

var _ = Suite(&SortedKeyMapSuite{
	builder: func() SortedKeyMap {
		return NewSafeSortedKeyMap()
	},
})
