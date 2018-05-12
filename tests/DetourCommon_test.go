package tests

import (
	"testing"

	"github.com/fananchong/recastnavigation-go/Detour"
)

func Test_dtSwap(t *testing.T) {
	{
		var a float64 = 1
		var b float64 = 2
		detour.DtSwapFloat64(&a, &b)
		detour.DtAssert(IsEquals(a, 2) && IsEquals(b, 1))
	}

	{
		var a uint32 = 1
		var b uint32 = 2
		detour.DtSwapUInt32(&a, &b)
		detour.DtAssert(a == 2 && b == 1)
	}
	{
		var a int32 = 1
		var b int32 = 2
		detour.DtSwapInt32(&a, &b)
		detour.DtAssert(a == 2 && b == 1)
	}
	{
		var a uint16 = 1
		var b uint16 = 2
		detour.DtSwapUInt16(&a, &b)
		detour.DtAssert(a == 2 && b == 1)
	}
	{
		var a int16 = 1
		var b int16 = 2
		detour.DtSwapInt16(&a, &b)
		detour.DtAssert(a == 2 && b == 1)
	}
}

func Test_dtMin(t *testing.T) {
	{
		var a float64 = 1
		var b float64 = 2
		c := detour.DtMinFloat64(a, b)
		detour.DtAssert(IsEquals(c, 1))
	}
	{
		var a uint32 = 1
		var b uint32 = 2
		c := detour.DtMinUInt32(a, b)
		detour.DtAssert(c == 1)
	}
	{
		var a int32 = 1
		var b int32 = 2
		c := detour.DtMinInt32(a, b)
		detour.DtAssert(c == 1)
	}
	{
		var a uint16 = 1
		var b uint16 = 2
		c := detour.DtMinUInt16(a, b)
		detour.DtAssert(c == 1)
	}
	{
		var a int16 = 1
		var b int16 = 2
		c := detour.DtMinInt16(a, b)
		detour.DtAssert(c == 1)
	}
}

func Test_dtMax(t *testing.T) {
	{
		var a float64 = 1
		var b float64 = 2
		c := detour.DtMaxFloat64(a, b)
		detour.DtAssert(IsEquals(c, 2))
	}
	{
		var a uint32 = 1
		var b uint32 = 2
		c := detour.DtMaxUInt32(a, b)
		detour.DtAssert(c == 2)
	}
	{
		var a int32 = 1
		var b int32 = 2
		c := detour.DtMaxInt32(a, b)
		detour.DtAssert(c == 2)
	}
	{
		var a uint16 = 1
		var b uint16 = 2
		c := detour.DtMaxUInt16(a, b)
		detour.DtAssert(c == 2)
	}
	{
		var a int16 = 1
		var b int16 = 2
		c := detour.DtMaxInt16(a, b)
		detour.DtAssert(c == 2)
	}
}

func Test_dtAbs(t *testing.T) {
	{
		var a1 float64 = -1
		var b1 float64 = 1
		a2 := detour.DtAbsFloat64(a1)
		b2 := detour.DtAbsFloat64(b1)
		detour.DtAssert(IsEquals(a2, 1) && IsEquals(b2, 1))
	}
	{
		var a1 int32 = -1
		var b1 int32 = 1
		a2 := detour.DtAbsInt32(a1)
		b2 := detour.DtAbsInt32(b1)
		detour.DtAssert(a2 == 1 && b2 == 1)
	}
	{
		var a1 int16 = -1
		var b1 int16 = 1
		a2 := detour.DtAbsInt16(a1)
		b2 := detour.DtAbsInt16(b1)
		detour.DtAssert(a2 == 1 && b2 == 1)
	}
}

func Test_dtSqr(t *testing.T) {
	{
		var b float64 = 2
		c := detour.DtSqrFloat64(b)
		detour.DtAssert(IsEquals(c, 4))
	}
	{
		var b uint32 = 2
		c := detour.DtSqrUInt32(b)
		detour.DtAssert(c == 4)
	}
	{
		var b int32 = 2
		c := detour.DtSqrInt32(b)
		detour.DtAssert(c == 4)
	}
	{
		var b uint16 = 2
		c := detour.DtSqrUInt16(b)
		detour.DtAssert(c == 4)
	}
	{
		var b int16 = 2
		c := detour.DtSqrInt16(b)
		detour.DtAssert(c == 4)
	}
}

func Test_dtClamp(t *testing.T) {
	{
		a := detour.DtClampFloat64(0, 1, 3)
		b := detour.DtClampFloat64(2, 1, 3)
		c := detour.DtClampFloat64(4, 1, 3)
		detour.DtAssert(IsEquals(a, 1))
		detour.DtAssert(IsEquals(b, 2))
		detour.DtAssert(IsEquals(c, 3))
	}
	{
		a := detour.DtClampUInt32(0, 1, 3)
		b := detour.DtClampUInt32(2, 1, 3)
		c := detour.DtClampUInt32(4, 1, 3)
		detour.DtAssert(a == 1)
		detour.DtAssert(b == 2)
		detour.DtAssert(c == 3)
	}
	{
		a := detour.DtClampInt32(0, 1, 3)
		b := detour.DtClampInt32(2, 1, 3)
		c := detour.DtClampInt32(4, 1, 3)
		detour.DtAssert(a == 1)
		detour.DtAssert(b == 2)
		detour.DtAssert(c == 3)
	}
	{
		a := detour.DtClampUInt16(0, 1, 3)
		b := detour.DtClampUInt16(2, 1, 3)
		c := detour.DtClampUInt16(4, 1, 3)
		detour.DtAssert(a == 1)
		detour.DtAssert(b == 2)
		detour.DtAssert(c == 3)
	}
	{
		a := detour.DtClampInt16(0, 1, 3)
		b := detour.DtClampInt16(2, 1, 3)
		c := detour.DtClampInt16(4, 1, 3)
		detour.DtAssert(a == 1)
		detour.DtAssert(b == 2)
		detour.DtAssert(c == 3)
	}
}

func Test_dtVcross(t *testing.T) {
	v1 := [3]float64{1, 2, 3}
	v2 := [3]float64{4, 5, 6}
	dest := [3]float64{0, 0, 0}
	detour.DtVcross(&dest, &v1, &v2)
	detour.DtAssert(dest[0] == -3 && dest[1] == 6 && dest[2] == -3)
}
