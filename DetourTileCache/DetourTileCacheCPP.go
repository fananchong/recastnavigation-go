package dtcache

import (
	"math"
	"unsafe"

	"github.com/fananchong/recastnavigation-go/Detour"
)

const (
	dtCompressedTileSize    = unsafe.Sizeof(DtCompressedTile{})
	dtTileCacheObstacleSize = unsafe.Sizeof(DtTileCacheObstacle{})
)

func DtAllocTileCache() *DtTileCache {
	c := &DtTileCache{}
	c.construct()
	return c
}

func DtFreeTileCache(tc *DtTileCache) {
	if tc != nil {
		tc.destructor()
	}
}

func contains(a []DtCompressedTileRef, n int32, v DtCompressedTileRef) bool {
	for i := int32(0); i < n; i++ {
		if a[i] == v {
			return true
		}
	}
	return false
}

func computeTileHash(x, y, mask int32) int32 {
	h1 := uint32(0x8da6b343) // Large multiplicative constants;
	h2 := uint32(0xd8163841) // here arbitrarily chosen primes
	n := h1*uint32(x) + h2*uint32(y)
	return int32(n & uint32(mask))
}

type NavMeshTileBuildContext struct {
	layer *DtTileCacheLayer
	lcset *DtTileCacheContourSet
	lmesh *DtTileCachePolyMesh
}

func (this *DtTileCache) construct() {

}

func (this *DtTileCache) destructor() {
	for i := 0; i < int(this.m_params.MaxTiles); i++ {
		if (this.m_tiles[i].Flags & DT_COMPRESSEDTILE_FREE_DATA) != 0 {
			this.m_tiles[i].Data = nil
		}
	}
	this.m_obstacles = nil
	this.m_posLookup = nil
	this.m_tiles = nil
	this.m_nreqs = 0
	this.m_nupdate = 0
}

func (this *DtTileCache) GetTileByRef(ref DtCompressedTileRef) *DtCompressedTile {
	if ref == 0 {
		return nil
	}
	tileIndex := this.DecodeTileIdTile(ref)
	tileSalt := this.DecodeTileIdSalt(ref)
	if int32(tileIndex) >= this.m_params.MaxTiles {
		return nil
	}
	tile := &this.m_tiles[tileIndex]
	if tile.Salt != tileSalt {
		return nil
	}
	return tile
}

func (this *DtTileCache) Init(params *DtTileCacheParams,
	tcomp DtTileCacheCompressor,
	tmproc DtTileCacheMeshProcess) detour.DtStatus {
	this.m_tcomp = tcomp
	this.m_tmproc = tmproc
	this.m_nreqs = 0
	this.m_params = *params

	// Alloc space for obstacles.
	this.m_obstacles = make([]DtTileCacheObstacle, this.m_params.MaxObstacles)
	if this.m_obstacles == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}
	this.m_nextFreeObstacle = nil
	for i := int(this.m_params.MaxObstacles - 1); i >= 0; i-- {
		this.m_obstacles[i].Salt = 1
		this.m_obstacles[i].Next = this.m_nextFreeObstacle
		this.m_nextFreeObstacle = &this.m_obstacles[i]
	}

	// Init tiles
	this.m_tileLutSize = int32(detour.DtNextPow2(uint32(this.m_params.MaxTiles / 4)))
	if this.m_tileLutSize == 0 {
		this.m_tileLutSize = 1
	}
	this.m_tileLutMask = this.m_tileLutSize - 1

	this.m_tiles = make([]DtCompressedTile, this.m_params.MaxTiles)
	if this.m_tiles == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}
	this.m_posLookup = make([]*DtCompressedTile, this.m_tileLutSize)
	if this.m_posLookup == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}
	this.m_nextFreeTile = nil
	for i := int(this.m_params.MaxTiles - 1); i >= 0; i-- {
		this.m_tiles[i].Salt = 1
		this.m_tiles[i].Next = this.m_nextFreeTile
		this.m_nextFreeTile = &this.m_tiles[i]
	}

	// Init ID generator values.
	this.m_tileBits = detour.DtIlog2(detour.DtNextPow2(uint32(this.m_params.MaxTiles)))
	// Only allow 31 salt bits, since the salt mask is calculated using 32bit uint and it will overflow.
	this.m_saltBits = detour.DtMinUInt32(uint32(31), 32-this.m_tileBits)
	if this.m_saltBits < 10 {
		return detour.DT_FAILURE | detour.DT_INVALID_PARAM
	}

	return detour.DT_SUCCESS
}

func (this *DtTileCache) GetTilesAt(tx, ty int32, tiles []DtCompressedTileRef, maxTiles int32) int32 {
	var n int32
	// Find tile based on hash.
	h := computeTileHash(tx, ty, this.m_tileLutMask)
	tile := this.m_posLookup[h]
	for tile != nil {
		if tile.Header != nil &&
			tile.Header.Tx == tx &&
			tile.Header.Ty == ty {
			if n < maxTiles {
				tiles[n] = this.GetTileRef(tile)
				n++
			}
		}
		tile = tile.Next
	}

	return n
}

func (this *DtTileCache) GetTileAt(tx, ty, tlayer int32) *DtCompressedTile {
	// Find tile based on hash.
	h := computeTileHash(tx, ty, this.m_tileLutMask)
	tile := this.m_posLookup[h]
	for tile != nil {
		if tile.Header != nil &&
			tile.Header.Tx == tx &&
			tile.Header.Ty == ty &&
			tile.Header.Tlayer == tlayer {
			return tile
		}
		tile = tile.Next
	}
	return nil
}

func (this *DtTileCache) GetTileRef(tile *DtCompressedTile) DtCompressedTileRef {
	if tile == nil {
		return 0
	}

	it := detour.SliceSizeFromPointer(unsafe.Pointer(tile), unsafe.Pointer(&this.m_tiles[0]), dtCompressedTileSize)
	return DtCompressedTileRef(this.EncodeTileId(tile.Salt, it))
}

func (this *DtTileCache) GetObstacleRef(ob *DtTileCacheObstacle) DtObstacleRef {
	if ob == nil {
		return 0
	}
	idx := detour.SliceSizeFromPointer(unsafe.Pointer(ob), unsafe.Pointer(&this.m_obstacles[0]), dtTileCacheObstacleSize)
	return this.EncodeObstacleId(uint32(ob.Salt), idx)
}

func (this *DtTileCache) GetObstacleByRef(ref DtObstacleRef) *DtTileCacheObstacle {
	if ref == 0 {
		return nil
	}
	idx := this.DecodeObstacleIdObstacle(ref)
	if int32(idx) >= this.m_params.MaxObstacles {
		return nil
	}
	ob := &this.m_obstacles[idx]
	salt := this.DecodeObstacleIdSalt(ref)
	if uint32(ob.Salt) != salt {
		return nil
	}
	return ob
}

func (this *DtTileCache) AddTile(data []byte, dataSize int32, flags uint8, result *DtCompressedTileRef) detour.DtStatus {
	// Make sure the data is in right format.
	header := (*DtTileCacheLayerHeader)(unsafe.Pointer(&data[0]))
	if header.Magic != DT_TILECACHE_MAGIC {
		return detour.DT_FAILURE | detour.DT_WRONG_MAGIC
	}
	if header.Version != DT_TILECACHE_VERSION {
		return detour.DT_FAILURE | detour.DT_WRONG_VERSION
	}

	// Make sure the location is free.
	if this.GetTileAt(header.Tx, header.Ty, header.Tlayer) != nil {
		return detour.DT_FAILURE
	}

	// Allocate a tile.
	var tile *DtCompressedTile
	if this.m_nextFreeTile != nil {
		tile = this.m_nextFreeTile
		this.m_nextFreeTile = tile.Next
		tile.Next = nil
	}

	// Make sure we could allocate a tile.
	if tile == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	// Insert tile into the position lut.
	h := computeTileHash(header.Tx, header.Ty, this.m_tileLutMask)
	tile.Next = this.m_posLookup[h]
	this.m_posLookup[h] = tile

	// Init tile.
	headerSize := int32(detour.DtAlign4(int(DtTileCacheLayerHeaderSize)))
	tile.Header = (*DtTileCacheLayerHeader)(unsafe.Pointer(&data[0]))
	tile.Data = data
	tile.DataSize = dataSize
	tile.Compressed = tile.Data[headerSize:]
	tile.CompressedSize = tile.DataSize - headerSize
	tile.Flags = uint32(flags)

	if result != nil {
		*result = this.GetTileRef(tile)
	}

	return detour.DT_SUCCESS
}

func (this *DtTileCache) RemoveTile(ref DtCompressedTileRef, data *[]byte, dataSize *int32) detour.DtStatus {
	if ref == 0 {
		return detour.DT_FAILURE | detour.DT_INVALID_PARAM
	}
	tileIndex := this.DecodeTileIdTile(ref)
	tileSalt := this.DecodeTileIdSalt(ref)
	if int32(tileIndex) >= this.m_params.MaxTiles {
		return detour.DT_FAILURE | detour.DT_INVALID_PARAM
	}
	tile := &this.m_tiles[tileIndex]
	if tile.Salt != tileSalt {
		return detour.DT_FAILURE | detour.DT_INVALID_PARAM
	}

	// Remove tile from hash lookup.
	h := computeTileHash(tile.Header.Tx, tile.Header.Ty, this.m_tileLutMask)
	var prev *DtCompressedTile
	cur := this.m_posLookup[h]
	for cur != nil {
		if cur == tile {
			if prev != nil {
				prev.Next = cur.Next
			} else {
				this.m_posLookup[h] = cur.Next
			}
			break
		}
		prev = cur
		cur = cur.Next
	}

	// Reset tile.
	if (tile.Flags & DT_COMPRESSEDTILE_FREE_DATA) != 0 {
		// Owns data
		tile.Data = nil
		tile.DataSize = 0
		if data != nil {
			*data = nil
		}
		if dataSize != nil {
			*dataSize = 0
		}
	} else {
		if data != nil {
			*data = tile.Data
		}
		if dataSize != nil {
			*dataSize = tile.DataSize
		}
	}

	tile.Header = nil
	tile.Data = nil
	tile.DataSize = 0
	tile.Compressed = nil
	tile.CompressedSize = 0
	tile.Flags = 0

	// Update salt, salt should never be zero.
	tile.Salt = (tile.Salt + 1) & ((1 << this.m_saltBits) - 1)
	if tile.Salt == 0 {
		tile.Salt++
	}

	// Add to free list.
	tile.Next = this.m_nextFreeTile
	this.m_nextFreeTile = tile

	return detour.DT_SUCCESS
}

func (this *DtTileCache) AddObstacle(pos []float32, radius, height float32, result *DtObstacleRef) detour.DtStatus {
	if this.m_nreqs >= MAX_REQUESTS {
		return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
	}

	var ob *DtTileCacheObstacle
	if this.m_nextFreeObstacle != nil {
		ob = this.m_nextFreeObstacle
		this.m_nextFreeObstacle = ob.Next
		ob.Next = nil
	}
	if ob == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	salt := ob.Salt
	*ob = DtTileCacheObstacle{}

	ob.Salt = salt
	ob.State = DT_OBSTACLE_PROCESSING
	ob.Type = DT_OBSTACLE_CYLINDER
	detour.DtVcopy(ob.Cylinder.Pos[:], pos)
	ob.Cylinder.Radius = radius
	ob.Cylinder.Height = height

	req := &this.m_reqs[this.m_nreqs]
	this.m_nreqs++
	*req = ObstacleRequest{}
	req.action = REQUEST_ADD
	req.ref = this.GetObstacleRef(ob)

	if result != nil {
		*result = req.ref
	}

	return detour.DT_SUCCESS
}

func (this *DtTileCache) AddBoxObstacle(bmin, bmax []float32, result *DtObstacleRef) detour.DtStatus {
	if this.m_nreqs >= MAX_REQUESTS {
		return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
	}

	var ob *DtTileCacheObstacle
	if this.m_nextFreeObstacle != nil {
		ob = this.m_nextFreeObstacle
		this.m_nextFreeObstacle = ob.Next
		ob.Next = nil
	}
	if ob == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	salt := ob.Salt
	*ob = DtTileCacheObstacle{}
	ob.Salt = salt
	ob.State = DT_OBSTACLE_PROCESSING
	ob.Type = DT_OBSTACLE_BOX
	detour.DtVcopy(ob.Box.Bmin[:], bmin)
	detour.DtVcopy(ob.Box.Bmax[:], bmax)

	req := &this.m_reqs[this.m_nreqs]
	this.m_nreqs++
	*req = ObstacleRequest{}
	req.action = REQUEST_ADD
	req.ref = this.GetObstacleRef(ob)

	if result != nil {
		*result = req.ref
	}

	return detour.DT_SUCCESS
}

func (this *DtTileCache) AddBoxObstacle2(center, halfExtents []float32, yRadians float32, result *DtObstacleRef) detour.DtStatus {
	if this.m_nreqs >= MAX_REQUESTS {
		return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
	}

	var ob *DtTileCacheObstacle
	if this.m_nextFreeObstacle != nil {
		ob = this.m_nextFreeObstacle
		this.m_nextFreeObstacle = ob.Next
		ob.Next = nil
	}
	if ob == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	salt := ob.Salt
	*ob = DtTileCacheObstacle{}
	ob.Salt = salt
	ob.State = DT_OBSTACLE_PROCESSING
	ob.Type = DT_OBSTACLE_ORIENTED_BOX
	detour.DtVcopy(ob.OrientedBox.Center[:], center)
	detour.DtVcopy(ob.OrientedBox.HalfExtents[:], halfExtents)

	coshalf := float32(math.Cos(0.5 * float64(yRadians)))
	sinhalf := float32(math.Sin(-0.5 * float64(yRadians)))
	ob.OrientedBox.RotAux[0] = coshalf * sinhalf
	ob.OrientedBox.RotAux[1] = coshalf*coshalf - 0.5

	req := &this.m_reqs[this.m_nreqs]
	this.m_nreqs++
	*req = ObstacleRequest{}
	req.action = REQUEST_ADD
	req.ref = this.GetObstacleRef(ob)

	if result != nil {
		*result = req.ref
	}

	return detour.DT_SUCCESS
}

func (this *DtTileCache) RemoveObstacle(ref DtObstacleRef) detour.DtStatus {
	if ref == 0 {
		return detour.DT_SUCCESS
	}
	if this.m_nreqs >= MAX_REQUESTS {
		return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
	}

	req := &this.m_reqs[this.m_nreqs]
	this.m_nreqs++
	*req = ObstacleRequest{}
	req.action = REQUEST_REMOVE
	req.ref = ref

	return detour.DT_SUCCESS
}

func (this *DtTileCache) QueryTiles(bmin, bmax []float32,
	results []DtCompressedTileRef, resultCount *int32, maxResults int32) detour.DtStatus {
	const MAX_TILES int32 = 32
	var tiles [MAX_TILES]DtCompressedTileRef

	var n int32

	tw := float32(this.m_params.Width) * this.m_params.Cs
	th := float32(this.m_params.Height) * this.m_params.Cs
	tx0 := int32(detour.DtMathFloorf((bmin[0] - this.m_params.Orig[0]) / tw))
	tx1 := int32(detour.DtMathFloorf((bmax[0] - this.m_params.Orig[0]) / tw))
	ty0 := int32(detour.DtMathFloorf((bmin[2] - this.m_params.Orig[2]) / th))
	ty1 := int32(detour.DtMathFloorf((bmax[2] - this.m_params.Orig[2]) / th))

	for ty := ty0; ty <= ty1; ty++ {
		for tx := tx0; tx <= tx1; tx++ {
			ntiles := this.GetTilesAt(tx, ty, tiles[:], MAX_TILES)

			for i := int32(0); i < ntiles; i++ {
				tile := &this.m_tiles[this.DecodeTileIdTile(tiles[i])]
				var tbmin, tbmax [3]float32
				this.CalcTightTileBounds(tile.Header, tbmin[:], tbmax[:])

				if detour.DtOverlapBounds(bmin, bmax, tbmin[:], tbmax[:]) {
					if n < maxResults {
						results[n] = tiles[i]
						n++
					}
				}
			}
		}
	}

	*resultCount = n

	return detour.DT_SUCCESS
}

func (this *DtTileCache) Update(dt float32, navmesh *detour.DtNavMesh,
	upToDate *bool) detour.DtStatus {
	if this.m_nupdate == 0 {
		// Process requests.
		for i := int32(0); i < this.m_nreqs; i++ {
			req := &this.m_reqs[i]

			idx := this.DecodeObstacleIdObstacle(req.ref)
			if int32(idx) >= this.m_params.MaxObstacles {
				continue
			}
			ob := &this.m_obstacles[idx]
			salt := this.DecodeObstacleIdSalt(req.ref)
			if uint32(ob.Salt) != salt {
				continue
			}

			if req.action == REQUEST_ADD {
				// Find touched tiles.
				var bmin, bmax [3]float32
				this.GetObstacleBounds(ob, bmin[:], bmax[:])

				var ntouched int32
				this.QueryTiles(bmin[:], bmax[:], ob.Touched[:], &ntouched, DT_MAX_TOUCHED_TILES)
				ob.Ntouched = uint8(ntouched)
				// Add tiles to update list.
				ob.Npending = 0
				for j := int32(0); j < int32(ob.Ntouched); j++ {
					if this.m_nupdate < MAX_UPDATE {
						if !contains(this.m_update[:], this.m_nupdate, ob.Touched[j]) {
							this.m_update[this.m_nupdate] = ob.Touched[j]
							this.m_nupdate++
						}
						ob.Pending[ob.Npending] = ob.Touched[j]
						ob.Npending++
					}
				}
			} else if req.action == REQUEST_REMOVE {
				// Prepare to remove obstacle.
				ob.State = DT_OBSTACLE_REMOVING
				// Add tiles to update list.
				ob.Npending = 0
				for j := int32(0); j < int32(ob.Ntouched); j++ {
					if this.m_nupdate < MAX_UPDATE {
						if !contains(this.m_update[:], this.m_nupdate, ob.Touched[j]) {
							this.m_update[this.m_nupdate] = ob.Touched[j]
							this.m_nupdate++
						}
						ob.Pending[ob.Npending] = ob.Touched[j]
						ob.Npending++
					}
				}
			}
		}

		this.m_nreqs = 0
	}

	status := detour.DT_SUCCESS
	// Process updates
	if this.m_nupdate > 0 {
		// Build mesh
		ref := this.m_update[0]
		status = this.BuildNavMeshTile(ref, navmesh)
		this.m_nupdate--
		if this.m_nupdate > 0 {
			for i := int32(0); i < this.m_nupdate; i++ {
				this.m_update[i] = this.m_update[i+1]
			}
		}

		// Update obstacle states.
		for i := int32(0); i < this.m_params.MaxObstacles; i++ {
			ob := &this.m_obstacles[i]
			if ob.State == DT_OBSTACLE_PROCESSING || ob.State == DT_OBSTACLE_REMOVING {
				// Remove handled tile from pending list.
				for j := int32(0); j < int32(ob.Npending); j++ {
					if ob.Pending[j] == ref {
						ob.Pending[j] = ob.Pending[int32(ob.Npending-1)]
						ob.Npending--
						break
					}
				}

				// If all pending tiles processed, change state.
				if ob.Npending == 0 {
					if ob.State == DT_OBSTACLE_PROCESSING {
						ob.State = DT_OBSTACLE_PROCESSED
					} else if ob.State == DT_OBSTACLE_REMOVING {
						ob.State = DT_OBSTACLE_EMPTY
						// Update salt, salt should never be zero.
						ob.Salt = (ob.Salt + 1) & ((1 << 16) - 1)
						if ob.Salt == 0 {
							ob.Salt++
						}
						// Return obstacle to free list.
						ob.Next = this.m_nextFreeObstacle
						this.m_nextFreeObstacle = ob
					}
				}
			}
		}
	}

	if upToDate != nil {
		*upToDate = ((this.m_nupdate == 0) && (this.m_nreqs == 0))
	}

	return status
}

func (this *DtTileCache) BuildNavMeshTilesAt(tx, ty int32, navmesh *detour.DtNavMesh) detour.DtStatus {
	const MAX_TILES = int32(32)
	var tiles [MAX_TILES]DtCompressedTileRef
	ntiles := this.GetTilesAt(tx, ty, tiles[:], MAX_TILES)

	for i := int32(0); i < ntiles; i++ {
		status := this.BuildNavMeshTile(tiles[i], navmesh)
		if detour.DtStatusFailed(status) {
			return status
		}
	}

	return detour.DT_SUCCESS
}

func (this *DtTileCache) BuildNavMeshTile(ref DtCompressedTileRef, navmesh *detour.DtNavMesh) detour.DtStatus {
	detour.DtAssert(this.m_tcomp != nil)

	idx := this.DecodeTileIdTile(ref)
	if idx > uint32(this.m_params.MaxTiles) {
		return detour.DT_FAILURE | detour.DT_INVALID_PARAM
	}
	tile := &this.m_tiles[idx]
	salt := this.DecodeTileIdSalt(ref)
	if tile.Salt != salt {
		return detour.DT_FAILURE | detour.DT_INVALID_PARAM
	}

	var bc NavMeshTileBuildContext
	walkableClimbVx := int32(this.m_params.WalkableClimb / this.m_params.Ch)
	var status detour.DtStatus

	// Decompress tile layer data.
	status = DtDecompressTileCacheLayer(this.m_tcomp, tile.Data, tile.DataSize, &bc.layer)
	if detour.DtStatusFailed(status) {
		return status
	}

	// Rasterize obstacles.
	for i := int32(0); i < this.m_params.MaxObstacles; i++ {
		ob := &this.m_obstacles[i]
		if ob.State == DT_OBSTACLE_EMPTY || ob.State == DT_OBSTACLE_REMOVING {
			continue
		}
		if contains(ob.Touched[:], int32(ob.Ntouched), ref) {
			if ob.Type == DT_OBSTACLE_CYLINDER {
				DtMarkCylinderArea(bc.layer, tile.Header.Bmin[:], this.m_params.Cs, this.m_params.Ch,
					ob.Cylinder.Pos[:], ob.Cylinder.Radius, ob.Cylinder.Height, 0)
			} else if ob.Type == DT_OBSTACLE_BOX {
				DtMarkBoxArea1(bc.layer, tile.Header.Bmin[:], this.m_params.Cs, this.m_params.Ch,
					ob.Box.Bmin[:], ob.Box.Bmax[:], 0)
			} else if ob.Type == DT_OBSTACLE_ORIENTED_BOX {
				DtMarkBoxArea2(bc.layer, tile.Header.Bmin[:], this.m_params.Cs, this.m_params.Ch,
					ob.OrientedBox.Center[:], ob.OrientedBox.HalfExtents[:], ob.OrientedBox.RotAux[:], 0)
			}
		}
	}

	// Build navmesh
	status = DtBuildTileCacheRegions(bc.layer, walkableClimbVx)
	if detour.DtStatusFailed(status) {
		return status
	}

	bc.lcset = DtAllocTileCacheContourSet()
	if bc.lcset == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}
	status = DtBuildTileCacheContours(bc.layer, walkableClimbVx,
		this.m_params.MaxSimplificationError, bc.lcset)
	if detour.DtStatusFailed(status) {
		return status
	}

	bc.lmesh = DtAllocTileCachePolyMesh()
	if bc.lmesh == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}
	status = DtBuildTileCachePolyMesh(bc.lcset, bc.lmesh)
	if detour.DtStatusFailed(status) {
		return status
	}

	// Early out if the mesh tile is empty.
	if bc.lmesh.Npolys == 0 {
		// Remove existing tile.
		navmesh.RemoveTile(navmesh.GetTileRefAt(tile.Header.Tx, tile.Header.Ty, tile.Header.Tlayer), nil, nil)
		return detour.DT_SUCCESS
	}

	var params detour.DtNavMeshCreateParams

	params.Verts = bc.lmesh.Verts
	params.VertCount = bc.lmesh.Nverts
	params.Polys = bc.lmesh.Polys
	params.PolyAreas = bc.lmesh.Areas
	params.PolyFlags = bc.lmesh.Flags
	params.PolyCount = bc.lmesh.Npolys
	params.Nvp = detour.DT_VERTS_PER_POLYGON
	params.WalkableHeight = this.m_params.WalkableHeight
	params.WalkableRadius = this.m_params.WalkableRadius
	params.WalkableClimb = this.m_params.WalkableClimb
	params.TileX = tile.Header.Tx
	params.TileY = tile.Header.Ty
	params.TileLayer = tile.Header.Tlayer
	params.Cs = this.m_params.Cs
	params.Ch = this.m_params.Ch
	params.BuildBvTree = false
	detour.DtVcopy(params.Bmin[:], tile.Header.Bmin[:])
	detour.DtVcopy(params.Bmax[:], tile.Header.Bmax[:])

	if this.m_tmproc != nil {
		this.m_tmproc.Process(&params, bc.lmesh.Areas[:], bc.lmesh.Flags[:])
	}

	var navData []byte
	var navDataSize int
	if !detour.DtCreateNavMeshData(&params, &navData, &navDataSize) {
		return detour.DT_FAILURE
	}

	// Remove existing tile.
	navmesh.RemoveTile(navmesh.GetTileRefAt(tile.Header.Tx, tile.Header.Ty, tile.Header.Tlayer), nil, nil)

	// Add new tile, or leave the location empty.
	if navData != nil {
		// Let the navmesh own the data.
		status = navmesh.AddTile(navData, navDataSize, detour.DT_TILE_FREE_DATA, 0, nil)
		if detour.DtStatusFailed(status) {
			return status
		}
	}

	return detour.DT_SUCCESS
}

func (this *DtTileCache) CalcTightTileBounds(header *DtTileCacheLayerHeader, bmin, bmax []float32) {
	cs := this.m_params.Cs
	bmin[0] = header.Bmin[0] + float32(header.Minx)*cs
	bmin[1] = header.Bmin[1]
	bmin[2] = header.Bmin[2] + float32(header.Miny)*cs
	bmax[0] = header.Bmin[0] + float32(header.Maxx+1)*cs
	bmax[1] = header.Bmax[1]
	bmax[2] = header.Bmin[2] + float32(header.Maxy+1)*cs
}

func (this *DtTileCache) GetObstacleBounds(ob *DtTileCacheObstacle, bmin, bmax []float32) {
	if ob.Type == DT_OBSTACLE_CYLINDER {
		cl := &ob.Cylinder

		bmin[0] = cl.Pos[0] - cl.Radius
		bmin[1] = cl.Pos[1]
		bmin[2] = cl.Pos[2] - cl.Radius
		bmax[0] = cl.Pos[0] + cl.Radius
		bmax[1] = cl.Pos[1] + cl.Height
		bmax[2] = cl.Pos[2] + cl.Radius
	} else if ob.Type == DT_OBSTACLE_BOX {
		detour.DtVcopy(bmin, ob.Box.Bmin[:])
		detour.DtVcopy(bmax, ob.Box.Bmax[:])
	} else if ob.Type == DT_OBSTACLE_ORIENTED_BOX {
		orientedBox := &ob.OrientedBox

		maxr := 1.41 * detour.DtMaxFloat32(orientedBox.HalfExtents[0], orientedBox.HalfExtents[2])
		bmin[0] = orientedBox.Center[0] - maxr
		bmax[0] = orientedBox.Center[0] + maxr
		bmin[1] = orientedBox.Center[1] - orientedBox.HalfExtents[1]
		bmax[1] = orientedBox.Center[1] + orientedBox.HalfExtents[1]
		bmin[2] = orientedBox.Center[2] - maxr
		bmax[2] = orientedBox.Center[2] + maxr
	}
}
