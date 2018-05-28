package benchmarks

import "testing"

func farray(p *[3]float64) {
	p[0] += 1
	p[1] += 1
	p[2] += 1
}

func fslice(p []float64) {
	p[0] += 1
	p[1] += 1
	p[2] += 1
}

func Benchmark_array(t *testing.B) {
	p := [3]float64{1, 2, 3}
	for i := 0; i < t.N; i++ {
		farray(&p)
	}
}

func Benchmark_slice(t *testing.B) {
	p := [3]float64{1, 2, 3}
	for i := 0; i < t.N; i++ {
		fslice(p[:])
	}
}
