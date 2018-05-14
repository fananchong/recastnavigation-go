package tests

import (
	"testing"

	"github.com/fananchong/recastnavigation-go/Detour"
)

func Test_dtSwap(t *testing.T) {
	{
		var a float32 = 1
		var b float32 = 2
		detour.DtSwapFloat32(&a, &b)
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
		var a float32 = 1
		var b float32 = 2
		c := detour.DtMinFloat32(a, b)
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
		var a float32 = 1
		var b float32 = 2
		c := detour.DtMaxFloat32(a, b)
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
		var a1 float32 = -1
		var b1 float32 = 1
		a2 := detour.DtAbsFloat32(a1)
		b2 := detour.DtAbsFloat32(b1)
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
		var b float32 = 2
		c := detour.DtSqrFloat32(b)
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
		a := detour.DtClampFloat32(0, 1, 3)
		b := detour.DtClampFloat32(2, 1, 3)
		c := detour.DtClampFloat32(4, 1, 3)
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
	v1 := [3]float32{1, 2, 3}
	v2 := [3]float32{4, 5, 6}
	dest := [3]float32{}
	detour.DtVcross(dest[:], v1[:], v2[:])
	detour.DtAssert(IsEquals(dest[0], -3) && IsEquals(dest[1], 6) && IsEquals(dest[2], -3))
}

func Test_dtVdot(t *testing.T) {
	v1 := [3]float32{1, 2, 3}
	v2 := [3]float32{4, 5, 6}
	v := detour.DtVdot(v1[:], v2[:])
	detour.DtAssert(IsEquals(v, 32))
}

func Test_dtVmad(t *testing.T) {
	v1 := [3]float32{1, 2, 3}
	v2 := [3]float32{4, 5, 6}
	dest := [3]float32{}
	detour.DtVmad(dest[:], v1[:], v2[:], 2)
	detour.DtAssert(IsEquals(dest[0], 9) && IsEquals(dest[1], 12) && IsEquals(dest[2], 15))
}

func Test_dtVlerp(t *testing.T) {
	v1 := [3]float32{1, 2, 3}
	v2 := [3]float32{4, 5, 6}
	dest := [3]float32{}
	detour.DtVlerp(dest[:], v1[:], v2[:], 0.2)
	detour.DtAssert(IsEquals(dest[0], 1.6) && IsEquals(dest[1], 2.6) && IsEquals(dest[2], 3.6))
}

func Test_dtVadd(t *testing.T) {
	v1 := [3]float32{1, 2, 3}
	v2 := [3]float32{4, 5, 6}
	dest := [3]float32{}
	detour.DtVadd(dest[:], v1[:], v2[:])
	detour.DtAssert(IsEquals(dest[0], 5) && IsEquals(dest[1], 7) && IsEquals(dest[2], 9))
}

func Test_dtVsub(t *testing.T) {
	v1 := [3]float32{1, 2, 3}
	v2 := [3]float32{4, 5, 6}
	dest := [3]float32{}
	detour.DtVsub(dest[:], v1[:], v2[:])
	detour.DtAssert(IsEquals(dest[0], -3) && IsEquals(dest[1], -3) && IsEquals(dest[2], -3))
}

func Test_dtVscale(t *testing.T) {
	v := [3]float32{1, 2, 3}
	dest := [3]float32{}
	detour.DtVscale(dest[:], v[:], 2)
	detour.DtAssert(IsEquals(dest[0], 2) && IsEquals(dest[1], 4) && IsEquals(dest[2], 6))
}

func Test_dtVmin(t *testing.T) {
	v := [3]float32{1, 2, 3}
	mn := [3]float32{-1, 3, 2}
	detour.DtVmin(mn[:], v[:])
	detour.DtAssert(IsEquals(mn[0], -1) && IsEquals(mn[1], 2) && IsEquals(mn[2], 2))
}

func Test_dtVmax(t *testing.T) {
	v := [3]float32{1, 2, 3}
	mx := [3]float32{-1, 3, 2}
	detour.DtVmax(mx[:], v[:])
	detour.DtAssert(IsEquals(mx[0], 1) && IsEquals(mx[1], 3) && IsEquals(mx[2], 3))
}

func Test_dtVset(t *testing.T) {
	dest := [3]float32{}
	detour.DtVset(dest[:], 1, 2, 3)
	detour.DtAssert(IsEquals(dest[0], 1) && IsEquals(dest[1], 2) && IsEquals(dest[2], 3))
}

func Test_dtVcopy(t *testing.T) {
	dest := [3]float32{}
	a := [3]float32{1, 2, 3}
	detour.DtVcopy(dest[:], a[:])
	detour.DtAssert(IsEquals(dest[0], 1) && IsEquals(dest[1], 2) && IsEquals(dest[2], 3))
}

func Test_dtVlen(t *testing.T) {
	v1 := [3]float32{1, 2, 3}
	v := detour.DtVlen(v1[:])
	detour.DtAssert(IsEquals(v, 3.74166))
}

func Test_dtVlenSqr(t *testing.T) {
	v1 := [3]float32{1, 2, 3}
	v := detour.DtVlenSqr(v1[:])
	detour.DtAssert(IsEquals(v, 14))
}

func Test_dtVdist(t *testing.T) {
	v1 := [3]float32{1, 2, 3}
	v2 := [3]float32{4, 5, 6}
	v := detour.DtVdist(v1[:], v2[:])
	detour.DtAssert(IsEquals(v, 5.19615))
}

func Test_dtVdistSqr(t *testing.T) {
	v1 := [3]float32{1, 2, 3}
	v2 := [3]float32{4, 5, 6}
	v := detour.DtVdistSqr(v1[:], v2[:])
	detour.DtAssert(IsEquals(v, 27))
}

func Test_dtVdist2D(t *testing.T) {
	v1 := [3]float32{1, 2, 3}
	v2 := [3]float32{4, 5, 6}
	v := detour.DtVdist2D(v1[:], v2[:])
	detour.DtAssert(IsEquals(v, 4.24264))
}

func Test_dtVdist2DSqr(t *testing.T) {
	v1 := [3]float32{1, 2, 3}
	v2 := [3]float32{4, 5, 6}
	v := detour.DtVdist2DSqr(v1[:], v2[:])
	detour.DtAssert(IsEquals(v, 18))
}

func Test_dtVnormalize(t *testing.T) {
	v := [3]float32{1, 2, 3}
	detour.DtVnormalize(v[:])
	detour.DtAssert(IsEquals(v[0], 0.26726) && IsEquals(v[1], 0.53452) && IsEquals(v[2], 0.80178))
}

func Test_dtVequal(t *testing.T) {
	var a float32 = 1.1111
	var b float32 = 2.2222
	var c float32 = 3.3333
	v1 := [3]float32{1.1111, 2.2222, 3.3333}
	v2 := [3]float32{a, b, c}
	v := detour.DtVequal(v1[:], v2[:])
	detour.DtAssert(v)
}

func Test_dtVdot2D(t *testing.T) {
	v1 := [3]float32{1, 2, 3}
	v2 := [3]float32{4, 5, 6}
	v := detour.DtVdot2D(v1[:], v2[:])
	detour.DtAssert(IsEquals(v, 22))
}

func Test_dtVperp2D(t *testing.T) {
	v1 := [3]float32{1, 2, 3}
	v2 := [3]float32{4, 5, 6}
	v := detour.DtVperp2D(v1[:], v2[:])
	detour.DtAssert(IsEquals(v, 6))
}

func Test_dtTriArea2D(t *testing.T) {
	v1 := [3]float32{1, 2, 3}
	v2 := [3]float32{4, 5, 6}
	v3 := [3]float32{7, 8, 3}
	v := detour.DtTriArea2D(v1[:], v2[:], v3[:])
	detour.DtAssert(IsEquals(v, 18))
}

func Test_dtOverlapQuantBounds(t *testing.T) {
	amin := [3]uint16{1, 1, 1}
	amax := [3]uint16{2, 1, 2}
	bmin := [3]uint16{3, 1, 2}
	bmax := [3]uint16{4, 1, 3}
	v := detour.DtOverlapQuantBounds(amin[:], amax[:], bmin[:], bmax[:])
	detour.DtAssert(v == false)
}

func Test_dtOverlapBounds(t *testing.T) {
	amin := [3]float32{1, 1, 1}
	amax := [3]float32{2, 1, 2}
	bmin := [3]float32{1, 1, 1}
	bmax := [3]float32{4, 1, 3}
	v := detour.DtOverlapBounds(amin[:], amax[:], bmin[:], bmax[:])
	detour.DtAssert(v == true)
}

func Test_dtClosestPtPointTriangle(t *testing.T) {
	closest := [3]float32{}
	p := [3]float32{-1, 2, -1}
	a := [3]float32{1, 2, 1}
	b := [3]float32{3, 2, 1}
	c := [3]float32{2, 2, 3}
	detour.DtClosestPtPointTriangle(closest[:], p[:], a[:], b[:], c[:])
	detour.DtAssert(IsEquals(closest[0], 1) && IsEquals(closest[1], 2) && IsEquals(closest[2], 1))
	p = [3]float32{4, 2, 0}
	detour.DtClosestPtPointTriangle(closest[:], p[:], a[:], b[:], c[:])
	detour.DtAssert(IsEquals(closest[0], 3) && IsEquals(closest[1], 2) && IsEquals(closest[2], 1))
	p = [3]float32{3, 2, 4}
	detour.DtClosestPtPointTriangle(closest[:], p[:], a[:], b[:], c[:])
	detour.DtAssert(IsEquals(closest[0], 2) && IsEquals(closest[1], 2) && IsEquals(closest[2], 3))
	p = [3]float32{2, 2, 0}
	detour.DtClosestPtPointTriangle(closest[:], p[:], a[:], b[:], c[:])
	detour.DtAssert(IsEquals(closest[0], 2) && IsEquals(closest[1], 2) && IsEquals(closest[2], 1))
	p = [3]float32{2, 2, 2}
	detour.DtClosestPtPointTriangle(closest[:], p[:], a[:], b[:], c[:])
	detour.DtAssert(IsEquals(closest[0], 2) && IsEquals(closest[1], 2) && IsEquals(closest[2], 2))
}

func Test_dtClosestHeightPointTriangle(t *testing.T) {
	p := [3]float32{2, 4, 2}
	a := [3]float32{1, 2, 1}
	b := [3]float32{3, 2, 1}
	c := [3]float32{2, 2, 3}
	var h float32
	ok := detour.DtClosestHeightPointTriangle(p[:], a[:], b[:], c[:], &h)
	detour.DtAssert(ok && IsEquals(h, 2))
}

func Test_for(t *testing.T) {
	nverts := 4
	indexs := make([]int, 0)
	for i, j := 0, nverts-1; i < nverts; j, i = i, i+1 {
		indexs = append(indexs, i, j)
	}
	detour.DtAssert(indexs[0] == 0 && indexs[1] == 3 && indexs[2] == 1 && indexs[3] == 0 && indexs[4] == 2 && indexs[5] == 1 && indexs[6] == 3 && indexs[7] == 2)
}

func Test_dtIntersectSegmentPoly2D(t *testing.T) {
	p0 := [3]float32{-1, 2, 1}
	p1 := [3]float32{5, 2, 1}
	verts := [12]float32{1, 2, 2, 4, 2, 2, 4, 2, 0, 1, 2, 0}
	nverts := len(verts) / 3
	var tmin, tmax float32
	var segMin, segMax int
	ok := detour.DtIntersectSegmentPoly2D(p0[:], p1[:], verts[:], nverts, &tmin, &tmax, &segMin, &segMax)
	detour.DtAssert(ok && segMin == 3 && segMax == 1 && IsEquals(tmin, 2/6.0) && IsEquals(tmax, 5/6.0))
}

func Test_dtIntersectSegSeg2D(t *testing.T) {
	ap := [3]float32{-1, 2, 1}
	aq := [3]float32{4, 2, 1}
	bp := [3]float32{0, 2, 0}
	bq := [3]float32{3, 2, 3}
	var s0, t0 float32
	ok := detour.DtIntersectSegSeg2D(ap[:], aq[:], bp[:], bq[:], &s0, &t0)
	detour.DtAssert(ok && IsEquals(s0, 2/5.0) && IsEquals(t0, 1/3.0))
}

func Test_dtPointInPolygon(t *testing.T) {
	pt := [3]float32{2, 2, 1}
	verts := [12]float32{1, 2, 2, 4, 2, 2, 4, 2, 0, 1, 2, 0}
	nverts := len(verts) / 3
	ok := detour.DtPointInPolygon(pt[:], verts[:], nverts)
	detour.DtAssert(ok)
	pt = [3]float32{2, 2, 3}
	ok = detour.DtPointInPolygon(pt[:], verts[:], nverts)
	detour.DtAssert(ok == false)
}

func Test_dtDistancePtPolyEdgesSqr(t *testing.T) {
	pt := [3]float32{2, 2, 1}
	verts := [12]float32{1, 2, 2, 4, 2, 2, 4, 2, 0, 1, 2, 0}
	nverts := len(verts) / 3
	ed := make([]float32, nverts)
	et := make([]float32, nverts)
	ok := detour.DtDistancePtPolyEdgesSqr(pt[:], verts[:], nverts, ed[:], et[:])
	detour.DtAssert(ok &&
		IsEquals(ed[0], 1) &&
		IsEquals(ed[1], 4) &&
		IsEquals(ed[2], 1) &&
		IsEquals(ed[3], 1) &&
		IsEquals(et[0], 0.3333333333333333) &&
		IsEquals(et[1], 0.5) &&
		IsEquals(et[2], 0.6666666666666666) &&
		IsEquals(et[3], 0.5))
}

func Test_dtCalcPolyCenter(t *testing.T) {
	tc := [3]float32{}
	verts := [12]float32{1, 2, 2, 4, 2, 2, 4, 2, 0, 1, 2, 0}
	idx := [4]uint16{0, 1, 2, 3}
	nidx := len(idx)
	detour.DtCalcPolyCenter(tc[:], idx[:], nidx, verts[:])
	detour.DtAssert(IsEquals(tc[0], 2.5) && IsEquals(tc[1], 2) && IsEquals(tc[2], 1))
}

func Test_dtOverlapPolyPoly2D(t *testing.T) {
	polya := [12]float32{1, 2, 2, 4, 2, 2, 4, 2, 0, 1, 2, 0}
	npolya := len(polya) / 3
	polyb := [9]float32{2, 2, 1, 2.5, 2, 3, 4, 2, 1}
	npolyb := len(polyb) / 3
	ok := detour.DtOverlapPolyPoly2D(polya[:], npolya, polyb[:], npolyb)
	detour.DtAssert(ok)
}

func Test_dtNextPow2(t *testing.T) {
	v := []uint32{
		detour.DtNextPow2(0),
		detour.DtNextPow2(1),
		detour.DtNextPow2(2),
		detour.DtNextPow2(3),
		detour.DtNextPow2(4),
		detour.DtNextPow2(5),
		detour.DtNextPow2(6),
		detour.DtNextPow2(7),
		detour.DtNextPow2(8),
		detour.DtNextPow2(9),
		detour.DtNextPow2(10)}
	detour.DtAssert(v[0] == 0 &&
		v[1] == 1 &&
		v[2] == 2 &&
		v[3] == 4 &&
		v[4] == 4 &&
		v[5] == 8 &&
		v[6] == 8 &&
		v[7] == 8 &&
		v[8] == 8 &&
		v[9] == 16 &&
		v[10] == 16)
}

func Test_dtIlog2(t *testing.T) {
	v := []uint32{
		detour.DtIlog2(0),
		detour.DtIlog2(1),
		detour.DtIlog2(2),
		detour.DtIlog2(3),
		detour.DtIlog2(4),
		detour.DtIlog2(5),
		detour.DtIlog2(6),
		detour.DtIlog2(7),
		detour.DtIlog2(8),
		detour.DtIlog2(9),
		detour.DtIlog2(10)}
	detour.DtAssert(v[0] == 0 &&
		v[1] == 0 &&
		v[2] == 1 &&
		v[3] == 1 &&
		v[4] == 2 &&
		v[5] == 2 &&
		v[6] == 2 &&
		v[7] == 2 &&
		v[8] == 3 &&
		v[9] == 3 &&
		v[10] == 3)
}

func Test_dtAlign4(t *testing.T) {
	v := []int{
		detour.DtAlign4(0),
		detour.DtAlign4(1),
		detour.DtAlign4(2),
		detour.DtAlign4(3),
		detour.DtAlign4(4),
		detour.DtAlign4(5),
		detour.DtAlign4(6),
		detour.DtAlign4(7),
		detour.DtAlign4(8),
		detour.DtAlign4(9),
		detour.DtAlign4(10)}
	detour.DtAssert(v[0] == 0 &&
		v[1] == 4 &&
		v[2] == 4 &&
		v[3] == 4 &&
		v[4] == 4 &&
		v[5] == 8 &&
		v[6] == 8 &&
		v[7] == 8 &&
		v[8] == 8 &&
		v[9] == 12 &&
		v[10] == 12)
}

func Test_dtOppositeTile(t *testing.T) {
	v := []int{
		detour.DtOppositeTile(0),
		detour.DtOppositeTile(1),
		detour.DtOppositeTile(2),
		detour.DtOppositeTile(3),
		detour.DtOppositeTile(4),
		detour.DtOppositeTile(5),
		detour.DtOppositeTile(6),
		detour.DtOppositeTile(7),
		detour.DtOppositeTile(8),
		detour.DtOppositeTile(9),
		detour.DtOppositeTile(10)}
	detour.DtAssert(v[0] == 4 &&
		v[1] == 5 &&
		v[2] == 6 &&
		v[3] == 7 &&
		v[4] == 0 &&
		v[5] == 1 &&
		v[6] == 2 &&
		v[7] == 3 &&
		v[8] == 4 &&
		v[9] == 5 &&
		v[10] == 6)
}

func Test_dtSwapByte(t *testing.T) {
	var a uint8 = 1
	var b uint8 = 2
	detour.DtSwapByte(&a, &b)
	detour.DtAssert(a == 2 && b == 1)
}

func Test_dtSwapEndian(t *testing.T) {
	{
		var v uint16 = 1
		detour.DtSwapEndianUInt16(&v)
		detour.DtAssert(v == 256)
	}
	{
		var v int16 = 1
		detour.DtSwapEndianInt16(&v)
		detour.DtAssert(v == 256)
	}
	{
		var v uint32 = 1
		detour.DtSwapEndianUInt32(&v)
		detour.DtAssert(v == 16777216)
	}
	{
		var v int32 = 1
		detour.DtSwapEndianInt32(&v)
		detour.DtAssert(v == 16777216)
	}
	{
		var v float32 = 100.9
		detour.DtSwapEndianFloat32(&v)
		detour.DtAssert(v == -429467712.000000)
	}
}
