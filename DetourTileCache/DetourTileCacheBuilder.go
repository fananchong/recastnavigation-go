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

	"github.com/fananchong/recastnavigation-go/Detour"
)

const (
	DT_TILECACHE_MAGIC         int32  = 'D'<<24 | 'T'<<16 | 'L'<<8 | 'R' ///< 'DTLR';
	DT_TILECACHE_VERSION       int32  = 1
	DT_TILECACHE_NULL_AREA     uint8  = 0
	DT_TILECACHE_WALKABLE_AREA uint8  = 63
	DT_TILECACHE_NULL_IDX      uint16 = 0xffff
)

const (
	ShortSize                  = int(unsafe.Sizeof(uint16(1)))
	DtTileCacheLayerSize       = unsafe.Sizeof(DtTileCacheLayer{})
	DtTileCacheLayerHeaderSize = unsafe.Sizeof(DtTileCacheLayerHeader{})
)

type DtTileCacheLayerHeader struct {
	Magic   int32 ///< Data magic
	Version int32 ///< Data version
	Tx      int32
	Ty      int32
	Tlayer  int32
	Bmin    [3]float32
	Bmax    [3]float32
	Hmin    uint16 ///< Height min/max range
	Hmax    uint16 ///< Height min/max range
	Width   uint8  ///< Dimension of the layer.
	Height  uint8  ///< Dimension of the layer.
	Minx    uint8  ///< Usable sub-region.
	Maxx    uint8  ///< Usable sub-region.
	Miny    uint8  ///< Usable sub-region.
	Maxy    uint8  ///< Usable sub-region.
}

type DtTileCacheLayer struct {
	Header   *DtTileCacheLayerHeader
	RegCount uint8 ///< Region count.
	Heights  []uint8
	Areas    []uint8
	Cons     []uint8
	Regs     []uint8
}

type DtTileCacheContour struct {
	Nverts int32
	Verts  []uint8
	Reg    uint8
	Area   uint8
}

type DtTileCacheContourSet struct {
	Nconts int32
	Conts  []DtTileCacheContour
}

type DtTileCachePolyMesh struct {
	Nvp    int32
	Nverts int32    ///< Number of vertices.
	Npolys int32    ///< Number of polygons.
	Verts  []uint16 ///< Vertices of the mesh, 3 elements per vertex.
	Polys  []uint16 ///< Polygons of the mesh, nvp*2 elements per polygon.
	Flags  []uint16 ///< Per polygon flags.
	Areas  []uint8  ///< Area ID of polygons.
}

type DtTileCacheCompressor interface {
	MaxCompressedSize(bufferSize int32) int32
	Compress(buffer []byte, bufferSize int32, compressed []byte, maxCompressedSize int32, compressedSize *int32) detour.DtStatus
	Decompress(compressed []byte, compressedSize int32, buffer []byte, maxBufferSize int32, bufferSize *int32) detour.DtStatus
}
