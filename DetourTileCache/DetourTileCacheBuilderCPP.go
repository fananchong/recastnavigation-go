//
// Copyright (c) 2009-2010 Mikko Mononen memon@inside.org
//
// This software is provided 'as-is', without any express or implied
// warranty.  In no event will the authors be held liable for any damages
// arising from the use of this software.
// Permission is granted to anyone to use this software for any purpose,
// including commercial applications, and to alter it and redistribute it
// freely, subject to the following restrictions:
// 1. The origin of this software must not be misrepresented; you must not
//    claim that you wrote the original software. If you use this software
//    in a product, an acknowledgment in the product documentation would be
//    appreciated but is not required.
// 2. Altered source versions must be plainly marked as such, and must not be
//    misrepresented as being the original software.
// 3. This notice may not be removed or altered from any source distribution.
//

package dtcache

import (
	"unsafe"

	detour "github.com/fananchong/recastnavigation-go/Detour"
)

var (
	offsetX = [4]int32{-1, 0, 1, 0}
	offsetY = [4]int32{0, 1, 0, -1}
)

func getDirOffsetX(dir int32) int32 {
	return offsetX[dir&0x03]
}

func getDirOffsetY(dir int32) int32 {
	return offsetY[dir&0x03]
}

const (
	MAX_VERTS_PER_POLY int32 = 6  // TODO: use the DT_VERTS_PER_POLYGON
	MAX_REM_EDGES      int32 = 48 // TODO: make this an expression.
)

func DtAllocTileCacheContourSet() *DtTileCacheContourSet {
	cset := &DtTileCacheContourSet{}
	return cset
}

func DtFreeTileCacheContourSet(cset *DtTileCacheContourSet) {

}

func DtAllocTileCachePolyMesh() *DtTileCachePolyMesh {
	lmesh := &DtTileCachePolyMesh{}
	return lmesh
}

func DtFreeTileCachePolyMesh(lmesh *DtTileCachePolyMesh) {

}

type dtLayerSweepSpan struct {
	ns  uint16 // number samples
	id  uint8  // region id
	nei uint8  // neighbour id
}

const DT_LAYER_MAX_NEIS int32 = 16

type dtLayerMonotoneRegion struct {
	area   int32
	neis   [DT_LAYER_MAX_NEIS]uint8
	nneis  uint8
	regId  uint8
	areaId uint8
}

type dtTempContour struct {
	verts  []uint8
	nverts int32
	cverts int32
	poly   []uint16
	npoly  int32
	cpoly  int32
}

func NewDtTempContour(vbuf []uint8, nvbuf int32, pbuf []uint16, npbuf int32) *dtTempContour {
	return &dtTempContour{
		verts:  vbuf,
		cverts: nvbuf,
		poly:   pbuf,
		cpoly:  npbuf,
	}
}

func overlapRangeExl(amin, amax, bmin, bmax uint16) bool {
	if amin >= bmax || amax <= bmin {
		return false
	}
	return true
}

func addUniqueLast(a []uint8, an *uint8, v uint8) {
	n := int32(*an)
	if n > 0 && a[n-1] == v {
		return
	}
	a[*an] = v
	(*an)++
}

func isConnected(layer *DtTileCacheLayer, ia, ib, walkableClimb int32) bool {
	if layer.areas[ia] != layer.areas[ib] {
		return false
	}
	if detour.DtAbsInt32(int32(layer.heights[ia])-int32(layer.heights[ib])) > walkableClimb {
		return false
	}
	return true
}

func canMerge(oldRegId, newRegId uint8, regs []dtLayerMonotoneRegion, nregs int32) bool {
	var count int32
	for i := int32(0); i < nregs; i++ {
		reg := &regs[i]
		if reg.regId != oldRegId {
			continue
		}
		nnei := int32(reg.nneis)
		for j := int32(0); j < nnei; j++ {
			if regs[reg.neis[j]].regId == newRegId {
				count++
			}
		}
	}
	return count == 1
}

func dtBuildTileCacheRegions(layer *DtTileCacheLayer, walkableClimb int32) detour.DtStatus {

	w := int32(layer.header.width)
	h := int32(layer.header.height)

	layer.regs = make([]uint8, w*h)
	for i := int32(0); i < w*h; i++ {
		layer.regs[i] = 0xff
	}

	nsweeps := w
	sweeps := make([]dtLayerSweepSpan, nsweeps)

	// Partition walkable area into monotone regions.
	var prevCount [256]uint8
	var regId uint8

	for y := int32(0); y < h; y++ {
		if regId > 0 {
			detour.MemsetUInt8(prevCount[:regId], 0)
		}
		var sweepId uint8

		for x := int32(0); x < w; x++ {
			idx := x + y*w
			if layer.areas[idx] == DT_TILECACHE_NULL_AREA {
				continue
			}

			sid := uint8(0xff)

			// -x
			xidx := (x - 1) + y*w
			if x > 0 && isConnected(layer, idx, xidx, walkableClimb) {
				if layer.regs[xidx] != 0xff {
					sid = layer.regs[xidx]
				}
			}

			if sid == 0xff {
				sid = sweepId
				sweepId++
				sweeps[sid].nei = 0xff
				sweeps[sid].ns = 0
			}

			// -y
			yidx := x + (y-1)*w
			if y > 0 && isConnected(layer, idx, yidx, walkableClimb) {
				nr := layer.regs[yidx]
				if nr != 0xff {
					// Set neighbour when first valid neighbour is encoutered.
					if sweeps[sid].ns == 0 {
						sweeps[sid].nei = nr
					}

					if sweeps[sid].nei == nr {
						// Update existing neighbour
						sweeps[sid].ns++
						prevCount[nr]++
					} else {
						// This is hit if there is nore than one neighbour.
						// Invalidate the neighbour.
						sweeps[sid].nei = 0xff
					}
				}
			}

			layer.regs[idx] = sid
		}

		// Create unique ID.
		for i := int32(0); i < int32(sweepId); i++ {
			// If the neighbour is set and there is only one continuous connection to it,
			// the sweep will be merged with the previous one, else new region is created.
			if sweeps[i].nei != 0xff && uint16(prevCount[sweeps[i].nei]) == sweeps[i].ns {
				sweeps[i].id = sweeps[i].nei
			} else {
				if regId == 255 {
					// Region ID's overflow.
					return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
				}
				sweeps[i].id = regId
				regId++
			}
		}

		// Remap local sweep ids to region ids.
		for x := int32(0); x < w; x++ {
			idx := x + y*w
			if layer.regs[idx] != 0xff {
				layer.regs[idx] = sweeps[layer.regs[idx]].id
			}
		}
	}

	// Allocate and init layer regions.
	nregs := int32(regId)
	regs := make([]dtLayerMonotoneRegion, nregs)

	for i := int32(0); i < nregs; i++ {
		regs[i].regId = 0xff
	}

	// Find region neighbours.
	for y := int32(0); y < h; y++ {
		for x := int32(0); x < w; x++ {
			idx := x + y*w
			ri := layer.regs[idx]
			if ri == 0xff {
				continue
			}

			// Update area.
			regs[ri].area++
			regs[ri].areaId = layer.areas[idx]

			// Update neighbours
			ymi := x + (y-1)*w
			if y > 0 && isConnected(layer, idx, ymi, walkableClimb) {
				rai := layer.regs[ymi]
				if rai != 0xff && rai != ri {
					addUniqueLast(regs[ri].neis[:], &regs[ri].nneis, rai)
					addUniqueLast(regs[rai].neis[:], &regs[rai].nneis, ri)
				}
			}
		}
	}

	for i := int32(0); i < nregs; i++ {
		regs[i].regId = uint8(i)
	}

	for i := int32(0); i < nregs; i++ {
		reg := &regs[i]

		merge := int32(-1)
		mergea := int32(0)
		for j := int32(0); j < int32(reg.nneis); j++ {
			nei := reg.neis[j]
			regn := &regs[nei]
			if reg.regId == regn.regId {
				continue
			}
			if reg.areaId != regn.areaId {
				continue
			}
			if regn.area > mergea {
				if canMerge(reg.regId, regn.regId, regs, nregs) {
					mergea = regn.area
					merge = int32(nei)
				}
			}
		}
		if merge != -1 {
			oldId := reg.regId
			newId := regs[merge].regId
			for j := int32(0); j < nregs; j++ {
				if regs[j].regId == oldId {
					regs[j].regId = newId
				}
			}
		}
	}

	// Compact ids.
	var remap [256]uint8
	detour.MemsetUInt8(remap[:], 0)
	// Find number of unique regions.
	regId = 0
	for i := int32(0); i < nregs; i++ {
		remap[regs[i].regId] = 1
	}
	for i := 0; i < 256; i++ {
		if remap[i] > 0 {
			remap[i] = regId
			regId++
		}
	}
	// Remap ids.
	for i := int32(0); i < nregs; i++ {
		regs[i].regId = remap[regs[i].regId]
	}

	layer.regCount = regId

	for i := int32(0); i < w*h; i++ {
		if layer.regs[i] != 0xff {
			layer.regs[i] = regs[layer.regs[i]].regId
		}
	}

	return detour.DT_SUCCESS
}

func appendVertex(cont *dtTempContour, x, y, z, r int32) bool {
	// Try to merge with existing segments.
	if cont.nverts > 1 {
		pa := cont.verts[(cont.nverts-2)*4:]
		pb := cont.verts[(cont.nverts-1)*4:]
		if int32(pb[3]) == r {
			if pa[0] == pb[0] && int32(pb[0]) == x {
				// The verts are aligned aling x-axis, update z.
				pb[1] = uint8(y)
				pb[2] = uint8(z)
				return true
			} else if pa[2] == pb[2] && int32(pb[2]) == z {
				// The verts are aligned aling z-axis, update x.
				pb[0] = uint8(x)
				pb[1] = uint8(y)
				return true
			}
		}
	}

	// Add new point.
	if cont.nverts+1 > cont.cverts {
		return false
	}

	v := cont.verts[cont.nverts*4:]
	v[0] = uint8(x)
	v[1] = uint8(y)
	v[2] = uint8(z)
	v[3] = uint8(r)
	cont.nverts++

	return true
}

func getNeighbourReg(layer *DtTileCacheLayer, ax, ay, dir int32) uint8 {
	w := int32(layer.header.width)
	ia := ax + ay*w

	con := layer.cons[ia] & 0xf
	portal := layer.cons[ia] >> 4
	mask := uint8(1 << uint32(dir))

	if (con & mask) == 0 {
		// No connection, return portal or hard edge.
		if portal&mask > 0 {
			return 0xf8 + uint8(dir)
		}
		return 0xff
	}

	bx := ax + getDirOffsetX(dir)
	by := ay + getDirOffsetY(dir)
	ib := bx + by*w

	return layer.regs[ib]
}

func walkContour(layer *DtTileCacheLayer, x, y int32, cont *dtTempContour) bool {
	w := int32(layer.header.width)
	h := int32(layer.header.height)

	cont.nverts = 0

	startX := x
	startY := y
	startDir := int32(-1)

	for i := int32(0); i < 4; i++ {
		dir := (i + 3) & 3
		rn := getNeighbourReg(layer, x, y, dir)
		if rn != layer.regs[x+y*w] {
			startDir = dir
			break
		}
	}
	if startDir == -1 {
		return true
	}

	dir := startDir
	maxIter := w * h

	var iter int32
	for iter < maxIter {
		rn := getNeighbourReg(layer, x, y, dir)

		nx := x
		ny := y
		ndir := dir

		if rn != layer.regs[x+y*w] {
			// Solid edge.
			px := x
			pz := y
			switch dir {
			case 0:
				pz++
			case 1:
				px++
			case 2:
				px++
			}

			// Try to merge with previous vertex.
			if !appendVertex(cont, px, int32(layer.heights[x+y*w]), pz, int32(rn)) {
				return false
			}

			ndir = (dir + 1) & 0x3 // Rotate CW
		} else {
			// Move to next.
			nx = x + getDirOffsetX(dir)
			ny = y + getDirOffsetY(dir)
			ndir = (dir + 3) & 0x3 // Rotate CCW
		}

		if iter > 0 && x == startX && y == startY && dir == startDir {
			break
		}

		x = nx
		y = ny
		dir = ndir

		iter++
	}

	// Remove last vertex if it is duplicate of the first one.
	pa := cont.verts[(cont.nverts-1)*4:]
	pb := cont.verts[0:]
	if pa[0] == pb[0] && pa[2] == pb[2] {
		cont.nverts--
	}

	return true
}

func distancePtSeg(x, z, px, pz, qx, qz int32) float32 {
	pqx := float32(qx - px)
	pqz := float32(qz - pz)
	dx := float32(x - px)
	dz := float32(z - pz)
	d := float32(pqx*pqx + pqz*pqz)
	t := float32(pqx*dx + pqz*dz)
	if d > 0 {
		t /= d
	}
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	dx = float32(px) + t*pqx - float32(x)
	dz = float32(pz) + t*pqz - float32(z)

	return dx*dx + dz*dz
}

func simplifyContour(cont *dtTempContour, maxError float32) {
	cont.npoly = 0

	for i := int32(0); i < cont.nverts; i++ {
		j := (i + 1) % cont.nverts
		// Check for start of a wall segment.
		ra := cont.verts[j*4+3]
		rb := cont.verts[i*4+3]
		if ra != rb {
			cont.poly[cont.npoly] = uint16(i)
			cont.npoly++
		}
	}
	if cont.npoly < 2 {
		// If there is no transitions at all,
		// create some initial points for the simplification process.
		// Find lower-left and upper-right vertices of the contour.
		llx := cont.verts[0]
		llz := cont.verts[2]
		lli := int32(0)
		urx := cont.verts[0]
		urz := cont.verts[2]
		uri := int32(0)
		for i := int32(1); i < cont.nverts; i++ {
			x := cont.verts[i*4+0]
			z := cont.verts[i*4+2]
			if x < llx || (x == llx && z < llz) {
				llx = x
				llz = z
				lli = i
			}
			if x > urx || (x == urx && z > urz) {
				urx = x
				urz = z
				uri = i
			}
		}
		cont.npoly = 0
		cont.poly[cont.npoly] = uint16(lli)
		cont.npoly++
		cont.poly[cont.npoly] = uint16(uri)
		cont.npoly++
	}

	// Add points until all raw points are within
	// error tolerance to the simplified shape.
	for i := int32(0); i < cont.npoly; {
		ii := (i + 1) % cont.npoly

		ai := int32(cont.poly[i])
		ax := int32(cont.verts[ai*4+0])
		az := int32(cont.verts[ai*4+2])

		bi := int32(cont.poly[ii])
		bx := int32(cont.verts[bi*4+0])
		bz := int32(cont.verts[bi*4+2])

		// Find maximum deviation from the segment.
		var maxd float32
		maxi := int32(-1)
		var ci, cinc, endi int32

		// Traverse the segment in lexilogical order so that the
		// max deviation is calculated similarly when traversing
		// opposite segments.
		if bx > ax || (bx == ax && bz > az) {
			cinc = 1
			ci = (ai + cinc) % cont.nverts
			endi = bi
		} else {
			cinc = cont.nverts - 1
			ci = (bi + cinc) % cont.nverts
			endi = ai
		}

		// Tessellate only outer edges or edges between areas.
		for ci != endi {
			d := distancePtSeg(int32(cont.verts[ci*4+0]), int32(cont.verts[ci*4+2]), ax, az, bx, bz)
			if d > maxd {
				maxd = d
				maxi = ci
			}
			ci = (ci + cinc) % cont.nverts
		}

		// If the max deviation is larger than accepted error,
		// add new point, else continue to next segment.
		if maxi != -1 && maxd > (maxError*maxError) {
			cont.npoly++
			for j := cont.npoly - 1; j > i; j-- {
				cont.poly[j] = cont.poly[j-1]
			}
			cont.poly[i+1] = uint16(maxi)
		} else {
			i++
		}
	}

	// Remap vertices
	var start int32
	for i := int32(1); i < cont.npoly; i++ {
		if cont.poly[i] < cont.poly[start] {
			start = i
		}
	}

	cont.nverts = 0
	for i := int32(0); i < cont.npoly; i++ {
		j := (start + i) % cont.npoly
		src := cont.verts[cont.poly[j]*4:]
		dst := cont.verts[cont.nverts*4:]
		dst[0] = src[0]
		dst[1] = src[1]
		dst[2] = src[2]
		dst[3] = src[3]
		cont.nverts++
	}
}

func getCornerHeight(layer *DtTileCacheLayer, x, y, z, walkableClimb int32) (shouldRemove bool, height uint8) {
	w := int32(layer.header.width)
	h := int32(layer.header.height)

	var n int32

	shouldRemove = false
	height = uint8(0)

	portal := uint8(0xf)
	preg := uint8(0xff)
	allSameReg := true

	for dz := int32(-1); dz <= 0; dz++ {
		for dx := int32(-1); dx <= 0; dx++ {
			px := x + dx
			pz := z + dz
			if px >= 0 && pz >= 0 && px < w && pz < h {
				idx := px + pz*w
				lh := int32(layer.heights[idx])
				if detour.DtAbsInt32(lh-y) <= walkableClimb && layer.areas[idx] != DT_TILECACHE_NULL_AREA {
					height = detour.DtMaxUInt8(height, uint8(lh))
					portal &= (layer.cons[idx] >> 4)
					if preg != 0xff && preg != layer.regs[idx] {
						allSameReg = false
					}
					preg = layer.regs[idx]
					n++
				}
			}
		}
	}

	var portalCount int32
	for dir := uint32(0); dir < 4; dir++ {
		if portal&(1<<dir) > 0 {
			portalCount++
		}
	}

	shouldRemove = false
	if n > 1 && portalCount == 1 && allSameReg {
		shouldRemove = true
	}

	return
}

// TODO: move this somewhere else, once the layer meshing is done.
func dtBuildTileCacheContours(layer *DtTileCacheLayer, walkableClimb int32, maxError float32, lcset *DtTileCacheContourSet) detour.DtStatus {
	w := int32(layer.header.width)
	h := int32(layer.header.height)

	lcset.nconts = int32(layer.regCount)
	lcset.conts = make([]DtTileCacheContour, lcset.nconts)

	// Allocate temp buffer for contour tracing.
	maxTempVerts := (w + h) * 2 * 2 // Twice around the layer.

	tempVerts := make([]uint8, maxTempVerts*4)

	tempPoly := make([]uint16, maxTempVerts)

	temp := NewDtTempContour(tempVerts, maxTempVerts, tempPoly, maxTempVerts)

	// Find contours.
	for y := int32(0); y < h; y++ {
		for x := int32(0); x < w; x++ {
			idx := x + y*w
			ri := layer.regs[idx]
			if ri == 0xff {
				continue
			}

			cont := &lcset.conts[ri]

			if cont.nverts > 0 {
				continue
			}

			cont.reg = ri
			cont.area = layer.areas[idx]

			if !walkContour(layer, x, y, temp) {
				// Too complex contour.
				// Note: If you hit here ofte, try increasing 'maxTempVerts'.
				return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
			}

			simplifyContour(temp, maxError)

			// Store contour.
			cont.nverts = temp.nverts
			if cont.nverts > 0 {
				cont.verts = make([]uint8, 4*temp.nverts)

				j := temp.nverts - 1
				for i := int32(0); i < temp.nverts; i++ {
					dst := cont.verts[j*4:]
					v := temp.verts[j*4:]
					vn := temp.verts[i*4:]
					nei := vn[3] // The neighbour reg is stored at segment vertex of a segment.
					shouldRemove, lh := getCornerHeight(layer, int32(v[0]), int32(v[1]), int32(v[2]), walkableClimb)

					dst[0] = v[0]
					dst[1] = lh
					dst[2] = v[2]

					// Store portal direction and remove status to the fourth component.
					dst[3] = 0x0f
					if nei != 0xff && nei >= 0xf8 {
						dst[3] = nei - 0xf8
					}
					if shouldRemove {
						dst[3] |= 0x80
					}
					j = i
				}
			}
		}
	}

	return detour.DT_SUCCESS
}

const (
	VERTEX_BUCKET_COUNT2 int32  = (1 << 8)
	h1                   uint32 = 0x8da6b343 // Large multiplicative constants;
	h2                   uint32 = 0xd8163841 // here arbitrarily chosen primes
	h3                   uint32 = 0xcb1ab31f
)

func computeVertexHash2(x, y, z int32) int32 {
	n := h1*uint32(x) + h2*uint32(y) + h3*uint32(z)
	return int32(n & uint32(VERTEX_BUCKET_COUNT2-1))
}

func addVertex(x, y, z uint16, verts, firstVert, nextVert []uint16, nv *int32) uint16 {
	bucket := computeVertexHash2(int32(x), 0, int32(z))
	i := firstVert[bucket]

	for i != DT_TILECACHE_NULL_IDX {
		v := verts[i*3:]
		if v[0] == x && v[2] == z && (detour.DtAbsInt32(int32(v[1])-int32(y)) <= 2) {
			return i
		}
		i = nextVert[i] // next
	}

	// Could not find, create new.
	i = uint16(*nv)
	(*nv)++
	v := verts[i*3:]
	v[0] = x
	v[1] = y
	v[2] = z
	nextVert[i] = firstVert[bucket]
	firstVert[bucket] = i

	return uint16(i)
}

type rcEdge struct {
	vert     [2]uint16
	polyEdge [2]uint16
	poly     [2]uint16
}

func buildMeshAdjacency(polys []uint16, npolys int32, verts []uint16, nverts int32, lcset *DtTileCacheContourSet) bool {
	// Based on code by Eric Lengyel from:
	// http://www.terathon.com/code/edges.php

	maxEdgeCount := npolys * MAX_VERTS_PER_POLY
	firstEdge := make([]uint16, nverts*maxEdgeCount)
	nextEdge := firstEdge[nverts:]

	var edgeCount int32

	edges := make([]rcEdge, maxEdgeCount)

	for i := int32(0); i < nverts; i++ {
		firstEdge[i] = DT_TILECACHE_NULL_IDX
	}

	for i := int32(0); i < npolys; i++ {
		t := polys[i*MAX_VERTS_PER_POLY*2:]
		for j := int32(0); j < MAX_VERTS_PER_POLY; j++ {
			if t[j] == DT_TILECACHE_NULL_IDX {
				break
			}
			v0 := t[j]
			v1 := t[j+1]
			if j+1 >= MAX_VERTS_PER_POLY || t[j+1] == DT_TILECACHE_NULL_IDX {
				v1 = t[0]
			}
			if v0 < v1 {
				edge := &edges[edgeCount]
				edge.vert[0] = v0
				edge.vert[1] = v1
				edge.poly[0] = uint16(i)
				edge.polyEdge[0] = uint16(j)
				edge.poly[1] = uint16(i)
				edge.polyEdge[1] = 0xff
				// Insert edge
				nextEdge[edgeCount] = firstEdge[v0]
				firstEdge[v0] = uint16(edgeCount)
				edgeCount++
			}
		}
	}

	for i := int32(0); i < npolys; i++ {
		t := polys[i*MAX_VERTS_PER_POLY*2:]
		for j := int32(0); j < MAX_VERTS_PER_POLY; j++ {
			if t[j] == DT_TILECACHE_NULL_IDX {
				break
			}
			v0 := t[j]
			v1 := t[j+1]
			if j+1 >= MAX_VERTS_PER_POLY || t[j+1] == DT_TILECACHE_NULL_IDX {
				v1 = t[0]
			}
			if v0 > v1 {
				found := false
				for e := firstEdge[v1]; e != DT_TILECACHE_NULL_IDX; e = nextEdge[e] {
					edge := &edges[e]
					if edge.vert[1] == v0 && edge.poly[0] == edge.poly[1] {
						edge.poly[1] = uint16(i)
						edge.polyEdge[1] = uint16(j)
						found = true
						break
					}
				}
				if !found {
					// Matching edge not found, it is an open edge, add it.
					edge := &edges[edgeCount]
					edge.vert[0] = v1
					edge.vert[1] = v0
					edge.poly[0] = uint16(i)
					edge.polyEdge[0] = uint16(j)
					edge.poly[1] = uint16(i)
					edge.polyEdge[1] = 0xff
					// Insert edge
					nextEdge[edgeCount] = firstEdge[v1]
					firstEdge[v1] = uint16(edgeCount)
					edgeCount++
				}
			}
		}
	}

	// Mark portal edges.
	for i := int32(0); i < lcset.nconts; i++ {
		cont := &lcset.conts[i]
		if cont.nverts < 3 {
			continue
		}

		k := cont.nverts - 1
		for j := int32(0); j < cont.nverts; j++ {
			va := cont.verts[k*4:]
			vb := cont.verts[j*4:]
			dir := va[3] & 0xf
			if dir == 0xf {
				continue
			}

			if dir == 0 || dir == 2 {
				// Find matching vertical edge
				x := uint16(va[0])
				zmin := uint16(va[2])
				zmax := uint16(vb[2])
				if zmin > zmax {
					zmin, zmax = zmax, zmin
					// detour.DtSwapUInt16(&zmin, &zmax)
				}

				for m := int32(0); m < edgeCount; m++ {
					e := &edges[m]
					// Skip connected edges.
					if e.poly[0] != e.poly[1] {
						continue
					}
					eva := verts[e.vert[0]*3:]
					evb := verts[e.vert[1]*3:]
					if eva[0] == x && evb[0] == x {
						ezmin := eva[2]
						ezmax := evb[2]
						if ezmin > ezmax {
							ezmin, ezmax = ezmax, ezmin
							// detour.DtSwapUInt16(&ezmin, &ezmax)
						}
						if overlapRangeExl(zmin, zmax, ezmin, ezmax) {
							// Reuse the other polyedge to store dir.
							e.polyEdge[1] = uint16(dir)
						}
					}
				}
			} else {
				// Find matching vertical edge
				z := uint16(va[2])
				xmin := uint16(va[0])
				xmax := uint16(vb[0])
				if xmin > xmax {
					xmin, xmax = xmax, xmin
					// detour.DtSwapUInt16(&xmin, &xmax)
				}
				for m := int32(0); m < edgeCount; m++ {
					e := &edges[m]
					// Skip connected edges.
					if e.poly[0] != e.poly[1] {
						continue
					}
					eva := verts[e.vert[0]*3:]
					evb := verts[e.vert[1]*3:]
					if eva[2] == z && evb[2] == z {
						exmin := eva[0]
						exmax := evb[0]
						if exmin > exmax {
							exmin, exmax = exmax, exmin
							// detour.DtSwapUInt16(&exmin, &exmax)
						}
						if overlapRangeExl(xmin, xmax, exmin, exmax) {
							// Reuse the other polyedge to store dir.
							e.polyEdge[1] = uint16(dir)
						}
					}
				}
			}
			k = j
		}
	}

	// Store adjacency
	for i := int32(0); i < edgeCount; i++ {
		e := &edges[i]
		if e.poly[0] != e.poly[1] {
			p0 := polys[int32(e.poly[0])*MAX_VERTS_PER_POLY*2:]
			p1 := polys[int32(e.poly[1])*MAX_VERTS_PER_POLY*2:]
			p0[MAX_VERTS_PER_POLY+int32(e.polyEdge[0])] = e.poly[1]
			p1[MAX_VERTS_PER_POLY+int32(e.polyEdge[1])] = e.poly[0]
		} else if e.polyEdge[1] != 0xff {
			p0 := polys[int32(e.poly[0])*MAX_VERTS_PER_POLY*2:]
			p0[MAX_VERTS_PER_POLY+int32(e.polyEdge[0])] = 0x8000 | uint16(e.polyEdge[1])
		}

	}

	return true
}

// Last time I checked the if version got compiled using cmov, which was a lot faster than module (with idiv).
func prev(i, n int32) int32 {
	if i-1 >= 0 {
		return i - 1
	}
	return n - 1
}
func next(i, n int32) int32 {
	if i+1 < n {
		return i + 1
	}
	return 0
}

func area2(a, b, c []uint8) int32 {
	return (int32(b[0])-int32(a[0]))*(int32(c[2])-int32(a[2])) - (int32(c[0])-int32(a[0]))*(int32(b[2])-int32(a[2]))
}

//	Exclusive or: true iff exactly one argument is true.
//	The arguments are negated to ensure that they are 0/1
//	values.  Then the bitwise Xor operator may apply.
//	(This idea is due to Michael Baldwin.)
func xorb(x, y bool) bool {
	//return !x ^ !y;
	if x == y {
		return false
	}
	return true
}

// Returns true iff c is strictly to the left of the directed
// line through a to b.
func left(a, b, c []uint8) bool {
	return area2(a, b, c) < 0
}

func leftOn(a, b, c []uint8) bool {
	return area2(a, b, c) <= 0
}

func collinear(a, b, c []uint8) bool {
	return area2(a, b, c) == 0
}

//	Returns true iff ab properly intersects cd: they share
//	a point interior to both segments.  The properness of the
//	intersection is ensured by using strict leftness.
func intersectProp(a, b, c, d []uint8) bool {
	// Eliminate improper cases.
	if collinear(a, b, c) || collinear(a, b, d) ||
		collinear(c, d, a) || collinear(c, d, b) {
		return false
	}

	return xorb(left(a, b, c), left(a, b, d)) && xorb(left(c, d, a), left(c, d, b))
}

// Returns T iff (a,b,c) are collinear and point c lies
// on the closed segement ab.
func between(a, b, c []uint8) bool {
	if !collinear(a, b, c) {
		return false
	}
	// If ab not vertical, check betweenness on x; else on y.
	if a[0] != b[0] {
		return ((a[0] <= c[0]) && (c[0] <= b[0])) || ((a[0] >= c[0]) && (c[0] >= b[0]))
	}
	return ((a[2] <= c[2]) && (c[2] <= b[2])) || ((a[2] >= c[2]) && (c[2] >= b[2]))

}

// Returns true iff segments ab and cd intersect, properly or improperly.
func intersect(a, b, c, d []uint8) bool {
	if intersectProp(a, b, c, d) {
		return true
	} else if between(a, b, c) || between(a, b, d) ||
		between(c, d, a) || between(c, d, b) {
		return true
	}

	return false
}

func vequal(a, b []uint8) bool {
	return a[0] == b[0] && a[2] == b[2]
}

// Returns T iff (v_i, v_j) is a proper internal *or* external
// diagonal of P, *ignoring edges incident to v_i and v_j*.
func diagonalie(i, j, n int32, verts []uint8, indices []uint16) bool {
	d0 := verts[(indices[i]&0x7fff)*4:]
	d1 := verts[(indices[j]&0x7fff)*4:]

	// For each edge (k,k+1) of P
	for k := int32(0); k < n; k++ {
		k1 := next(k, n)
		// Skip edges incident to i or j
		if !((k == i) || (k1 == i) || (k == j) || (k1 == j)) {
			p0 := verts[(indices[k]&0x7fff)*4:]
			p1 := verts[(indices[k1]&0x7fff)*4:]

			if vequal(d0, p0) || vequal(d1, p0) || vequal(d0, p1) || vequal(d1, p1) {
				continue
			}

			if intersect(d0, d1, p0, p1) {
				return false
			}
		}
	}
	return true
}

// Returns true iff the diagonal (i,j) is strictly internal to the
// polygon P in the neighborhood of the i endpoint.
func inCone(i, j, n int32, verts []uint8, indices []uint16) bool {
	pi := verts[(indices[i]&0x7fff)*4:]
	pj := verts[(indices[j]&0x7fff)*4:]
	pi1 := verts[(indices[next(i, n)]&0x7fff)*4:]
	pin1 := verts[(indices[prev(i, n)]&0x7fff)*4:]

	// If P[i] is a convex vertex [ i+1 left or on (i-1,i) ].
	if leftOn(pin1, pi, pi1) {
		return left(pi, pj, pin1) && left(pj, pi, pi1)
	}
	// Assume (i-1,i,i+1) not collinear.
	// else P[i] is reflex.
	return !(leftOn(pi, pj, pi1) && leftOn(pj, pi, pin1))
}

// Returns T iff (v_i, v_j) is a proper internal
// diagonal of P.
func diagonal(i, j, n int32, verts []uint8, indices []uint16) bool {
	return inCone(i, j, n, verts, indices) && diagonalie(i, j, n, verts, indices)
}

func triangulate(n int32, verts []uint8, indices, tris []uint16) int32 {
	var ntris int32
	dst := tris[:]

	// The last bit of the index is used to indicate if the vertex can be removed.
	for i := int32(0); i < n; i++ {
		i1 := next(i, n)
		i2 := next(i1, n)
		if diagonal(i, i2, n, verts, indices) {
			indices[i1] |= 0x8000
		}
	}

	for n > 3 {
		minLen := int32(-1)
		mini := int32(-1)
		for i := int32(0); i < n; i++ {
			i1 := next(i, n)
			if indices[i1]&0x8000 > 0 {
				p0 := verts[(indices[i]&0x7fff)*4:]
				p2 := verts[(indices[next(i1, n)]&0x7fff)*4:]

				dx := int32(p2[0]) - int32(p0[0])
				dz := int32(p2[2]) - int32(p0[2])
				len := dx*dx + dz*dz
				if minLen < 0 || len < minLen {
					minLen = len
					mini = i
				}
			}
		}

		if mini == -1 {
			// Should not happen.
			/*			printf("mini == -1 ntris=%d n=%d\n", ntris, n);
						for (int i = 0; i < n; i++)
						{
						printf("%d ", indices[i] & 0x0fffffff);
						}
						printf("\n");*/
			return -ntris
		}

		i := mini
		i1 := next(i, n)
		i2 := next(i1, n)

		dst[0] = indices[i] & 0x7fff
		dst[1] = indices[i1] & 0x7fff
		dst[2] = indices[i2] & 0x7fff
		dst = dst[3:]
		ntris++

		// Removes P[i1] by copying P[i+1]...P[n-1] left one index.
		n--
		for k := i1; k < n; k++ {
			indices[k] = indices[k+1]
		}

		if i1 >= n {
			i1 = 0
		}
		i = prev(i1, n)
		// Update diagonal flags.
		if diagonal(prev(i, n), i1, n, verts, indices) {
			indices[i] |= 0x8000
		} else {
			indices[i] &= 0x7fff
		}

		if diagonal(i, next(i1, n), n, verts, indices) {
			indices[i1] |= 0x8000
		} else {
			indices[i1] &= 0x7fff
		}
	}

	// Append the remaining triangle.
	dst[0] = indices[0] & 0x7fff
	dst[1] = indices[1] & 0x7fff
	dst[2] = indices[2] & 0x7fff
	dst = dst[3:]
	ntris++

	return ntris
}

func countPolyVerts(p []uint16) int32 {
	for i := int32(0); i < MAX_VERTS_PER_POLY; i++ {
		if p[i] == DT_TILECACHE_NULL_IDX {
			return i
		}
	}
	return MAX_VERTS_PER_POLY
}

func uleft(a, b, c []uint16) bool {
	return (int32(b[0])-int32(a[0]))*(int32(c[2])-int32(a[2]))-(int32(c[0])-int32(a[0]))*(int32(b[2])-int32(a[2])) < 0
}

func getPolyMergeValue(pa, pb, verts []uint16) (ea, eb, v int32) {
	na := countPolyVerts(pa)
	nb := countPolyVerts(pb)

	// If the merged polygon would be too big, do not merge.
	if na+nb-2 > MAX_VERTS_PER_POLY {
		v = -1
		return
	}

	// Check if the polygons share an edge.
	ea = -1
	eb = -1

	for i := int32(0); i < na; i++ {
		va0 := pa[i]
		va1 := pa[(i+1)%na]
		if va0 > va1 {
			va0, va1 = va1, va0
			// detour.DtSwapUInt16(&va0, &va1)
		}
		for j := int32(0); j < nb; j++ {
			vb0 := pb[j]
			vb1 := pb[(j+1)%nb]
			if vb0 > vb1 {
				vb0, vb1 = vb1, vb0
				// detour.DtSwapUInt16(&vb0, &vb1)
			}
			if va0 == vb0 && va1 == vb1 {
				ea = i
				eb = j
				break
			}
		}
	}

	// No common edge, cannot merge.
	if ea == -1 || eb == -1 {
		v = -1
		return
	}

	// Check to see if the merged polygon would be convex.
	var va, vb, vc uint16

	va = pa[(ea+na-1)%na]
	vb = pa[ea]
	vc = pb[(eb+2)%nb]
	if !uleft(verts[va*3:], verts[vb*3:], verts[vc*3:]) {
		v = -1
		return
	}

	va = pb[(eb+nb-1)%nb]
	vb = pb[eb]
	vc = pa[(ea+2)%na]
	if !uleft(verts[va*3:], verts[vb*3:], verts[vc*3:]) {
		v = -1
		return
	}

	va = pa[ea]
	vb = pa[(ea+1)%na]

	dx := int32(verts[va*3+0]) - int32(verts[vb*3+0])
	dy := int32(verts[va*3+2]) - int32(verts[vb*3+2])

	v = dx*dx + dy*dy
	return
}

func mergePolys(pa, pb []uint16, ea, eb int32) {
	var tmp [MAX_VERTS_PER_POLY * 2]uint16

	na := countPolyVerts(pa)
	nb := countPolyVerts(pb)

	// Merge polygons.
	detour.MemsetUInt16(tmp[:], 0xff)

	var n int32
	// Add pa
	for i := int32(0); i < na-1; i++ {
		tmp[n] = pa[(ea+1+i)%na]
		n++
	}
	// Add pb
	for i := int32(0); i < nb-1; i++ {
		tmp[n] = pb[(eb+1+i)%nb]
		n++
	}

	copy(pa, tmp[:MAX_VERTS_PER_POLY])
}

func pushFront(v uint16, arr []uint16, an *int32) {
	*an++
	for i := *an - 1; i > 0; i-- {
		arr[i] = arr[i-1]
	}
	arr[0] = v
}

func pushBack(v uint16, arr []uint16, an *int32) {
	arr[*an] = v
	*an++
}

func canRemoveVertex(mesh *DtTileCachePolyMesh, rem uint16) bool {
	// Count number of polygons to remove.
	var numRemovedVerts, numTouchedVerts, numRemainingEdges int32
	for i := int32(0); i < mesh.npolys; i++ {
		p := mesh.polys[i*MAX_VERTS_PER_POLY*2:]
		nv := countPolyVerts(p)
		var numRemoved, numVerts int32
		for j := int32(0); j < nv; j++ {
			if p[j] == rem {
				numTouchedVerts++
				numRemoved++
			}
			numVerts++
		}
		if numRemoved > 0 {
			numRemovedVerts += numRemoved
			numRemainingEdges += numVerts - (numRemoved + 1)
		}
	}

	// There would be too few edges remaining to create a polygon.
	// This can happen for example when a tip of a triangle is marked
	// as deletion, but there are no other polys that share the vertex.
	// In this case, the vertex should not be removed.
	if numRemainingEdges <= 2 {
		return false
	}

	// Check that there is enough memory for the test.
	maxEdges := numTouchedVerts * 2
	if maxEdges > MAX_REM_EDGES {
		return false
	}

	// Find edges which share the removed vertex.
	var edges [MAX_REM_EDGES]uint16
	var nedges int32

	for i := int32(0); i < mesh.npolys; i++ {
		p := mesh.polys[i*MAX_VERTS_PER_POLY*2:]
		nv := countPolyVerts(p)

		// Collect edges which touches the removed vertex.
		k := nv - 1
		for j := int32(0); j < nv; j++ {
			if p[j] == rem || p[k] == rem {
				// Arrange edge so that a=rem.
				a := p[j]
				b := p[k]
				if b == rem {
					a, b = b, a
					// detour.DtSwapUInt16(&a, &b)
				}

				// Check if the edge exists
				exists := false
				for m := int32(0); m < nedges; m++ {
					e := edges[m*3:]
					if e[1] == b {
						// Exists, increment vertex share count.
						e[2]++
						exists = true
					}
				}
				// Add new edge.
				if !exists {
					e := edges[nedges*3:]
					e[0] = uint16(a)
					e[1] = uint16(b)
					e[2] = 1
					nedges++
				}
			}
			k = j
		}
	}

	// There should be no more than 2 open edges.
	// This catches the case that two non-adjacent polygons
	// share the removed vertex. In that case, do not remove the vertex.
	var numOpenEdges int32
	for i := int32(0); i < nedges; i++ {
		if edges[i*3+2] < 2 {
			numOpenEdges++
		}
	}
	if numOpenEdges > 2 {
		return false
	}

	return true
}

func removeVertex(mesh *DtTileCachePolyMesh, rem uint16, maxTris int32) detour.DtStatus {
	// Count number of polygons to remove.
	var numRemovedVerts int32
	for i := int32(0); i < mesh.npolys; i++ {
		p := mesh.polys[i*MAX_VERTS_PER_POLY*2:]
		nv := countPolyVerts(p)
		for j := int32(0); j < nv; j++ {
			if p[j] == rem {
				numRemovedVerts++
			}
		}
	}

	var nedges, nhole, nharea int32
	var edges [MAX_REM_EDGES * 3]uint16
	var hole, harea [MAX_REM_EDGES]uint16

	for i := int32(0); i < mesh.npolys; i++ {
		p := mesh.polys[i*MAX_VERTS_PER_POLY*2:]
		nv := countPolyVerts(p)
		hasRem := false
		for j := int32(0); j < nv; j++ {
			if p[j] == rem {
				hasRem = true
			}
		}
		if hasRem {
			// Collect edges which does not touch the removed vertex.
			k := nv - 1
			for j := int32(0); j < nv; j++ {
				if p[j] != rem && p[k] != rem {
					if nedges >= MAX_REM_EDGES {
						return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
					}
					e := edges[nedges*3:]
					e[0] = p[k]
					e[1] = p[j]
					e[2] = uint16(mesh.areas[i])
					nedges++
				}
				k = j
			}
			// Remove the polygon.
			p2 := mesh.polys[(mesh.npolys-1)*MAX_VERTS_PER_POLY*2:]
			copy(p, p2[:MAX_VERTS_PER_POLY])
			detour.MemsetUInt16(p[MAX_VERTS_PER_POLY:MAX_VERTS_PER_POLY*2], 0xff)

			mesh.areas[i] = mesh.areas[mesh.npolys-1]
			mesh.npolys--
			i--
		}
	}

	// Remove vertex.
	for i := int32(rem); i < mesh.nverts; i++ {
		mesh.verts[i*3+0] = mesh.verts[(i+1)*3+0]
		mesh.verts[i*3+1] = mesh.verts[(i+1)*3+1]
		mesh.verts[i*3+2] = mesh.verts[(i+1)*3+2]
	}
	mesh.nverts--

	// Adjust indices to match the removed vertex layout.
	for i := int32(0); i < mesh.npolys; i++ {
		p := mesh.polys[i*MAX_VERTS_PER_POLY*2:]
		nv := countPolyVerts(p)
		for j := int32(0); j < nv; j++ {
			if p[j] > rem {
				p[j]--
			}
		}
	}
	for i := int32(0); i < nedges; i++ {
		if edges[i*3+0] > rem {
			edges[i*3+0]--
		}
		if edges[i*3+1] > rem {
			edges[i*3+1]--
		}
	}

	if nedges == 0 {
		return detour.DT_SUCCESS
	}

	// Start with one vertex, keep appending connected
	// segments to the start and end of the hole.
	pushBack(edges[0], hole[:], &nhole)
	pushBack(edges[2], harea[:], &nharea)

	for nedges > 0 {
		match := false

		for i := int32(0); i < nedges; i++ {
			ea := edges[i*3+0]
			eb := edges[i*3+1]
			a := edges[i*3+2]
			add := false
			if hole[0] == eb {
				// The segment matches the beginning of the hole boundary.
				if nhole >= MAX_REM_EDGES {
					return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
				}
				pushFront(ea, hole[:], &nhole)
				pushFront(a, harea[:], &nharea)
				add = true
			} else if hole[nhole-1] == ea {
				// The segment matches the end of the hole boundary.
				if nhole >= MAX_REM_EDGES {
					return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
				}
				pushBack(eb, hole[:], &nhole)
				pushBack(a, harea[:], &nharea)
				add = true
			}
			if add {
				// The edge segment was added, remove it.
				edges[i*3+0] = edges[(nedges-1)*3+0]
				edges[i*3+1] = edges[(nedges-1)*3+1]
				edges[i*3+2] = edges[(nedges-1)*3+2]
				nedges--
				match = true
				i--
			}
		}

		if !match {
			break
		}
	}

	var tris [MAX_REM_EDGES * 3]uint16
	var tverts [MAX_REM_EDGES * 3]uint8
	var tpoly [MAX_REM_EDGES * 3]uint16

	// Generate temp vertex array for triangulation.
	for i := int32(0); i < nhole; i++ {
		pi := hole[i]
		tverts[i*4+0] = uint8(mesh.verts[pi*3+0])
		tverts[i*4+1] = uint8(mesh.verts[pi*3+1])
		tverts[i*4+2] = uint8(mesh.verts[pi*3+2])
		tverts[i*4+3] = 0
		tpoly[i] = uint16(i)
	}

	// Triangulate the hole.
	ntris := triangulate(nhole, tverts[:], tpoly[:], tris[:])
	if ntris < 0 {
		// TODO: issue warning!
		ntris = -ntris
	}

	if ntris > MAX_REM_EDGES {
		return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
	}

	var polys [MAX_REM_EDGES * MAX_VERTS_PER_POLY]uint16
	var pareas [MAX_REM_EDGES]uint8

	// Build initial polygons.
	var npolys int32
	detour.MemsetUInt16(polys[:ntris*MAX_VERTS_PER_POLY], 0xff)
	for j := int32(0); j < ntris; j++ {
		t := tris[j*3:]
		if t[0] != t[1] && t[0] != t[2] && t[1] != t[2] {
			polys[npolys*MAX_VERTS_PER_POLY+0] = hole[t[0]]
			polys[npolys*MAX_VERTS_PER_POLY+1] = hole[t[1]]
			polys[npolys*MAX_VERTS_PER_POLY+2] = hole[t[2]]
			pareas[npolys] = uint8(harea[t[0]])
			npolys++
		}
	}
	if npolys == 0 {
		return detour.DT_SUCCESS
	}

	// Merge polygons.
	var maxVertsPerPoly int32 = MAX_VERTS_PER_POLY
	if maxVertsPerPoly > 3 {
		for {
			// Find best polygons to merge.
			var bestMergeVal, bestPa, bestPb, bestEa, bestEb int32

			for j := int32(0); j < npolys-1; j++ {
				pj := polys[j*MAX_VERTS_PER_POLY:]
				for k := j + 1; k < npolys; k++ {
					pk := polys[k*MAX_VERTS_PER_POLY:]
					ea, eb, v := getPolyMergeValue(pj, pk, mesh.verts)
					if v > bestMergeVal {
						bestMergeVal = v
						bestPa = j
						bestPb = k
						bestEa = ea
						bestEb = eb
					}
				}
			}

			if bestMergeVal > 0 {
				// Found best, merge.
				pa := polys[bestPa*MAX_VERTS_PER_POLY:]
				pb := polys[bestPb*MAX_VERTS_PER_POLY:]
				mergePolys(pa, pb, bestEa, bestEb)
				copy(pb, polys[(npolys-1)*MAX_VERTS_PER_POLY:(npolys-1)*MAX_VERTS_PER_POLY+MAX_VERTS_PER_POLY])
				pareas[bestPb] = pareas[npolys-1]
				npolys--
			} else {
				// Could not merge any polygons, stop.
				break
			}
		}
	}

	// Store polygons.
	for i := int32(0); i < npolys; i++ {
		if mesh.npolys >= maxTris {
			break
		}
		p := mesh.polys[mesh.npolys*MAX_VERTS_PER_POLY*2:]
		detour.MemsetUInt16(p[:MAX_VERTS_PER_POLY*2], 0xff)
		for j := int32(0); j < MAX_VERTS_PER_POLY; j++ {
			p[j] = polys[i*MAX_VERTS_PER_POLY+j]
		}
		mesh.areas[mesh.npolys] = pareas[i]
		mesh.npolys++
		if mesh.npolys > maxTris {
			return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
		}
	}

	return detour.DT_SUCCESS
}

func dtBuildTileCachePolyMesh(lcset *DtTileCacheContourSet, mesh *DtTileCachePolyMesh) detour.DtStatus {
	var maxVertices, maxTris, maxVertsPerCont int32
	for i := int32(0); i < lcset.nconts; i++ {
		// Skip null contours.
		if lcset.conts[i].nverts < 3 {
			continue
		}
		maxVertices += lcset.conts[i].nverts
		maxTris += lcset.conts[i].nverts - 2
		maxVertsPerCont = detour.DtMaxInt32(maxVertsPerCont, lcset.conts[i].nverts)
	}

	// TODO: warn about too many vertices?

	mesh.nvp = MAX_VERTS_PER_POLY

	vflags := make([]uint8, maxVertices)

	mesh.verts = make([]uint16, maxVertices*3)

	mesh.polys = make([]uint16, maxTris*MAX_VERTS_PER_POLY*2)

	mesh.areas = make([]uint8, maxTris)

	mesh.flags = make([]uint16, maxTris)

	mesh.nverts = 0
	mesh.npolys = 0

	detour.MemsetUInt16(mesh.polys[:], 0xff)

	var firstVert [VERTEX_BUCKET_COUNT2]uint16
	detour.MemsetUInt16(firstVert[:], DT_TILECACHE_NULL_IDX)

	nextVert := make([]uint16, maxVertices)
	indices := make([]uint16, maxVertsPerCont)
	tris := make([]uint16, maxVertsPerCont*3)
	polys := make([]uint16, maxVertsPerCont*MAX_VERTS_PER_POLY)

	for i := int32(0); i < lcset.nconts; i++ {
		cont := &lcset.conts[i]

		// Skip null contours.
		if cont.nverts < 3 {
			continue
		}

		// Triangulate contour
		for j := int32(0); j < cont.nverts; j++ {
			indices[j] = uint16(j)
		}

		ntris := triangulate(cont.nverts, cont.verts, indices[:], tris[:])
		if ntris <= 0 {
			// TODO: issue warning!
			ntris = -ntris
		}

		// Add and merge vertices.
		for j := int32(0); j < cont.nverts; j++ {
			v := cont.verts[j*4:]
			indices[j] = addVertex(uint16(v[0]), uint16(v[1]), uint16(v[2]),
				mesh.verts, firstVert[:], nextVert[:], &mesh.nverts)
			if v[3]&0x80 > 0 {
				// This vertex should be removed.
				vflags[indices[j]] = 1
			}
		}

		// Build initial polygons.
		var npolys int32
		detour.MemsetUInt16(polys[:maxVertsPerCont*MAX_VERTS_PER_POLY], 0xff)
		for j := int32(0); j < ntris; j++ {
			t := tris[j*3:]
			if t[0] != t[1] && t[0] != t[2] && t[1] != t[2] {
				polys[npolys*MAX_VERTS_PER_POLY+0] = indices[t[0]]
				polys[npolys*MAX_VERTS_PER_POLY+1] = indices[t[1]]
				polys[npolys*MAX_VERTS_PER_POLY+2] = indices[t[2]]
				npolys++
			}
		}
		if npolys == 0 {
			continue
		}

		// Merge polygons.
		maxVertsPerPoly := MAX_VERTS_PER_POLY
		if maxVertsPerPoly > 3 {
			for {
				// Find best polygons to merge.\
				var bestMergeVal, bestPa, bestPb, bestEa, bestEb int32

				for j := int32(0); j < npolys-1; j++ {
					pj := polys[j*MAX_VERTS_PER_POLY:]
					for k := j + 1; k < npolys; k++ {
						pk := polys[k*MAX_VERTS_PER_POLY:]
						ea, eb, v := getPolyMergeValue(pj, pk, mesh.verts)
						if v > bestMergeVal {
							bestMergeVal = v
							bestPa = j
							bestPb = k
							bestEa = ea
							bestEb = eb
						}
					}
				}

				if bestMergeVal > 0 {
					// Found best, merge.
					pa := polys[bestPa*MAX_VERTS_PER_POLY:]
					pb := polys[bestPb*MAX_VERTS_PER_POLY:]
					mergePolys(pa, pb, bestEa, bestEb)
					copy(pb, polys[(npolys-1)*MAX_VERTS_PER_POLY:(npolys-1)*MAX_VERTS_PER_POLY+MAX_VERTS_PER_POLY])
					npolys--
				} else {
					// Could not merge any polygons, stop.
					break
				}
			}
		}

		// Store polygons.
		for j := int32(0); j < npolys; j++ {
			p := mesh.polys[mesh.npolys*MAX_VERTS_PER_POLY*2:]
			q := polys[j*MAX_VERTS_PER_POLY:]
			for k := int32(0); k < MAX_VERTS_PER_POLY; k++ {
				p[k] = q[k]
			}

			mesh.areas[mesh.npolys] = cont.area
			mesh.npolys++
			if mesh.npolys > maxTris {
				return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
			}
		}
	}

	// Remove edge vertices.
	for i := int32(0); i < mesh.nverts; i++ {
		if vflags[i] > 0 {
			if !canRemoveVertex(mesh, uint16(i)) {
				continue
			}
			status := removeVertex(mesh, uint16(i), maxTris)
			if detour.DtStatusFailed(status) {
				return status
			}
			// Remove vertex
			// Note: mesh.nverts is already decremented inside removeVertex()!
			for j := i; j < mesh.nverts; j++ {
				vflags[j] = vflags[j+1]
			}
			i--
		}
	}

	// Calculate adjacency.
	if !buildMeshAdjacency(mesh.polys, mesh.npolys, mesh.verts, mesh.nverts, lcset) {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	return detour.DT_SUCCESS
}

func dtMarkCylinderArea(layer *DtTileCacheLayer, orig []float32, cs, ch float32, pos []float32, radius, height float32, areaId uint8) detour.DtStatus {
	var bmin, bmax [3]float32
	bmin[0] = pos[0] - radius
	bmin[1] = pos[1]
	bmin[2] = pos[2] - radius
	bmax[0] = pos[0] + radius
	bmax[1] = pos[1] + height
	bmax[2] = pos[2] + radius
	r2 := detour.DtSqrFloat32(radius/cs + 0.5)

	w := int32(layer.header.width)
	h := int32(layer.header.height)
	ics := 1.0 / cs
	ich := 1.0 / ch

	px := (pos[0] - orig[0]) * ics
	pz := (pos[2] - orig[2]) * ics

	minx := int32(detour.DtMathFloorf((bmin[0] - orig[0]) * ics))
	miny := int32(detour.DtMathFloorf((bmin[1] - orig[1]) * ich))
	minz := int32(detour.DtMathFloorf((bmin[2] - orig[2]) * ics))
	maxx := int32(detour.DtMathFloorf((bmax[0] - orig[0]) * ics))
	maxy := int32(detour.DtMathFloorf((bmax[1] - orig[1]) * ich))
	maxz := int32(detour.DtMathFloorf((bmax[2] - orig[2]) * ics))

	if maxx < 0 {
		return detour.DT_SUCCESS
	}
	if minx >= w {
		return detour.DT_SUCCESS
	}
	if maxz < 0 {
		return detour.DT_SUCCESS
	}
	if minz >= h {
		return detour.DT_SUCCESS
	}

	if minx < 0 {
		minx = 0
	}
	if maxx >= w {
		maxx = w - 1
	}
	if minz < 0 {
		minz = 0
	}
	if maxz >= h {
		maxz = h - 1
	}

	for z := minz; z <= maxz; z++ {
		for x := minx; x <= maxx; x++ {
			dx := float32(x) + 0.5 - px
			dz := float32(z) + 0.5 - pz
			if dx*dx+dz*dz > r2 {
				continue
			}
			y := int32(layer.heights[x+z*w])
			if y < miny || y > maxy {
				continue
			}
			layer.areas[x+z*w] = areaId
		}
	}

	return detour.DT_SUCCESS
}

func dtMarkBoxArea1(layer *DtTileCacheLayer, orig []float32, cs, ch float32, bmin, bmax []float32, areaId uint8) detour.DtStatus {
	w := int32(layer.header.width)
	h := int32(layer.header.height)
	ics := 1.0 / cs
	ich := 1.0 / ch

	minx := int32(detour.DtMathFloorf((bmin[0] - orig[0]) * ics))
	miny := int32(detour.DtMathFloorf((bmin[1] - orig[1]) * ich))
	minz := int32(detour.DtMathFloorf((bmin[2] - orig[2]) * ics))
	maxx := int32(detour.DtMathFloorf((bmax[0] - orig[0]) * ics))
	maxy := int32(detour.DtMathFloorf((bmax[1] - orig[1]) * ich))
	maxz := int32(detour.DtMathFloorf((bmax[2] - orig[2]) * ics))

	if maxx < 0 {
		return detour.DT_SUCCESS
	}
	if minx >= w {
		return detour.DT_SUCCESS
	}
	if maxz < 0 {
		return detour.DT_SUCCESS
	}
	if minz >= h {
		return detour.DT_SUCCESS
	}

	if minx < 0 {
		minx = 0
	}
	if maxx >= w {
		maxx = w - 1
	}
	if minz < 0 {
		minz = 0
	}
	if maxz >= h {
		maxz = h - 1
	}

	for z := minz; z <= maxz; z++ {
		for x := minx; x <= maxx; x++ {
			y := int32(layer.heights[x+z*w])
			if y < miny || y > maxy {
				continue
			}
			layer.areas[x+z*w] = areaId
		}
	}

	return detour.DT_SUCCESS
}

func dtMarkBoxArea2(layer *DtTileCacheLayer, orig []float32, cs, ch float32, center, halfExtents, rotAux []float32, areaId uint8) detour.DtStatus {
	w := int32(layer.header.width)
	h := int32(layer.header.height)
	ics := 1.0 / cs
	ich := 1.0 / ch

	cx := (center[0] - orig[0]) * ics
	cz := (center[2] - orig[2]) * ics

	maxr := 1.41 * detour.DtMaxFloat32(halfExtents[0], halfExtents[2])
	minx := int32(detour.DtMathFloorf(cx - maxr*ics))
	maxx := int32(detour.DtMathFloorf(cx + maxr*ics))
	minz := int32(detour.DtMathFloorf(cz - maxr*ics))
	maxz := int32(detour.DtMathFloorf(cz + maxr*ics))
	miny := int32(detour.DtMathFloorf((center[1] - halfExtents[1] - orig[1]) * ich))
	maxy := int32(detour.DtMathFloorf((center[1] + halfExtents[1] - orig[1]) * ich))

	if maxx < 0 {
		return detour.DT_SUCCESS
	}
	if minx >= w {
		return detour.DT_SUCCESS
	}
	if maxz < 0 {
		return detour.DT_SUCCESS
	}
	if minz >= h {
		return detour.DT_SUCCESS
	}
	if minx < 0 {
		minx = 0
	}
	if maxx >= w {
		maxx = w - 1
	}
	if minz < 0 {
		minz = 0
	}
	if maxz >= h {
		maxz = h - 1
	}

	xhalf := halfExtents[0]*ics + 0.5
	zhalf := halfExtents[2]*ics + 0.5

	for z := minz; z <= maxz; z++ {
		for x := minx; x <= maxx; x++ {
			x2 := 2.0 * (float32(x) - cx)
			z2 := 2.0 * (float32(z) - cz)
			xrot := rotAux[1]*x2 + rotAux[0]*z2
			if xrot > xhalf || xrot < -xhalf {
				continue
			}
			zrot := rotAux[1]*z2 - rotAux[0]*x2
			if zrot > zhalf || zrot < -zhalf {
				continue
			}
			y := int32(layer.heights[x+z*w])
			if y < miny || y > maxy {
				continue
			}
			layer.areas[x+z*w] = areaId
		}
	}

	return detour.DT_SUCCESS
}

func dtBuildTileCacheLayer(comp DtTileCacheCompressor, header *DtTileCacheLayerHeader, heights, areas, cons []uint8) (status detour.DtStatus, outData []uint8, outDataSize int32) {
	headerSize := int32(detour.DtAlign4(int(unsafe.Sizeof(DtTileCacheLayerHeader{}))))
	gridSize := int32(header.width) * int32(header.height)
	maxDataSize := headerSize + comp.MaxCompressedSize(gridSize*3)
	data := make([]uint8, maxDataSize)

	// Store header
	*(*DtTileCacheLayerHeader)(unsafe.Pointer(&data[0])) = *header

	// Concatenate grid data for compression.
	bufferSize := gridSize * 3
	buffer := make([]uint8, bufferSize)

	copy(buffer[:gridSize], heights)
	copy(buffer[gridSize:gridSize*2], areas)
	copy(buffer[gridSize*2:gridSize*3], cons)

	// Compress
	compressed := data[headerSize:]
	maxCompressedSize := maxDataSize - headerSize
	var compressedSize int32
	compressedSize, status = comp.Compress(buffer, bufferSize, compressed, maxCompressedSize)
	if detour.DtStatusFailed(status) {
		return status, nil, 0
	}

	outData = data
	outDataSize = headerSize + compressedSize

	return
}

func dtFreeTileCacheLayer(layer *DtTileCacheLayer) {

}

func dtDecompressTileCacheLayer(comp DtTileCacheCompressor, compressed []uint8, compressedSize int32) (status detour.DtStatus, layerOut *DtTileCacheLayer) {
	if compressed == nil {
		status = detour.DT_FAILURE | detour.DT_INVALID_PARAM
		return
	}

	layerOut = nil

	compressedHeader := (*DtTileCacheLayerHeader)(unsafe.Pointer(&compressed[0]))
	if compressedHeader.magic != DT_TILECACHE_MAGIC {
		status = detour.DT_FAILURE | detour.DT_WRONG_MAGIC
		return
	}
	if compressedHeader.version != DT_TILECACHE_VERSION {
		status = detour.DT_FAILURE | detour.DT_WRONG_VERSION
		return
	}

	headerSize := int32(detour.DtAlign4(int(unsafe.Sizeof(DtTileCacheLayerHeader{}))))
	gridSize := int32(compressedHeader.width) * int32(compressedHeader.height)

	header := &DtTileCacheLayerHeader{}
	layerOut = &DtTileCacheLayer{}
	grids := make([]uint8, gridSize*4)

	// Copy header
	*header = *compressedHeader
	// Decompress grid.
	// var size int32
	_, status = comp.Decompress(compressed[headerSize:], compressedSize-headerSize, grids, gridSize*4)
	if detour.DtStatusFailed(status) {
		return
	}

	layerOut.header = header
	layerOut.heights = grids
	layerOut.areas = grids[gridSize:]
	layerOut.cons = grids[gridSize*2:]
	layerOut.regs = grids[gridSize*3:]

	status = detour.DT_SUCCESS
	return
}

func dtTileCacheHeaderSwapEndian(data []uint8, dataSize int32) bool {
	// dtIgnoreUnused(dataSize)
	header := (*DtTileCacheLayerHeader)(unsafe.Pointer(&data[0]))

	swappedMagic := DT_TILECACHE_MAGIC
	swappedVersion := DT_TILECACHE_VERSION
	detour.DtSwapEndianInt32(&swappedMagic)
	detour.DtSwapEndianInt32(&swappedVersion)

	if (header.magic != DT_TILECACHE_MAGIC || header.version != DT_TILECACHE_VERSION) &&
		(header.magic != swappedMagic || header.version != swappedVersion) {
		return false
	}

	detour.DtSwapEndianInt32(&header.magic)
	detour.DtSwapEndianInt32(&header.version)
	detour.DtSwapEndianInt32(&header.tx)
	detour.DtSwapEndianInt32(&header.ty)
	detour.DtSwapEndianInt32(&header.tlayer)
	detour.DtSwapEndianFloat32(&header.bmin[0])
	detour.DtSwapEndianFloat32(&header.bmin[1])
	detour.DtSwapEndianFloat32(&header.bmin[2])
	detour.DtSwapEndianFloat32(&header.bmax[0])
	detour.DtSwapEndianFloat32(&header.bmax[1])
	detour.DtSwapEndianFloat32(&header.bmax[2])
	detour.DtSwapEndianUInt16(&header.hmin)
	detour.DtSwapEndianUInt16(&header.hmax)

	// width, height, minx, maxx, miny, maxy are unsigned char, no need to swap.

	return true
}
