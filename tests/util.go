package tests

import (
	"math"
)

func IsEquals(a, b float32) bool {
	return math.Abs(float64(a-b)) < 0.00001
}
