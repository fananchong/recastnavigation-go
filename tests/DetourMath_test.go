package tests

import (
	"math"
	"testing"

	"github.com/fananchong/recastnavigation-go/Detour"
)

func Test_dtMath(t *testing.T) {
	detour.DtAssert(IsEquals(detour.DtMathFabsf(-3.1), 3.1))
	detour.DtAssert(IsEquals(detour.DtMathSqrtf(4), 2))
	detour.DtAssert(IsEquals(detour.DtMathFloorf(5.3), 5))
	detour.DtAssert(IsEquals(detour.DtMathCeilf(6.6), 7))
	detour.DtAssert(IsEquals(detour.DtMathCosf(math.Pi), -1))
	detour.DtAssert(IsEquals(detour.DtMathSinf(math.Pi), 0))
	detour.DtAssert(IsEquals(detour.DtMathAtan2f(1, 1), math.Pi/4))
}
