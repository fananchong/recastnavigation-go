package dtcache

import (
	detour "github.com/fananchong/recastnavigation-go/Detour"
)

type DtObstacleRef uint32

type DtCompressedTileRef uint32

/// Flags for addTile
const (
	DT_COMPRESSEDTILE_FREE_DATA = 0x01 ///< Navmesh owns the tile memory and should free it.
)

type DtCompressedTile struct {
	salt           uint32 ///< Counter describing modifications to the tile.
	header         *DtTileCacheLayerHeader
	compressed     []uint8
	compressedSize int32
	data           []uint8
	dataSize       int32
	flags          uint32
	next           *DtCompressedTile
}

const (
	DT_OBSTACLE_EMPTY      int32 = 0
	DT_OBSTACLE_PROCESSING int32 = 1
	DT_OBSTACLE_PROCESSED  int32 = 2
	DT_OBSTACLE_REMOVING   int32 = 3

	DT_OBSTACLE_CYLINDER     int32 = 0
	DT_OBSTACLE_BOX          int32 = 1 // AABB
	DT_OBSTACLE_ORIENTED_BOX int32 = 2 // OBB
)

type DtObstacleCylinder struct {
	pos    [3]float32
	radius float32
	height float32
}

type DtObstacleBox struct {
	bmin [3]float32
	bmax [3]float32
}

type DtObstacleOrientedBox struct {
	center      [3]float32
	halfExtents [3]float32
	rotAux      [2]float32 //{ cos(0.5f*angle)*sin(-0.5f*angle); cos(0.5f*angle)*cos(0.5f*angle) - 0.5 }
}

const DT_MAX_TOUCHED_TILES int32 = 8

type DtTileCacheObstacle struct {
	cylinder    DtObstacleCylinder
	box         DtObstacleBox
	orientedBox DtObstacleOrientedBox

	touched  [DT_MAX_TOUCHED_TILES]DtCompressedTileRef
	pending  [DT_MAX_TOUCHED_TILES]DtCompressedTileRef
	salt     uint16
	_type    uint8
	state    uint8
	ntouched uint8
	npending uint8
	next     *DtTileCacheObstacle
}

type DtTileCacheParams struct {
	orig                   [3]float32
	cs                     float32
	ch                     float32
	width                  int32
	height                 int32
	walkableHeight         float32
	walkableRadius         float32
	walkableClimb          float32
	maxSimplificationError float32
	maxTiles               int32
	maxObstacles           int32
}

type DtTileCacheMeshProcess interface {
	DtFreeTileCacheMeshProcess()

	process(params *detour.DtNavMeshCreateParams, polyAreas []uint8, polyFlags []uint16)
}

const MAX_REQUESTS int32 = 64
const MAX_UPDATE int32 = 64

const REQUEST_ADD int32 = 0
const REQUEST_REMOVE int32 = 1

type ObstacleRequest struct {
	action int32
	ref    DtObstacleRef
}

type DtTileCache struct {
	m_tileLutSize int32 ///< Tile hash lookup size (must be pot).
	m_tileLutMask int32 ///< Tile hash lookup mask.

	m_posLookup    []*DtCompressedTile ///< Tile hash lookup.
	m_nextFreeTile *DtCompressedTile   ///< Freelist of tiles.
	m_tiles        []DtCompressedTile  ///< List of tiles.

	m_saltBits uint32 ///< Number of salt bits in the tile ID.
	m_tileBits uint32 ///< Number of tile bits in the tile ID.

	m_params DtTileCacheParams

	m_tcomp  DtTileCacheCompressor
	m_tmproc DtTileCacheMeshProcess

	m_obstacles        []DtTileCacheObstacle
	m_nextFreeObstacle *DtTileCacheObstacle

	m_reqs  [MAX_REQUESTS]ObstacleRequest
	m_nreqs int32

	m_update  [MAX_UPDATE]DtCompressedTileRef
	m_nupdate int32
}

/// Encodes a tile id.
func (this *DtTileCache) encodeTileId(salt, it uint32) DtCompressedTileRef {
	return (DtCompressedTileRef(salt) << this.m_tileBits) | DtCompressedTileRef(it)
}

/// Decodes a tile salt.
func (this *DtTileCache) decodeTileIdSalt(ref DtCompressedTileRef) uint32 {
	saltMask := (DtCompressedTileRef(1) << this.m_saltBits) - 1
	return uint32((ref >> this.m_tileBits) & saltMask)
}

/// Decodes a tile id.
func (this *DtTileCache) decodeTileIdTile(ref DtCompressedTileRef) uint32 {
	tileMask := (DtCompressedTileRef(1) << this.m_tileBits) - 1
	return uint32(ref & tileMask)
}

/// Encodes an obstacle id.
func (this *DtTileCache) encodeObstacleId(salt, it uint32) DtObstacleRef {
	return (DtObstacleRef(salt) << 16) | DtObstacleRef(it)
}

/// Decodes an obstacle salt.
func (this *DtTileCache) decodeObstacleIdSalt(ref DtObstacleRef) uint32 {
	saltMask := (DtObstacleRef(1) << 16) - 1
	return uint32((ref >> 16) & saltMask)
}

/// Decodes an obstacle id.
func (this *DtTileCache) decodeObstacleIdObstacle(ref DtObstacleRef) uint32 {
	tileMask := (DtObstacleRef(1) << 16) - 1
	return uint32(ref & tileMask)
}

// #endif
