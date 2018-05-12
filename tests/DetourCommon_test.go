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
	detour.DtAssert(IsEquals(dest[0], -3) && IsEquals(dest[1], 6) && IsEquals(dest[2], -3))
}

func Test_dtVdot(t *testing.T) {
	v1 := [3]float64{1, 2, 3}
	v2 := [3]float64{4, 5, 6}
	v := detour.DtVdot(&v1, &v2)
	detour.DtAssert(IsEquals(v, 32))
}

func Test_dtVmad(t *testing.T) {
	v1 := [3]float64{1, 2, 3}
	v2 := [3]float64{4, 5, 6}
	dest := [3]float64{0, 0, 0}
	detour.DtVmad(&dest, &v1, &v2, 2)
	detour.DtAssert(IsEquals(dest[0], 9) && IsEquals(dest[1], 12) && IsEquals(dest[2], 15))
}

func Test_dtVlerp(t *testing.T) {
	v1 := [3]float64{1, 2, 3}
	v2 := [3]float64{4, 5, 6}
	dest := [3]float64{0, 0, 0}
	detour.DtVlerp(&dest, &v1, &v2, 0.2)
	detour.DtAssert(IsEquals(dest[0], 1.6) && IsEquals(dest[1], 2.6) && IsEquals(dest[2], 3.6))
}

func Test_dtVadd(t *testing.T) {
	v1 := [3]float64{1, 2, 3}
	v2 := [3]float64{4, 5, 6}
	dest := [3]float64{0, 0, 0}
	detour.DtVadd(&dest, &v1, &v2)
	detour.DtAssert(IsEquals(dest[0], 5) && IsEquals(dest[1], 7) && IsEquals(dest[2], 9))
}

func Test_dtVsub(t *testing.T) {
	v1 := [3]float64{1, 2, 3}
	v2 := [3]float64{4, 5, 6}
	dest := [3]float64{0, 0, 0}
	detour.DtVsub(&dest, &v1, &v2)
	detour.DtAssert(IsEquals(dest[0], -3) && IsEquals(dest[1], -3) && IsEquals(dest[2], -3))
}

func Test_dtVscale(t *testing.T) {
	v := [3]float64{1, 2, 3}
	dest := [3]float64{0, 0, 0}
	detour.DtVscale(&dest, &v, 2)
	detour.DtAssert(IsEquals(dest[0], 2) && IsEquals(dest[1], 4) && IsEquals(dest[2], 6))
}

func Test_dtVmin(t *testing.T) {
	v := [3]float64{1, 2, 3}
	mn := [3]float64{-1, 3, 2}
	detour.DtVmin(&mn, &v)
	detour.DtAssert(IsEquals(mn[0], -1) && IsEquals(mn[1], 2) && IsEquals(mn[2], 2))
}

func Test_dtVmax(t *testing.T) {
	v := [3]float64{1, 2, 3}
	mx := [3]float64{-1, 3, 2}
	detour.DtVmax(&mx, &v)
	detour.DtAssert(IsEquals(mx[0], 1) && IsEquals(mx[1], 3) && IsEquals(mx[2], 3))
}

func Test_dtVset(t *testing.T) {
	dest := [3]float64{0, 0, 0}
	detour.DtVset(&dest, 1, 2, 3)
	detour.DtAssert(IsEquals(dest[0], 1) && IsEquals(dest[1], 2) && IsEquals(dest[2], 3))
}

func Test_dtVcopy(t *testing.T) {
	dest := [3]float64{0, 0, 0}
	a := [3]float64{1, 2, 3}
	detour.DtVcopy(&dest, &a)
	detour.DtAssert(IsEquals(dest[0], 1) && IsEquals(dest[1], 2) && IsEquals(dest[2], 3))
}

func Test_dtVlen(t *testing.T) {
	v1 := [3]float64{1, 2, 3}
	v := detour.DtVlen(&v1)
	detour.DtAssert(IsEquals(v, 3.74166))
}

func Test_dtVlenSqr(t *testing.T) {
	v1 := [3]float64{1, 2, 3}
	v := detour.DtVlenSqr(&v1)
	detour.DtAssert(IsEquals(v, 14))
}

func Test_dtVdist(t *testing.T) {
	v1 := [3]float64{1, 2, 3}
	v2 := [3]float64{4, 5, 6}
	v := detour.DtVdist(&v1, &v2)
	detour.DtAssert(IsEquals(v, 5.19615))
}

func Test_dtVdistSqr(t *testing.T) {
	v1 := [3]float64{1, 2, 3}
	v2 := [3]float64{4, 5, 6}
	v := detour.DtVdistSqr(&v1, &v2)
	detour.DtAssert(IsEquals(v, 27))
}

func Test_dtVdist2D(t *testing.T) {
	v1 := [3]float64{1, 2, 3}
	v2 := [3]float64{4, 5, 6}
	v := detour.DtVdist2D(&v1, &v2)
	detour.DtAssert(IsEquals(v, 4.24264))
}

func Test_dtVdist2DSqr(t *testing.T) {
	v1 := [3]float64{1, 2, 3}
	v2 := [3]float64{4, 5, 6}
	v := detour.DtVdist2DSqr(&v1, &v2)
	detour.DtAssert(IsEquals(v, 18))
}

func Test_dtVnormalize(t *testing.T) {
	v := [3]float64{1, 2, 3}
	detour.DtVnormalize(&v)
	detour.DtAssert(IsEquals(v[0], 0.26726) && IsEquals(v[1], 0.53452) && IsEquals(v[2], 0.80178))
}

func Test_dtVequal(t *testing.T) {
	a := 1.1111
	b := 2.2222
	c := 3.3333
	v1 := [3]float64{1.1111, 2.2222, 3.3333}
	v2 := [3]float64{a, b, c}
	v := detour.DtVequal(&v1, &v2)
	detour.DtAssert(v)
}

func Test_dtVdot2D(t *testing.T) {
	v1 := [3]float64{1, 2, 3}
	v2 := [3]float64{4, 5, 6}
	v := detour.DtVdot2D(&v1, &v2)
	detour.DtAssert(IsEquals(v, 22))
}

func Test_dtVperp2D(t *testing.T) {
	v1 := [3]float64{1, 2, 3}
	v2 := [3]float64{4, 5, 6}
	v := detour.DtVperp2D(&v1, &v2)
	detour.DtAssert(IsEquals(v, 6))
}

func Test_dtTriArea2D(t *testing.T) {
	v1 := [3]float64{1, 2, 3}
	v2 := [3]float64{4, 5, 6}
	v3 := [3]float64{7, 8, 3}
	v := detour.DtTriArea2D(&v1, &v2, &v3)
	detour.DtAssert(IsEquals(v, 18))
}

func Test_dtOverlapQuantBounds(t *testing.T) {
	amin := [3]uint16{1, 1, 1}
	amax := [3]uint16{2, 1, 2}
	bmin := [3]uint16{3, 1, 2}
	bmax := [3]uint16{4, 1, 3}
	v := detour.DtOverlapQuantBounds(&amin, &amax, &bmin, &bmax)
	detour.DtAssert(v == false)
}

func Test_dtOverlapBounds(t *testing.T) {
	amin := [3]float64{1, 1, 1}
	amax := [3]float64{2, 1, 2}
	bmin := [3]float64{1, 1, 1}
	bmax := [3]float64{4, 1, 3}
	v := detour.DtOverlapBounds(&amin, &amax, &bmin, &bmax)
	detour.DtAssert(v == true)
}
