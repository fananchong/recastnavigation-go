package tests

import (
	"math"
	"testing"

	"github.com/fananchong/recastnavigation-go/Detour"
)

func Test_dtMath(t *testing.T) {
	detour.DtAssert(detour.DtMathFabsf(-3.1) == 3.1)
	detour.DtAssert(detour.DtMathSqrtf(4) == 2)
	detour.DtAssert(detour.DtMathFloorf(5.3) == 5)
	detour.DtAssert(detour.DtMathCeilf(6.6) == 7)
	detour.DtAssert(detour.DtMathCosf(180) == -1)
	detour.DtAssert(detour.DtMathSinf(180) == 0)
	detour.DtAssert(detour.DtMathAtan2f(1, 1) == math.Pi/4)
}
