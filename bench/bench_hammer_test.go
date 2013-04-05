package bench

import (
	"testing"
)

func Add2Ints(i, j int) int {
	return i + j
}

func Benchmark_Hammer(b *testing.B) {
	b.StopTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Add2Ints(i, 100)
	}
}
