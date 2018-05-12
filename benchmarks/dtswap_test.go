package benchmarks

import "testing"

func dtSwap1(a interface{}, b interface{}) { t := a; a = b; b = t }
func dtSwap2(a float64, b float64)         { t := a; a = b; b = t }

func Benchmark_interface(t *testing.B) {
	for i := 0; i < t.N; i++ {
		var a float64 = float64(i)
		var b float64 = float64(i + 100)
		dtSwap1(a, b)
	}
}

func Benchmark_float64(t *testing.B) {
	for i := 0; i < t.N; i++ {
		var a float64 = float64(i)
		var b float64 = float64(i + 100)
		dtSwap2(a, b)
	}
}
