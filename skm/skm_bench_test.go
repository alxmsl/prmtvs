package skm

import (
	"strconv"
	"testing"
)

const (
	datasetSize = 100000
	latestIndex = datasetSize - 1
)

var (
	skm  = NewSKM()
	sskm = NewSafeSKM()
	keys = make([]string, datasetSize)
)

func init() {
	for i := 0; i < datasetSize; i++ {
		keys[i] = strconv.Itoa(i)
	}
}

func BenchmarkAddInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		skm.Add(keys[i%latestIndex], i)
	}
}

func BenchmarkGetByIndexInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		skm.GetByIndex(i % latestIndex)
	}
}

func BenchmarkExistsIndexInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		skm.ExistsIndex(i % latestIndex)
	}
}

func BenchmarkSafeAddInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sskm.Add(keys[i%latestIndex], i)
	}
}

func BenchmarkSafeGetByIndexInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sskm.GetByIndex(i % latestIndex)
	}
}

func BenchmarkSafeExistsIndexInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sskm.ExistsIndex(i % latestIndex)
	}
}
