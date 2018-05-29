package dtcache

import (
	detour "github.com/fananchong/recastnavigation-go/Detour"
)

type DtObstacleRef uint32

type DtCompressedTileRef uint32

/// Flags for AddTile
const (
	DT_COMPRESSEDTILE_FREE_DATA = 0x01 ///< Navmesh owns the tile memory and should free it.
)

type DtCompressedTile struct {
	Salt           uint32 ///< Counter describing modifications to the tile.
	Header         *DtTileCacheLayerHeader
	Compressed     []byte
	CompressedSize int32
	Data           []byte
	DataSize       int32
	Flags          uint32
	Next           *DtCompressedTile
}

type ObstacleState uint8

const (
	DT_OBSTACLE_EMPTY      ObstacleState = 0
	DT_OBSTACLE_PROCESSING ObstacleState = 1
	DT_OBSTACLE_PROCESSED  ObstacleState = 2
	DT_OBSTACLE_REMOVING   ObstacleState = 3
)

type ObstacleType uint8

const (
	DT_OBSTACLE_CYLINDER     ObstacleType = 0
	DT_OBSTACLE_BOX          ObstacleType = 1 // AABB
	DT_OBSTACLE_ORIENTED_BOX ObstacleType = 2 // OBB
)

type DtObstacleCylinder struct {
	Pos    [3]float32
	Radius float32
	Height float32
}

type DtObstacleBox struct {
	Bmin [3]float32
	Bmax [3]float32
}

type DtObstacleOrientedBox struct {
	Center      [3]float32
	HalfExtents [3]float32
	RotAux      [2]float32 //{ cos(0.5f*angle)*sin(-0.5f*angle); cos(0.5f*angle)*cos(0.5f*angle) - 0.5 }
}

const DT_MAX_TOUCHED_TILES int32 = 8

type DtTileCacheObstacle struct {
	Cylinder    DtObstacleCylinder
	Box         DtObstacleBox
	OrientedBox DtObstacleOrientedBox

	Touched  [DT_MAX_TOUCHED_TILES]DtCompressedTileRef
	Pending  [DT_MAX_TOUCHED_TILES]DtCompressedTileRef
	Salt     uint16
	Type     ObstacleType
	State    ObstacleState
	Ntouched uint8
	Npending uint8
	Next     *DtTileCacheObstacle
}

type DtTileCacheParams struct {
	Orig                   [3]float32
	Cs                     float32
	Ch                     float32
	Width                  int32
	Height                 int32
	WalkableHeight         float32
	WalkableRadius         float32
	WalkableClimb          float32
	MaxSimplificationError float32
	MaxTiles               int32
	MaxObstacles           int32
}

type DtTileCacheMeshProcess interface {
	Process(params *detour.DtNavMeshCreateParams, polyAreas []uint8, polyFlags []uint16)
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

func (this *DtTileCache) GetCompressor() DtTileCacheCompressor   { return this.m_tcomp }
func (this *DtTileCache) GetParams() *DtTileCacheParams          { return &this.m_params }
func (this *DtTileCache) GetTileCount() int                      { return int(this.m_params.MaxTiles) }
func (this *DtTileCache) GetTile(i int) *DtCompressedTile        { return &this.m_tiles[i] }
func (this *DtTileCache) GetObstacleCount() int                  { return int(this.m_params.MaxObstacles) }
func (this *DtTileCache) GetObstacle(i int) *DtTileCacheObstacle { return &this.m_obstacles[i] }

/// Encodes a tile id.
func (this *DtTileCache) EncodeTileId(salt, it uint32) DtCompressedTileRef {
	return (DtCompressedTileRef(salt) << this.m_tileBits) | DtCompressedTileRef(it)
}

/// Decodes a tile salt.
func (this *DtTileCache) DecodeTileIdSalt(ref DtCompressedTileRef) uint32 {
	saltMask := (DtCompressedTileRef(1) << this.m_saltBits) - 1
	return uint32((ref >> this.m_tileBits) & saltMask)
}

/// Decodes a tile id.
func (this *DtTileCache) DecodeTileIdTile(ref DtCompressedTileRef) uint32 {
	tileMask := (DtCompressedTileRef(1) << this.m_tileBits) - 1
	return uint32(ref & tileMask)
}

/// Encodes an obstacle id.
func (this *DtTileCache) EncodeObstacleId(salt, it uint32) DtObstacleRef {
	return (DtObstacleRef(salt) << 16) | DtObstacleRef(it)
}

/// Decodes an obstacle salt.
func (this *DtTileCache) DecodeObstacleIdSalt(ref DtObstacleRef) uint32 {
	saltMask := (DtObstacleRef(1) << 16) - 1
	return uint32((ref >> 16) & saltMask)
}

/// Decodes an obstacle id.
func (this *DtTileCache) DecodeObstacleIdObstacle(ref DtObstacleRef) uint32 {
	tileMask := (DtObstacleRef(1) << 16) - 1
	return uint32(ref & tileMask)
}

// #endif
