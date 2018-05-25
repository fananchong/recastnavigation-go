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
	"github.com/fananchong/recastnavigation-go/Detour"
)

const (
	DT_TILECACHE_MAGIC         int32  = 'D'<<24 | 'T'<<16 | 'L'<<8 | 'R' ///< 'DTLR';
	DT_TILECACHE_VERSION       int32  = 1
	DT_TILECACHE_NULL_AREA     uint8  = 0
	DT_TILECACHE_WALKABLE_AREA uint8  = 63
	DT_TILECACHE_NULL_IDX      uint16 = 0xffff
)

type DtTileCacheLayerHeader struct {
	magic   int32 ///< Data magic
	version int32 ///< Data version
	tx      int32
	ty      int32
	tlayer  int32
	bmin    [3]float32
	bmax    [3]float32
	hmin    uint16
	hmax    uint16 ///< Height min/max range
	width   uint8
	height  uint8 ///< Dimension of the layer.
	minx    uint8
	maxx    uint8
	miny    uint8
	maxy    uint8 ///< Usable sub-region.
}

type DtTileCacheLayer struct {
	header   *DtTileCacheLayerHeader
	regCount uint8 ///< Region count.
	heights  []uint8
	areas    []uint8
	cons     []uint8
	regs     []uint8
}

type DtTileCacheContour struct {
	nverts int32
	verts  []uint8
	reg    uint8
	area   uint8
}

type DtTileCacheContourSet struct {
	nconts int32
	conts  []DtTileCacheContour
}

type DtTileCachePolyMesh struct {
	nvp    int32
	nverts int32    ///< Number of vertices.
	npolys int32    ///< Number of polygons.
	verts  []uint16 ///< Vertices of the mesh, 3 elements per vertex.
	polys  []uint16 ///< Polygons of the mesh, nvp*2 elements per polygon.
	flags  []uint16 ///< Per polygon flags.
	areas  []uint8  ///< Area ID of polygons.
}

type DtTileCacheCompressor interface {
	DtFreeTileCacheCompressor()

	MaxCompressedSize(bufferSize int32) int32
	Compress(buffer []uint8, bufferSize int32, compressed []uint8, maxCompressedSize int32) (compressedSize int32, status detour.DtStatus)
	Decompress(compressed []uint8, compressedSize int32, buffer []uint8, maxBufferSize int32) (bufferSize int32, status detour.DtStatus)
}
