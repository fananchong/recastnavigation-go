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
	"reflect"
	"unsafe"

	detour "github.com/fananchong/recastnavigation-go/Detour"
)

var offsetX = [4]int32{-1, 0, 1, 0}

func getDirOffsetX(dir int32) int32 {
	return offsetX[dir&0x03]
}

var offsetY = [4]int32{0, 1, 0, -1}

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
	if cset == nil {
		return
	}
	cset.Conts = nil
	cset.Nconts = 0
}

func DtAllocTileCachePolyMesh() *DtTileCachePolyMesh {
	lmesh := &DtTileCachePolyMesh{}
	return lmesh
}

func DtFreeTileCachePolyMesh(lmesh *DtTileCachePolyMesh) {
	if lmesh == nil {
		return
	}
	lmesh.Verts = nil
	lmesh.Nverts = 0
	lmesh.Polys = nil
	lmesh.Npolys = 0
	lmesh.Flags = nil
	lmesh.Areas = nil
}

type dtLayerSweepSpan struct {
	ns  uint16 // number samples
	id  uint8  // region id
	nei uint8  // neighbour id
}

const DT_LAYER_MAX_NEIS int32 = 16

type dtLayerMonotoneRegion struct {
	Area   int32
	Neis   [DT_LAYER_MAX_NEIS]uint8
	Nneis  uint8
	RegId  uint8
	AreaId uint8
}

type dtTempContour struct {
	Verts  []uint8
	Nverts int32
	Cverts int32
	Poly   []uint16
	Npoly  int32
	Cpoly  int32
}

func (this *dtTempContour) init(vbuf []uint8, nvbuf int32, pbuf []uint16, npbuf int32) {
	this.Verts = vbuf
	this.Nverts = 0
	this.Cverts = nvbuf
	this.Poly = pbuf
	this.Npoly = 0
	this.Cpoly = npbuf

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
	if layer.Areas[ia] != layer.Areas[ib] {
		return false
	}
	if detour.DtAbsInt32(int32(layer.Heights[ia])-int32(layer.Heights[ib])) > walkableClimb {
		return false
	}
	return true
}

func canMerge(oldRegId, newRegId uint8, regs []dtLayerMonotoneRegion, nregs int32) bool {
	var count int32
	for i := int32(0); i < nregs; i++ {
		reg := &regs[i]
		if reg.RegId != oldRegId {
			continue
		}
		nnei := int32(reg.Nneis)
		for j := int32(0); j < nnei; j++ {
			if regs[reg.Neis[j]].RegId == newRegId {
				count++
			}
		}
	}
	return count == 1
}

func DtBuildTileCacheRegions(layer *DtTileCacheLayer, walkableClimb int32) detour.DtStatus {

	w := int32(layer.Header.Width)
	h := int32(layer.Header.Height)

	layer.Regs = make([]uint8, w*h)
	for i := int32(0); i < w*h; i++ {
		layer.Regs[i] = 0xff
	}

	nsweeps := w
	sweeps := make([]dtLayerSweepSpan, nsweeps)

	// Partition walkable area into monotone regions.
	var prevCount [256]uint8
	var regId uint8

	for y := int32(0); y < h; y++ {
		if regId > 0 {
			detour.Memset(uintptr(unsafe.Pointer(&(prevCount[0]))), 0, int(regId))
		}
		var sweepId uint8

		for x := int32(0); x < w; x++ {
			idx := x + y*w
			if layer.Areas[idx] == DT_TILECACHE_NULL_AREA {
				continue
			}

			sid := uint8(0xff)

			// -x
			xidx := (x - 1) + y*w
			if x > 0 && isConnected(layer, idx, xidx, walkableClimb) {
				if layer.Regs[xidx] != 0xff {
					sid = layer.Regs[xidx]
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
				nr := layer.Regs[yidx]
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

			layer.Regs[idx] = sid
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
			if layer.Regs[idx] != 0xff {
				layer.Regs[idx] = sweeps[layer.Regs[idx]].id
			}
		}
	}

	// Allocate and init layer regions.
	nregs := int32(regId)
	regs := make([]dtLayerMonotoneRegion, nregs)

	for i := int32(0); i < nregs; i++ {
		regs[i].RegId = 0xff
	}

	// Find region neighbours.
	for y := int32(0); y < h; y++ {
		for x := int32(0); x < w; x++ {
			idx := x + y*w
			ri := layer.Regs[idx]
			if ri == 0xff {
				continue
			}

			// Update area.
			regs[ri].Area++
			regs[ri].AreaId = layer.Areas[idx]

			// Update neighbours
			ymi := x + (y-1)*w
			if y > 0 && isConnected(layer, idx, ymi, walkableClimb) {
				rai := layer.Regs[ymi]
				if rai != 0xff && rai != ri {
					addUniqueLast(regs[ri].Neis[:], &regs[ri].Nneis, rai)
					addUniqueLast(regs[rai].Neis[:], &regs[rai].Nneis, ri)
				}
			}
		}
	}

	for i := int32(0); i < nregs; i++ {
		regs[i].RegId = uint8(i)
	}

	for i := int32(0); i < nregs; i++ {
		reg := &regs[i]

		merge := int32(-1)
		mergea := int32(0)
		for j := int32(0); j < int32(reg.Nneis); j++ {
			nei := reg.Neis[j]
			regn := &regs[nei]
			if reg.RegId == regn.RegId {
				continue
			}
			if reg.AreaId != regn.AreaId {
				continue
			}
			if regn.Area > mergea {
				if canMerge(reg.RegId, regn.RegId, regs, nregs) {
					mergea = regn.Area
					merge = int32(nei)
				}
			}
		}
		if merge != -1 {
			oldId := reg.RegId
			newId := regs[merge].RegId
			for j := int32(0); j < nregs; j++ {
				if regs[j].RegId == oldId {
					regs[j].RegId = newId
				}
			}
		}
	}

	// Compact ids.
	var remap [256]uint8
	// Find number of unique regions.
	regId = 0
	for i := int32(0); i < nregs; i++ {
		remap[regs[i].RegId] = 1
	}
	for i := 0; i < 256; i++ {
		if remap[i] > 0 {
			remap[i] = regId
			regId++
		}
	}
	// Remap ids.
	for i := int32(0); i < nregs; i++ {
		regs[i].RegId = remap[regs[i].RegId]
	}

	layer.RegCount = regId

	for i := int32(0); i < w*h; i++ {
		if layer.Regs[i] != 0xff {
			layer.Regs[i] = regs[layer.Regs[i]].RegId
		}
	}

	return detour.DT_SUCCESS
}

func appendVertex(cont *dtTempContour, x, y, z, r int32) bool {
	// Try to merge with existing segments.
	if cont.Nverts > 1 {
		pa := cont.Verts[(cont.Nverts-2)*4:]
		pb := cont.Verts[(cont.Nverts-1)*4:]
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
	if cont.Nverts+1 > cont.Cverts {
		return false
	}

	v := cont.Verts[cont.Nverts*4:]
	v[0] = uint8(x)
	v[1] = uint8(y)
	v[2] = uint8(z)
	v[3] = uint8(r)
	cont.Nverts++

	return true
}

func getNeighbourReg(layer *DtTileCacheLayer, ax, ay, dir int32) uint8 {
	w := int32(layer.Header.Width)
	ia := ax + ay*w

	con := layer.Cons[ia] & 0xf
	portal := layer.Cons[ia] >> 4
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

	return layer.Regs[ib]
}

func walkContour(layer *DtTileCacheLayer, x, y int32, cont *dtTempContour) bool {
	w := int32(layer.Header.Width)
	h := int32(layer.Header.Height)

	cont.Nverts = 0

	startX := x
	startY := y
	startDir := int32(-1)

	for i := int32(0); i < 4; i++ {
		dir := (i + 3) & 3
		rn := getNeighbourReg(layer, x, y, dir)
		if rn != layer.Regs[x+y*w] {
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

		if rn != layer.Regs[x+y*w] {
			// Solid edge.
			px := x
			pz := y
			switch dir {
			case 0:
				pz++
			case 1:
				px++
				pz++
			case 2:
				px++
			}

			// Try to merge with previous vertex.
			if !appendVertex(cont, px, int32(layer.Heights[x+y*w]), pz, int32(rn)) {
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
	pa := cont.Verts[(cont.Nverts-1)*4:]
	pb := cont.Verts[0:]
	if pa[0] == pb[0] && pa[2] == pb[2] {
		cont.Nverts--
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
	cont.Npoly = 0

	for i := int32(0); i < cont.Nverts; i++ {
		j := (i + 1) % cont.Nverts
		// Check for start of a wall segment.
		ra := cont.Verts[j*4+3]
		rb := cont.Verts[i*4+3]
		if ra != rb {
			cont.Poly[cont.Npoly] = uint16(i)
			cont.Npoly++
		}
	}
	if cont.Npoly < 2 {
		// If there is no transitions at all,
		// create some initial points for the simplification process.
		// Find lower-left and upper-right vertices of the contour.
		llx := cont.Verts[0]
		llz := cont.Verts[2]
		lli := int32(0)
		urx := cont.Verts[0]
		urz := cont.Verts[2]
		uri := int32(0)
		for i := int32(1); i < cont.Nverts; i++ {
			x := cont.Verts[i*4+0]
			z := cont.Verts[i*4+2]
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
		cont.Npoly = 0
		cont.Poly[cont.Npoly] = uint16(lli)
		cont.Npoly++
		cont.Poly[cont.Npoly] = uint16(uri)
		cont.Npoly++
	}

	// Add points until all raw points are within
	// error tolerance to the simplified shape.
	for i := int32(0); i < cont.Npoly; {
		ii := (i + 1) % cont.Npoly

		ai := int32(cont.Poly[i])
		ax := int32(cont.Verts[ai*4+0])
		az := int32(cont.Verts[ai*4+2])

		bi := int32(cont.Poly[ii])
		bx := int32(cont.Verts[bi*4+0])
		bz := int32(cont.Verts[bi*4+2])

		// Find maximum deviation from the segment.
		var maxd float32
		maxi := int32(-1)
		var ci, cinc, endi int32

		// Traverse the segment in lexilogical order so that the
		// max deviation is calculated similarly when traversing
		// opposite segments.
		if bx > ax || (bx == ax && bz > az) {
			cinc = 1
			ci = (ai + cinc) % cont.Nverts
			endi = bi
		} else {
			cinc = cont.Nverts - 1
			ci = (bi + cinc) % cont.Nverts
			endi = ai
		}

		// Tessellate only outer edges or edges between areas.
		for ci != endi {
			d := distancePtSeg(int32(cont.Verts[ci*4+0]), int32(cont.Verts[ci*4+2]), ax, az, bx, bz)
			if d > maxd {
				maxd = d
				maxi = ci
			}
			ci = (ci + cinc) % cont.Nverts
		}

		// If the max deviation is larger than accepted error,
		// add new point, else continue to next segment.
		if maxi != -1 && maxd > (maxError*maxError) {
			cont.Npoly++
			for j := cont.Npoly - 1; j > i; j-- {
				cont.Poly[j] = cont.Poly[j-1]
			}
			cont.Poly[i+1] = uint16(maxi)
		} else {
			i++
		}
	}

	// Remap vertices
	var start int32
	for i := int32(1); i < cont.Npoly; i++ {
		if cont.Poly[i] < cont.Poly[start] {
			start = i
		}
	}

	cont.Nverts = 0
	for i := int32(0); i < cont.Npoly; i++ {
		j := (start + i) % cont.Npoly
		src := cont.Verts[cont.Poly[j]*4:]
		dst := cont.Verts[cont.Nverts*4:]
		dst[0] = src[0]
		dst[1] = src[1]
		dst[2] = src[2]
		dst[3] = src[3]
		cont.Nverts++
	}
}

func getCornerHeight(layer *DtTileCacheLayer, x, y, z, walkableClimb int32, shouldRemove *bool) uint8 {
	w := int32(layer.Header.Width)
	h := int32(layer.Header.Height)

	var n int32

	portal := uint8(0xf)
	height := uint8(0)
	preg := uint8(0xff)
	allSameReg := true

	for dz := int32(-1); dz <= 0; dz++ {
		for dx := int32(-1); dx <= 0; dx++ {
			px := x + dx
			pz := z + dz
			if px >= 0 && pz >= 0 && px < w && pz < h {
				idx := px + pz*w
				lh := int32(layer.Heights[idx])
				if detour.DtAbsInt32(lh-y) <= walkableClimb && layer.Areas[idx] != DT_TILECACHE_NULL_AREA {
					height = detour.DtMaxUInt8(height, uint8(lh))
					portal &= (layer.Cons[idx] >> 4)
					if preg != 0xff && preg != layer.Regs[idx] {
						allSameReg = false
					}
					preg = layer.Regs[idx]
					n++
				}
			}
		}
	}

	var portalCount int32
	for dir := uint32(0); dir < 4; dir++ {
		if (portal & (1 << dir)) != 0 {
			portalCount++
		}
	}

	*shouldRemove = false
	if n > 1 && portalCount == 1 && allSameReg {
		*shouldRemove = true
	}

	return height
}

// TODO: move this somewhere else, once the layer meshing is done.
func DtBuildTileCacheContours(layer *DtTileCacheLayer, walkableClimb int32, maxError float32, lcset *DtTileCacheContourSet) detour.DtStatus {
	w := int32(layer.Header.Width)
	h := int32(layer.Header.Height)

	lcset.Nconts = int32(layer.RegCount)
	lcset.Conts = make([]DtTileCacheContour, lcset.Nconts)
	if lcset.Conts == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	// Allocate temp buffer for contour tracing.
	maxTempVerts := (w + h) * 2 * 2 // Twice around the layer.

	tempVerts := make([]uint8, maxTempVerts*4)
	if tempVerts == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	tempPoly := make([]uint16, maxTempVerts)
	if tempPoly == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	var temp dtTempContour
	temp.init(tempVerts, maxTempVerts, tempPoly, maxTempVerts)

	// Find contours.
	for y := int32(0); y < h; y++ {
		for x := int32(0); x < w; x++ {
			idx := x + y*w
			ri := layer.Regs[idx]
			if ri == 0xff {
				continue
			}

			cont := &lcset.Conts[ri]

			if cont.Nverts > 0 {
				continue
			}

			cont.Reg = ri
			cont.Area = layer.Areas[idx]

			if !walkContour(layer, x, y, &temp) {
				// Too complex contour.
				// Note: If you hit here ofte, try increasing 'maxTempVerts'.
				return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
			}

			simplifyContour(&temp, maxError)

			// Store contour.
			cont.Nverts = temp.Nverts
			if cont.Nverts > 0 {
				cont.Verts = make([]uint8, 4*temp.Nverts)
				if cont.Verts == nil {
					return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
				}

				for i, j := int32(0), temp.Nverts-1; i < temp.Nverts; j, i = i, i+1 {
					dst := cont.Verts[j*4:]
					v := temp.Verts[j*4:]
					vn := temp.Verts[i*4:]
					nei := vn[3] // The neighbour reg is stored at segment vertex of a segment.
					shouldRemove := false
					lh := getCornerHeight(layer, int32(v[0]), int32(v[1]), int32(v[2]),
						walkableClimb, &shouldRemove)

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
				}
			}
		}
	}

	return detour.DT_SUCCESS
}

const VERTEX_BUCKET_COUNT2 int32 = (1 << 8)

func computeVertexHash2(x, y, z int32) int32 {
	const h1 uint32 = 0x8da6b343 // Large multiplicative constants;
	const h2 uint32 = 0xd8163841 // here arbitrarily chosen primes
	const h3 uint32 = 0xcb1ab31f
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
	Vert     [2]uint16
	PolyEdge [2]uint16
	Poly     [2]uint16
}

func buildMeshAdjacency(polys []uint16, npolys int32,
	verts []uint16, nverts int32,
	lcset *DtTileCacheContourSet) bool {
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
				edge.Vert[0] = v0
				edge.Vert[1] = v1
				edge.Poly[0] = uint16(i)
				edge.PolyEdge[0] = uint16(j)
				edge.Poly[1] = uint16(i)
				edge.PolyEdge[1] = 0xff
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
					if edge.Vert[1] == v0 && edge.Poly[0] == edge.Poly[1] {
						edge.Poly[1] = uint16(i)
						edge.PolyEdge[1] = uint16(j)
						found = true
						break
					}
				}
				if !found {
					// Matching edge not found, it is an open edge, add it.
					edge := &edges[edgeCount]
					edge.Vert[0] = v1
					edge.Vert[1] = v0
					edge.Poly[0] = uint16(i)
					edge.PolyEdge[0] = uint16(j)
					edge.Poly[1] = uint16(i)
					edge.PolyEdge[1] = 0xff
					// Insert edge
					nextEdge[edgeCount] = firstEdge[v1]
					firstEdge[v1] = uint16(edgeCount)
					edgeCount++
				}
			}
		}
	}

	// Mark portal edges.
	for i := int32(0); i < lcset.Nconts; i++ {
		cont := &lcset.Conts[i]
		if cont.Nverts < 3 {
			continue
		}

		for j, k := int32(0), cont.Nverts-1; j < cont.Nverts; k, j = j, j+1 {
			va := cont.Verts[k*4:]
			vb := cont.Verts[j*4:]
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
					if e.Poly[0] != e.Poly[1] {
						continue
					}
					eva := verts[e.Vert[0]*3:]
					evb := verts[e.Vert[1]*3:]
					if eva[0] == x && evb[0] == x {
						ezmin := eva[2]
						ezmax := evb[2]
						if ezmin > ezmax {
							ezmin, ezmax = ezmax, ezmin
							// detour.DtSwapUInt16(&ezmin, &ezmax)
						}
						if overlapRangeExl(zmin, zmax, ezmin, ezmax) {
							// Reuse the other polyedge to store dir.
							e.PolyEdge[1] = uint16(dir)
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
					if e.Poly[0] != e.Poly[1] {
						continue
					}
					eva := verts[e.Vert[0]*3:]
					evb := verts[e.Vert[1]*3:]
					if eva[2] == z && evb[2] == z {
						exmin := eva[0]
						exmax := evb[0]
						if exmin > exmax {
							exmin, exmax = exmax, exmin
							// detour.DtSwapUInt16(&exmin, &exmax)
						}
						if overlapRangeExl(xmin, xmax, exmin, exmax) {
							// Reuse the other polyedge to store dir.
							e.PolyEdge[1] = uint16(dir)
						}
					}
				}
			}
		}
	}

	// Store adjacency
	for i := int32(0); i < edgeCount; i++ {
		e := &edges[i]
		if e.Poly[0] != e.Poly[1] {
			p0 := polys[int32(e.Poly[0])*MAX_VERTS_PER_POLY*2:]
			p1 := polys[int32(e.Poly[1])*MAX_VERTS_PER_POLY*2:]
			p0[MAX_VERTS_PER_POLY+int32(e.PolyEdge[0])] = e.Poly[1]
			p1[MAX_VERTS_PER_POLY+int32(e.PolyEdge[1])] = e.Poly[0]
		} else if e.PolyEdge[1] != 0xff {
			p0 := polys[int32(e.Poly[0])*MAX_VERTS_PER_POLY*2:]
			p0[MAX_VERTS_PER_POLY+int32(e.PolyEdge[0])] = 0x8000 | uint16(e.PolyEdge[1])
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

func getPolyMergeValue(pa, pb, verts []uint16, ea, eb *int32) int32 {
	na := countPolyVerts(pa)
	nb := countPolyVerts(pb)

	// If the merged polygon would be too big, do not merge.
	if na+nb-2 > MAX_VERTS_PER_POLY {
		return -1
	}

	// Check if the polygons share an edge.
	*ea = -1
	*eb = -1

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
				*ea = i
				*eb = j
				break
			}
		}
	}

	// No common edge, cannot merge.
	if *ea == -1 || *eb == -1 {
		return -1
	}

	// Check to see if the merged polygon would be convex.
	var va, vb, vc uint16

	va = pa[(*ea+na-1)%na]
	vb = pa[*ea]
	vc = pb[(*eb+2)%nb]
	if !uleft(verts[va*3:], verts[vb*3:], verts[vc*3:]) {
		return -1
	}

	va = pb[(*eb+nb-1)%nb]
	vb = pb[*eb]
	vc = pa[(*ea+2)%na]
	if !uleft(verts[va*3:], verts[vb*3:], verts[vc*3:]) {
		return -1
	}

	va = pa[*ea]
	vb = pa[(*ea+1)%na]

	dx := int32(verts[va*3+0]) - int32(verts[vb*3+0])
	dy := int32(verts[va*3+2]) - int32(verts[vb*3+2])

	return dx*dx + dy*dy
}

func mergePolys(pa, pb []uint16, ea, eb int32) {
	var tmp [MAX_VERTS_PER_POLY * 2]uint16

	na := countPolyVerts(pa)
	nb := countPolyVerts(pb)

	// Merge polygons.
	detour.Memset(uintptr(unsafe.Pointer(&tmp[0])), 0xff, int(MAX_VERTS_PER_POLY*2))

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
	for i := int32(0); i < mesh.Npolys; i++ {
		p := mesh.Polys[i*MAX_VERTS_PER_POLY*2:]
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

	for i := int32(0); i < mesh.Npolys; i++ {
		p := mesh.Polys[i*MAX_VERTS_PER_POLY*2:]
		nv := countPolyVerts(p)

		// Collect edges which touches the removed vertex.
		for j, k := int32(0), nv-1; j < nv; k, j = j, j+1 {
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
	for i := int32(0); i < mesh.Npolys; i++ {
		p := mesh.Polys[i*MAX_VERTS_PER_POLY*2:]
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

	for i := int32(0); i < mesh.Npolys; i++ {
		p := mesh.Polys[i*MAX_VERTS_PER_POLY*2:]
		nv := countPolyVerts(p)
		hasRem := false
		for j := int32(0); j < nv; j++ {
			if p[j] == rem {
				hasRem = true
			}
		}
		if hasRem {
			// Collect edges which does not touch the removed vertex.
			for j, k := int32(0), nv-1; j < nv; k, j = j, j+1 {
				if p[j] != rem && p[k] != rem {
					if nedges >= MAX_REM_EDGES {
						return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
					}
					e := edges[nedges*3:]
					e[0] = p[k]
					e[1] = p[j]
					e[2] = uint16(mesh.Areas[i])
					nedges++
				}
			}
			// Remove the polygon.
			p2 := mesh.Polys[(mesh.Npolys-1)*MAX_VERTS_PER_POLY*2:]
			copy(p, p2[:MAX_VERTS_PER_POLY])
			detour.Memset(uintptr(unsafe.Pointer(&p[MAX_VERTS_PER_POLY])), 0xff, ShortSize*int(MAX_VERTS_PER_POLY))

			mesh.Areas[i] = mesh.Areas[mesh.Npolys-1]
			mesh.Npolys--
			i--
		}
	}

	// Remove vertex.
	for i := int32(rem); i < mesh.Nverts; i++ {
		mesh.Verts[i*3+0] = mesh.Verts[(i+1)*3+0]
		mesh.Verts[i*3+1] = mesh.Verts[(i+1)*3+1]
		mesh.Verts[i*3+2] = mesh.Verts[(i+1)*3+2]
	}
	mesh.Nverts--

	// Adjust indices to match the removed vertex layout.
	for i := int32(0); i < mesh.Npolys; i++ {
		p := mesh.Polys[i*MAX_VERTS_PER_POLY*2:]
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

	for nedges != 0 {
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
		tverts[i*4+0] = uint8(mesh.Verts[pi*3+0])
		tverts[i*4+1] = uint8(mesh.Verts[pi*3+1])
		tverts[i*4+2] = uint8(mesh.Verts[pi*3+2])
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
	detour.Memset(uintptr(unsafe.Pointer(&polys[0])), 0xff, int(ntris*MAX_VERTS_PER_POLY)*ShortSize)
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
					var ea, eb int32
					v := getPolyMergeValue(pj, pk, mesh.Verts, &ea, &eb)
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
		if mesh.Npolys >= maxTris {
			break
		}
		p := mesh.Polys[mesh.Npolys*MAX_VERTS_PER_POLY*2:]
		detour.Memset(uintptr(unsafe.Pointer(&p[0])), 0xff, ShortSize*int(MAX_VERTS_PER_POLY*2))
		for j := int32(0); j < MAX_VERTS_PER_POLY; j++ {
			p[j] = polys[i*MAX_VERTS_PER_POLY+j]
		}
		mesh.Areas[mesh.Npolys] = pareas[i]
		mesh.Npolys++
		if mesh.Npolys > maxTris {
			return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
		}
	}

	return detour.DT_SUCCESS
}

func DtBuildTileCachePolyMesh(lcset *DtTileCacheContourSet, mesh *DtTileCachePolyMesh) detour.DtStatus {
	var maxVertices, maxTris, maxVertsPerCont int32
	for i := int32(0); i < lcset.Nconts; i++ {
		// Skip null contours.
		if lcset.Conts[i].Nverts < 3 {
			continue
		}
		maxVertices += lcset.Conts[i].Nverts
		maxTris += lcset.Conts[i].Nverts - 2
		maxVertsPerCont = detour.DtMaxInt32(maxVertsPerCont, lcset.Conts[i].Nverts)
	}

	// TODO: warn about too many vertices?

	mesh.Nvp = MAX_VERTS_PER_POLY

	vflags := make([]uint8, maxVertices)
	if vflags == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	mesh.Verts = make([]uint16, maxVertices*3)
	if mesh.Verts == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	mesh.Polys = make([]uint16, maxTris*MAX_VERTS_PER_POLY*2)
	if mesh.Polys == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	mesh.Areas = make([]uint8, maxTris)
	if mesh.Areas == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	mesh.Flags = make([]uint16, maxTris)
	if mesh.Flags == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	mesh.Nverts = 0
	mesh.Npolys = 0

	if len(mesh.Polys) != 0 {
		detour.Memset(uintptr(unsafe.Pointer(&(mesh.Polys[0]))), 0xff, ShortSize*int(maxTris*MAX_VERTS_PER_POLY*2))
	}

	var firstVert [VERTEX_BUCKET_COUNT2]uint16
	detour.Memset(uintptr(unsafe.Pointer(&firstVert[0])), 0xff, ShortSize*int(VERTEX_BUCKET_COUNT2))

	nextVert := make([]uint16, maxVertices)
	if nextVert == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	indices := make([]uint16, maxVertsPerCont)
	if indices == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	tris := make([]uint16, maxVertsPerCont*3)
	if tris == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	polys := make([]uint16, maxVertsPerCont*MAX_VERTS_PER_POLY)
	if polys == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	for i := int32(0); i < lcset.Nconts; i++ {
		cont := &lcset.Conts[i]

		// Skip null contours.
		if cont.Nverts < 3 {
			continue
		}

		// Triangulate contour
		for j := int32(0); j < cont.Nverts; j++ {
			indices[j] = uint16(j)
		}

		ntris := triangulate(cont.Nverts, cont.Verts, indices[:], tris[:])
		if ntris <= 0 {
			// TODO: issue warning!
			ntris = -ntris
		}

		// Add and merge vertices.
		for j := int32(0); j < cont.Nverts; j++ {
			v := cont.Verts[j*4:]
			indices[j] = addVertex(uint16(v[0]), uint16(v[1]), uint16(v[2]),
				mesh.Verts, firstVert[:], nextVert[:], &mesh.Nverts)
			if v[3]&0x80 != 0 {
				// This vertex should be removed.
				vflags[indices[j]] = 1
			}
		}

		// Build initial polygons.
		var npolys int32
		detour.Memset(uintptr(unsafe.Pointer(&(polys[0]))), 0xff, ShortSize*int(maxVertsPerCont*MAX_VERTS_PER_POLY))
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
				// Find best polygons to merge.
				var bestMergeVal, bestPa, bestPb, bestEa, bestEb int32

				for j := int32(0); j < npolys-1; j++ {
					pj := polys[j*MAX_VERTS_PER_POLY:]
					for k := j + 1; k < npolys; k++ {
						pk := polys[k*MAX_VERTS_PER_POLY:]
						var ea, eb int32
						v := getPolyMergeValue(pj, pk, mesh.Verts, &ea, &eb)
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
			p := mesh.Polys[mesh.Npolys*MAX_VERTS_PER_POLY*2:]
			q := polys[j*MAX_VERTS_PER_POLY:]
			for k := int32(0); k < MAX_VERTS_PER_POLY; k++ {
				p[k] = q[k]
			}

			mesh.Areas[mesh.Npolys] = cont.Area
			mesh.Npolys++
			if mesh.Npolys > maxTris {
				return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
			}
		}
	}

	// Remove edge vertices.
	for i := int32(0); i < mesh.Nverts; i++ {
		if vflags[i] != 0 {
			if !canRemoveVertex(mesh, uint16(i)) {
				continue
			}
			status := removeVertex(mesh, uint16(i), maxTris)
			if detour.DtStatusFailed(status) {
				return status
			}
			// Remove vertex
			// Note: mesh.Nverts is already decremented inside removeVertex()!
			for j := i; j < mesh.Nverts; j++ {
				vflags[j] = vflags[j+1]
			}
			i--
		}
	}

	// Calculate adjacency.
	if !buildMeshAdjacency(mesh.Polys, mesh.Npolys, mesh.Verts, mesh.Nverts, lcset) {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	return detour.DT_SUCCESS
}

func DtMarkCylinderArea(layer *DtTileCacheLayer, orig []float32, cs, ch float32,
	pos []float32, radius, height float32, areaId uint8) detour.DtStatus {
	var bmin, bmax [3]float32
	bmin[0] = pos[0] - radius
	bmin[1] = pos[1]
	bmin[2] = pos[2] - radius
	bmax[0] = pos[0] + radius
	bmax[1] = pos[1] + height
	bmax[2] = pos[2] + radius
	r2 := detour.DtSqrFloat32(radius/cs + 0.5)

	w := int32(layer.Header.Width)
	h := int32(layer.Header.Height)
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
			y := int32(layer.Heights[x+z*w])
			if y < miny || y > maxy {
				continue
			}
			layer.Areas[x+z*w] = areaId
		}
	}

	return detour.DT_SUCCESS
}

func DtMarkBoxArea1(layer *DtTileCacheLayer, orig []float32, cs, ch float32,
	bmin, bmax []float32, areaId uint8) detour.DtStatus {
	w := int32(layer.Header.Width)
	h := int32(layer.Header.Height)
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
			y := int32(layer.Heights[x+z*w])
			if y < miny || y > maxy {
				continue
			}
			layer.Areas[x+z*w] = areaId
		}
	}

	return detour.DT_SUCCESS
}

func DtMarkBoxArea2(layer *DtTileCacheLayer, orig []float32, cs, ch float32,
	center, halfExtents, rotAux []float32, areaId uint8) detour.DtStatus {
	w := int32(layer.Header.Width)
	h := int32(layer.Header.Height)
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
			y := int32(layer.Heights[x+z*w])
			if y < miny || y > maxy {
				continue
			}
			layer.Areas[x+z*w] = areaId
		}
	}

	return detour.DT_SUCCESS
}

func DtBuildTileCacheLayer(comp DtTileCacheCompressor,
	header *DtTileCacheLayerHeader,
	heights, areas, cons []uint8,
	outData *[]uint8, outDataSize *int32) detour.DtStatus {

	headerSize := int32(detour.DtAlign4(int(DtTileCacheLayerHeaderSize)))
	gridSize := int32(header.Width) * int32(header.Height)
	maxDataSize := headerSize + comp.MaxCompressedSize(gridSize*3)
	data := make([]byte, maxDataSize)
	if data == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	// Store header
	*(*DtTileCacheLayerHeader)(unsafe.Pointer(&data[0])) = *header

	// Concatenate grid data for compression.
	bufferSize := gridSize * 3
	buffer := make([]uint8, bufferSize)
	if buffer == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	copy(buffer[:gridSize], heights)
	copy(buffer[gridSize:gridSize*2], areas)
	copy(buffer[gridSize*2:gridSize*3], cons)

	// Compress
	compressed := data[headerSize:]
	maxCompressedSize := maxDataSize - headerSize
	var compressedSize int32
	status := comp.Compress(buffer, bufferSize, compressed, maxCompressedSize, &compressedSize)
	if detour.DtStatusFailed(status) {
		return status
	}

	*outData = data
	*outDataSize = headerSize + compressedSize

	buffer = nil

	return detour.DT_SUCCESS
}

func DtFreeTileCacheLayer(layer *DtTileCacheLayer) {

}

func DtDecompressTileCacheLayer(comp DtTileCacheCompressor,
	compressed []uint8, compressedSize int32,
	layerOut **DtTileCacheLayer) detour.DtStatus {

	detour.DtAssert(comp != nil)

	if layerOut == nil {
		return detour.DT_FAILURE | detour.DT_INVALID_PARAM
	}
	if compressed == nil {
		return detour.DT_FAILURE | detour.DT_INVALID_PARAM
	}

	*layerOut = nil

	compressedHeader := (*DtTileCacheLayerHeader)(unsafe.Pointer(&compressed[0]))
	if compressedHeader.Magic != DT_TILECACHE_MAGIC {
		return detour.DT_FAILURE | detour.DT_WRONG_MAGIC
	}
	if compressedHeader.Version != DT_TILECACHE_VERSION {
		return detour.DT_FAILURE | detour.DT_WRONG_VERSION
	}

	layerSize := int32(detour.DtAlign4(int(DtTileCacheLayerSize)))
	headerSize := int32(detour.DtAlign4(int(DtTileCacheLayerHeaderSize)))
	gridSize := int32(compressedHeader.Width) * int32(compressedHeader.Height)
	bufferSize := layerSize + headerSize + gridSize*4

	buffer := make([]byte, bufferSize)
	if buffer == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	layer := (*DtTileCacheLayer)(unsafe.Pointer(&buffer[0]))
	header := (*DtTileCacheLayerHeader)(unsafe.Pointer(uintptr(unsafe.Pointer(&buffer[0])) + uintptr(layerSize)))

	var grids []byte
	gridsSize := bufferSize - (layerSize + headerSize)
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&grids)))
	sliceHeader.Cap = int(gridsSize)
	sliceHeader.Len = int(gridsSize)
	sliceHeader.Data = uintptr(unsafe.Pointer(&buffer[0])) + uintptr(layerSize+headerSize)

	// Copy header
	*header = *compressedHeader
	// Decompress grid.
	var size int32
	status := comp.Decompress(compressed[headerSize:], compressedSize-headerSize, grids, gridsSize, &size)
	detour.DtIgnoreUnused(size)

	if detour.DtStatusFailed(status) {
		buffer = nil
		return status
	}

	layer.Header = header
	layer.Heights = grids
	layer.Areas = grids[gridSize:]
	layer.Cons = grids[gridSize*2:]
	layer.Regs = grids[gridSize*3:]

	*layerOut = layer

	return detour.DT_SUCCESS
}

func DtTileCacheHeaderSwapEndian(data []uint8, dataSize int32) bool {
	// dtIgnoreUnused(dataSize)
	header := (*DtTileCacheLayerHeader)(unsafe.Pointer(&data[0]))

	swappedMagic := DT_TILECACHE_MAGIC
	swappedVersion := DT_TILECACHE_VERSION
	detour.DtSwapEndianInt32(&swappedMagic)
	detour.DtSwapEndianInt32(&swappedVersion)

	if (header.Magic != DT_TILECACHE_MAGIC || header.Version != DT_TILECACHE_VERSION) &&
		(header.Magic != swappedMagic || header.Version != swappedVersion) {
		return false
	}

	detour.DtSwapEndianInt32(&header.Magic)
	detour.DtSwapEndianInt32(&header.Version)
	detour.DtSwapEndianInt32(&header.Tx)
	detour.DtSwapEndianInt32(&header.Ty)
	detour.DtSwapEndianInt32(&header.Tlayer)
	detour.DtSwapEndianFloat32(&header.Bmin[0])
	detour.DtSwapEndianFloat32(&header.Bmin[1])
	detour.DtSwapEndianFloat32(&header.Bmin[2])
	detour.DtSwapEndianFloat32(&header.Bmax[0])
	detour.DtSwapEndianFloat32(&header.Bmax[1])
	detour.DtSwapEndianFloat32(&header.Bmax[2])
	detour.DtSwapEndianUInt16(&header.Hmin)
	detour.DtSwapEndianUInt16(&header.Hmax)

	// width, height, minx, maxx, miny, maxy are unsigned char, no need to swap.

	return true
}
