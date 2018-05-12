package tests

import (
	"math"
)

func IsEquals(a, b float64) bool {
	return math.Abs(a-b) < 0.00001
}
