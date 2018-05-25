package mgl

import (
	"math"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
)

const (
	NUM_SEG = 16
)

var (
	dir   [NUM_SEG * 2]float32
	_init = false
)

func init() {
	for i := 0; i < NUM_SEG; i++ {
		a := float64(i) / float64(NUM_SEG) * float64(math.Pi*2)
		dir[i*2] = float32(math.Cos(a))
		dir[i*2+1] = float32(math.Sin(a))
	}
}

func DrawCylinder(minx, miny, minz, maxx, maxy, maxz float32, col uint32) {
	Begin(gl.TRIANGLES, 1)
	col2 := DuMultCol(col, 160)
	cx := (maxx + minx) / 2
	cz := (maxz + minz) / 2
	rx := (maxx - minx) / 2
	rz := (maxz - minz) / 2

	for i := 2; i < NUM_SEG; i++ {
		a := 0
		b := i - 1
		c := i
		Vertex2(cx+dir[a*2+0]*rx, miny, cz+dir[a*2+1]*rz, col2)
		Vertex2(cx+dir[b*2+0]*rx, miny, cz+dir[b*2+1]*rz, col2)
		Vertex2(cx+dir[c*2+0]*rx, miny, cz+dir[c*2+1]*rz, col2)
	}
	for i := 2; i < NUM_SEG; i++ {
		a := 0
		b := i
		c := i - 1
		Vertex2(cx+dir[a*2+0]*rx, maxy, cz+dir[a*2+1]*rz, col)
		Vertex2(cx+dir[b*2+0]*rx, maxy, cz+dir[b*2+1]*rz, col)
		Vertex2(cx+dir[c*2+0]*rx, maxy, cz+dir[c*2+1]*rz, col)
	}
	j := NUM_SEG - 1
	for i := 0; i < NUM_SEG; i++ {
		Vertex2(cx+dir[i*2+0]*rx, miny, cz+dir[i*2+1]*rz, col2)
		Vertex2(cx+dir[j*2+0]*rx, miny, cz+dir[j*2+1]*rz, col2)
		Vertex2(cx+dir[j*2+0]*rx, maxy, cz+dir[j*2+1]*rz, col)

		Vertex2(cx+dir[i*2+0]*rx, miny, cz+dir[i*2+1]*rz, col2)
		Vertex2(cx+dir[j*2+0]*rx, maxy, cz+dir[j*2+1]*rz, col)
		Vertex2(cx+dir[i*2+0]*rx, maxy, cz+dir[i*2+1]*rz, col)
		j = i
	}

	End()
}

func DuRGBA(r, g, b, a uint32) uint32 {
	return r | (g << 8) | (b << 16) | (a << 24)
}

func DuLerpCol(ca, cb, u uint32) uint32 {
	ra := ca & 0xff
	ga := (ca >> 8) & 0xff
	ba := (ca >> 16) & 0xff
	aa := (ca >> 24) & 0xff
	rb := cb & 0xff
	gb := (cb >> 8) & 0xff
	bb := (cb >> 16) & 0xff
	ab := (cb >> 24) & 0xff

	r := (ra*(255-u) + rb*u) / 255
	g := (ga*(255-u) + gb*u) / 255
	b := (ba*(255-u) + bb*u) / 255
	a := (aa*(255-u) + ab*u) / 255
	return DuRGBA(r, g, b, a)
}

func Begin(prim uint32, size float32) {
	if prim == gl.POINTS {
		gl.PointSize(size)
	} else if prim == gl.LINES {
		gl.LineWidth(size)
	}
	gl.Begin(prim)
}

func End() {
	gl.End()
	gl.LineWidth(1.0)
	gl.PointSize(1.0)
}

func Vertex(pos []float32, color uint32, uv [2]float32) {
	// fmt.Println("pos:", pos, " uv:", uv)
	gl.Color4ubv((*uint8)(unsafe.Pointer(&color)))
	gl.TexCoord2fv((*float32)(unsafe.Pointer(&uv)))
	gl.Vertex3fv((*float32)(unsafe.Pointer(&pos[0])))
}

func Vertex2(x, y, z float32, color uint32) {
	gl.Color4ubv((*uint8)(unsafe.Pointer(&color)))
	gl.Vertex3f(x, y, z)
}

func Texture(state bool) {
	if state {
		gl.Enable(gl.TEXTURE_2D)
		g_tex.Bind()
	} else {
		gl.Disable(gl.TEXTURE_2D)
	}
}

func isectSegAABB(sp, sq *[3]float64, amin, amax *[3]float32) (bool, float64, float64) {
	EPS := 1e-6
	var d = [3]float64{sq[0] - sp[0], sq[1] - sp[1], sq[2] - sp[2]}
	var tmin, tmax float64 = 0, 1
	for i := 0; i < 3; i++ {
		if math.Abs(d[i]) < EPS {
			if sp[i] < float64(amin[i]) || sp[i] > float64(amax[i]) {
				return false, 0, 0
			}
		} else {
			ood := 1.0 / d[i]
			t1 := (float64(amin[i]) - sp[i]) * ood
			t2 := (float64(amax[i]) - sp[i]) * ood
			if t1 > t2 {
				t1, t2 = t2, t1
			}
			if t1 > tmin {
				tmin = t1
			}
			if t2 < tmax {
				tmax = t2
			}
			if tmin > tmax {
				return false, 0, 0
			}
		}
	}
	return true, tmin, tmax
}

func intersectSegmentTriangle(sp, sq *[3]float64, a, b, c []float32) (bool, float32) {
	var v, w float32
	var ab, ac, qp, ap, norm, e [3]float32
	vSub(ab[0:], b, a)
	vSub(ac[0:], c, a)
	qp[0] = float32(sp[0] - sq[0])
	qp[1] = float32(sp[1] - sq[1])
	qp[2] = float32(sp[2] - sq[2])

	vCross(norm[0:], ab[0:], ac[0:])

	d := vDot(qp[0:], norm[0:])
	if d <= 0 {
		return false, 0
	}

	ap[0] = float32(sp[0]) - a[0]
	ap[1] = float32(sp[1]) - a[1]
	ap[2] = float32(sp[2]) - a[2]

	t := vDot(ap[0:], norm[0:])
	if t < 0 {
		return false, 0
	}
	if t > d {
		return false, 0
	}

	vCross(e[0:], qp[0:], ap[0:])

	v = vDot(ac[0:], e[0:])
	if v < 0 || v > d {
		return false, 0
	}
	w = -vDot(ab[0:], e[0:])
	if w < 0 || v+w > d {
		return false, 0
	}

	t /= d

	return true, t
}

func vSub(dest, a, b []float32) {
	dest[0] = a[0] - b[0]
	dest[1] = a[1] - b[1]
	dest[2] = a[2] - b[2]
}

func vCross(dest, v1, v2 []float32) {
	dest[0] = v1[1]*v2[2] - v1[2]*v2[1]
	dest[1] = v1[2]*v2[0] - v1[0]*v2[2]
	dest[2] = v1[0]*v2[1] - v1[1]*v2[0]
}

func vDot(v1, v2 []float32) float32 {
	return v1[0]*v2[0] + v1[1]*v2[1] + v1[2]*v2[2]
}

func DuMultCol(col, d uint32) uint32 {
	r := col & 0xff
	g := (col >> 8) & 0xff
	b := (col >> 16) & 0xff
	a := (col >> 24) & 0xff
	return DuRGBA((r*d)>>8, (g*d)>>8, (b*d)>>8, a)
}
