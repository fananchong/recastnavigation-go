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

func dtAllocTileCache() *DtTileCache {
	c := &DtTileCache{}
	c.construct()
	return c
}

func dtFreeTileCache(tc *DtTileCache) {
	tc.destructor()
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

}

func (this *DtTileCache) getTileByRef(ref DtCompressedTileRef) *DtCompressedTile {
	if ref == 0 {
		return nil
	}
	tileIndex := this.decodeTileIdTile(ref)
	tileSalt := this.decodeTileIdSalt(ref)
	if int32(tileIndex) >= this.m_params.maxTiles {
		return nil
	}
	tile := &this.m_tiles[tileIndex]
	if tile.salt != tileSalt {
		return nil
	}
	return tile
}

func (this *DtTileCache) init(params *DtTileCacheParams, tcomp DtTileCacheCompressor, tmproc DtTileCacheMeshProcess) detour.DtStatus {
	this.m_tcomp = tcomp
	this.m_tmproc = tmproc
	this.m_nreqs = 0
	this.m_params = *params

	// Alloc space for obstacles.
	this.m_obstacles = make([]DtTileCacheObstacle, this.m_params.maxObstacles)

	this.m_nextFreeObstacle = nil
	for i := this.m_params.maxObstacles - 1; i >= 0; i-- {
		this.m_obstacles[i].salt = 1
		this.m_obstacles[i].next = this.m_nextFreeObstacle
		this.m_nextFreeObstacle = &this.m_obstacles[i]
	}

	// Init tiles
	this.m_tileLutSize = int32(detour.DtNextPow2(uint32(this.m_params.maxTiles / 4)))
	if this.m_tileLutSize == 0 {
		this.m_tileLutSize = 1
	}
	this.m_tileLutMask = this.m_tileLutSize - 1

	this.m_tiles = make([]DtCompressedTile, this.m_params.maxTiles)
	this.m_posLookup = make([]*DtCompressedTile, this.m_tileLutSize)
	this.m_nextFreeTile = nil
	for i := this.m_params.maxTiles - 1; i >= 0; i-- {
		this.m_tiles[i].salt = 1
		this.m_tiles[i].next = this.m_nextFreeTile
		this.m_nextFreeTile = &this.m_tiles[i]
	}

	// Init ID generator values.
	this.m_tileBits = detour.DtIlog2(detour.DtNextPow2(uint32(this.m_params.maxTiles)))
	// Only allow 31 salt bits, since the salt mask is calculated using 32bit uint and it will overflow.
	this.m_saltBits = detour.DtMinUInt32(uint32(31), 32-this.m_tileBits)
	if this.m_saltBits < 10 {
		return detour.DT_FAILURE | detour.DT_INVALID_PARAM
	}

	return detour.DT_SUCCESS
}

func (this *DtTileCache) getTilesAt(tx, ty int32, tiles []DtCompressedTileRef, maxTiles int32) int32 {
	var n int32
	// Find tile based on hash.
	h := computeTileHash(tx, ty, this.m_tileLutMask)
	tile := this.m_posLookup[h]
	for tile != nil {
		if tile.header != nil && tile.header.tx == tx && tile.header.ty == ty {
			if n < maxTiles {
				tiles[n] = this.getTileRef(tile)
				n++
			}
		}
		tile = tile.next
	}

	return n
}

func (this *DtTileCache) getTileAt(tx, ty, tlayer int32) *DtCompressedTile {
	// Find tile based on hash.
	h := computeTileHash(tx, ty, this.m_tileLutMask)
	tile := this.m_posLookup[h]
	for tile != nil {
		if tile.header != nil && tile.header.tx == tx && tile.header.ty == ty && tile.header.tlayer == tlayer {
			return tile
		}
		tile = tile.next
	}
	return nil
}

func (this *DtTileCache) getTileRef(tile *DtCompressedTile) DtCompressedTileRef {
	if tile == nil {
		return 0
	}

	it := detour.SliceSizeFromPointer(unsafe.Pointer(tile), unsafe.Pointer(&this.m_tiles[0]), dtCompressedTileSize)
	return DtCompressedTileRef(this.encodeTileId(tile.salt, it))
}

func (this *DtTileCache) getObstacleRef(ob *DtTileCacheObstacle) DtObstacleRef {
	if ob == nil {
		return 0
	}
	idx := detour.SliceSizeFromPointer(unsafe.Pointer(ob), unsafe.Pointer(&this.m_obstacles[0]), dtTileCacheObstacleSize)
	return this.encodeObstacleId(uint32(ob.salt), idx)
}

func (this *DtTileCache) getObstacleByRef(ref DtObstacleRef) *DtTileCacheObstacle {
	if ref == 0 {
		return nil
	}
	idx := this.decodeObstacleIdObstacle(ref)
	if int32(idx) >= this.m_params.maxObstacles {
		return nil
	}
	ob := &this.m_obstacles[idx]
	salt := this.decodeObstacleIdSalt(ref)
	if uint32(ob.salt) != salt {
		return nil
	}
	return ob
}

func (this *DtTileCache) addTile(data []uint8, dataSize int32, flags uint8, result *DtCompressedTileRef) detour.DtStatus {
	// Make sure the data is in right format.
	header := (*DtTileCacheLayerHeader)(unsafe.Pointer(&data[0]))
	if header.magic != DT_TILECACHE_MAGIC {
		return detour.DT_FAILURE | detour.DT_WRONG_MAGIC
	}
	if header.version != DT_TILECACHE_VERSION {
		return detour.DT_FAILURE | detour.DT_WRONG_VERSION
	}

	// Make sure the location is free.
	if this.getTileAt(header.tx, header.ty, header.tlayer) != nil {
		return detour.DT_FAILURE
	}

	// Allocate a tile.
	var tile *DtCompressedTile
	if this.m_nextFreeTile != nil {
		tile = this.m_nextFreeTile
		this.m_nextFreeTile = tile.next
		tile.next = nil
	}

	// Make sure we could allocate a tile.
	if tile == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	// Insert tile into the position lut.
	h := computeTileHash(header.tx, header.ty, this.m_tileLutMask)
	tile.next = this.m_posLookup[h]
	this.m_posLookup[h] = tile

	// Init tile.
	headerSize := int32(detour.DtAlign4(int(DtTileCacheLayerHeaderSize)))
	tile.header = (*DtTileCacheLayerHeader)(unsafe.Pointer(&data[0]))
	tile.data = data
	tile.dataSize = dataSize
	tile.compressed = tile.data[headerSize:]
	tile.compressedSize = tile.dataSize - headerSize
	tile.flags = uint32(flags)

	if result != nil {
		*result = this.getTileRef(tile)
	}

	return detour.DT_SUCCESS
}

func (this *DtTileCache) removeTile(ref DtCompressedTileRef, data *[]uint8, dataSize *int32) detour.DtStatus {
	if ref == 0 {
		return detour.DT_FAILURE | detour.DT_INVALID_PARAM
	}
	tileIndex := this.decodeTileIdTile(ref)
	tileSalt := this.decodeTileIdSalt(ref)
	if int32(tileIndex) >= this.m_params.maxTiles {
		return detour.DT_FAILURE | detour.DT_INVALID_PARAM
	}
	tile := &this.m_tiles[tileIndex]
	if tile.salt != tileSalt {
		return detour.DT_FAILURE | detour.DT_INVALID_PARAM
	}

	// Remove tile from hash lookup.
	h := computeTileHash(tile.header.tx, tile.header.ty, this.m_tileLutMask)
	var prev *DtCompressedTile
	cur := this.m_posLookup[h]
	for cur != nil {
		if cur == tile {
			if prev != nil {
				prev.next = cur.next
			} else {
				this.m_posLookup[h] = cur.next
			}
			break
		}
		prev = cur
		cur = cur.next
	}

	// Reset tile.
	if tile.flags&DT_COMPRESSEDTILE_FREE_DATA > 0 {
		// Owns data
		tile.data = nil
		tile.dataSize = 0
		if data != nil {
			*data = nil
		}
		if dataSize != nil {
			*dataSize = 0
		}
	} else {
		if data != nil {
			*data = tile.data
		}
		if dataSize != nil {
			*dataSize = tile.dataSize
		}
	}

	tile.header = nil
	tile.data = nil
	tile.dataSize = 0
	tile.compressed = nil
	tile.compressedSize = 0
	tile.flags = 0

	// Update salt, salt should never be zero.
	tile.salt = (tile.salt + 1) & ((1 << this.m_saltBits) - 1)
	if tile.salt == 0 {
		tile.salt++
	}

	// Add to free list.
	tile.next = this.m_nextFreeTile
	this.m_nextFreeTile = tile

	return detour.DT_SUCCESS
}

func (this *DtTileCache) addObstacle(pos []float32, radius, height float32, result *DtObstacleRef) detour.DtStatus {
	if this.m_nreqs >= MAX_REQUESTS {
		return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
	}

	var ob *DtTileCacheObstacle
	if this.m_nextFreeObstacle != nil {
		ob = this.m_nextFreeObstacle
		this.m_nextFreeObstacle = ob.next
		ob.next = nil
	}
	if ob == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	salt := ob.salt
	*ob = DtTileCacheObstacle{}

	ob.salt = salt
	ob.state = uint8(DT_OBSTACLE_PROCESSING)
	ob._type = uint8(DT_OBSTACLE_CYLINDER)
	detour.DtVcopy(ob.cylinder.pos[:], pos)
	ob.cylinder.radius = radius
	ob.cylinder.height = height

	req := &this.m_reqs[this.m_nreqs]
	this.m_nreqs++
	*req = ObstacleRequest{}
	req.action = REQUEST_ADD
	req.ref = this.getObstacleRef(ob)

	if result != nil {
		*result = req.ref
	}

	return detour.DT_SUCCESS
}

func (this *DtTileCache) addBoxObstacle(bmin, bmax []float32, result *DtObstacleRef) detour.DtStatus {
	if this.m_nreqs >= MAX_REQUESTS {
		return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
	}

	var ob *DtTileCacheObstacle
	if this.m_nextFreeObstacle != nil {
		ob = this.m_nextFreeObstacle
		this.m_nextFreeObstacle = ob.next
		ob.next = nil
	}
	if ob == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	salt := ob.salt
	*ob = DtTileCacheObstacle{}
	ob.salt = salt
	ob.state = uint8(DT_OBSTACLE_PROCESSING)
	ob._type = uint8(DT_OBSTACLE_BOX)
	detour.DtVcopy(ob.box.bmin[:], bmin)
	detour.DtVcopy(ob.box.bmax[:], bmax)

	req := &this.m_reqs[this.m_nreqs]
	this.m_nreqs++
	*req = ObstacleRequest{}
	req.action = REQUEST_ADD
	req.ref = this.getObstacleRef(ob)

	if result != nil {
		*result = req.ref
	}

	return detour.DT_SUCCESS
}

func (this *DtTileCache) addBoxObstacle2(center, halfExtents []float32, yRadians float32, result *DtObstacleRef) detour.DtStatus {
	if this.m_nreqs >= MAX_REQUESTS {
		return detour.DT_FAILURE | detour.DT_BUFFER_TOO_SMALL
	}

	var ob *DtTileCacheObstacle
	if this.m_nextFreeObstacle != nil {
		ob = this.m_nextFreeObstacle
		this.m_nextFreeObstacle = ob.next
		ob.next = nil
	}
	if ob == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}

	salt := ob.salt
	*ob = DtTileCacheObstacle{}
	ob.salt = salt
	ob.state = uint8(DT_OBSTACLE_PROCESSING)
	ob._type = uint8(DT_OBSTACLE_ORIENTED_BOX)
	detour.DtVcopy(ob.orientedBox.center[:], center)
	detour.DtVcopy(ob.orientedBox.halfExtents[:], halfExtents)

	coshalf := float32(math.Cos(0.5 * float64(yRadians)))
	sinhalf := float32(math.Sin(-0.5 * float64(yRadians)))
	ob.orientedBox.rotAux[0] = coshalf * sinhalf
	ob.orientedBox.rotAux[1] = coshalf*coshalf - 0.5

	req := &this.m_reqs[this.m_nreqs]
	this.m_nreqs++
	*req = ObstacleRequest{}
	req.action = REQUEST_ADD
	req.ref = this.getObstacleRef(ob)

	if result != nil {
		*result = req.ref
	}

	return detour.DT_SUCCESS
}

func (this *DtTileCache) removeObstacle(ref DtObstacleRef) detour.DtStatus {
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

const MAX_TILES int32 = 32

func (this *DtTileCache) queryTiles(bmin, bmax []float32, results []DtCompressedTileRef, resultCount *int32, maxResults int32) detour.DtStatus {
	var tiles [MAX_TILES]DtCompressedTileRef

	var n int32

	tw := float32(this.m_params.width) * this.m_params.cs
	th := float32(this.m_params.height) * this.m_params.cs
	tx0 := int32(detour.DtMathFloorf((bmin[0] - this.m_params.orig[0]) / tw))
	tx1 := int32(detour.DtMathFloorf((bmax[0] - this.m_params.orig[0]) / tw))
	ty0 := int32(detour.DtMathFloorf((bmin[2] - this.m_params.orig[2]) / th))
	ty1 := int32(detour.DtMathFloorf((bmax[2] - this.m_params.orig[2]) / th))

	for ty := ty0; ty <= ty1; ty++ {
		for tx := tx0; tx <= tx1; tx++ {
			ntiles := this.getTilesAt(tx, ty, tiles[:], MAX_TILES)

			for i := int32(0); i < ntiles; i++ {
				tile := &this.m_tiles[this.decodeTileIdTile(tiles[i])]
				var tbmin, tbmax [3]float32
				this.calcTightTileBounds(tile.header, tbmin[:], tbmax[:])

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

func (this *DtTileCache) update(dt float32, navmesh *detour.DtNavMesh, upToDate *bool) detour.DtStatus {
	if this.m_nupdate == 0 {
		// Process requests.
		for i := int32(0); i < this.m_nreqs; i++ {
			req := &this.m_reqs[i]

			idx := this.decodeObstacleIdObstacle(req.ref)
			if int32(idx) >= this.m_params.maxObstacles {
				continue
			}
			ob := &this.m_obstacles[idx]
			salt := this.decodeObstacleIdSalt(req.ref)
			if uint32(ob.salt) != salt {
				continue
			}

			if req.action == REQUEST_ADD {
				// Find touched tiles.
				var bmin, bmax [3]float32
				this.getObstacleBounds(ob, bmin[:], bmax[:])

				var ntouched int32
				this.queryTiles(bmin[:], bmax[:], ob.touched[:], &ntouched, DT_MAX_TOUCHED_TILES)
				ob.ntouched = uint8(ntouched)
				// Add tiles to update list.
				ob.npending = 0
				for j := int32(0); j < int32(ob.ntouched); j++ {
					if this.m_nupdate < MAX_UPDATE {
						if !contains(this.m_update[:], this.m_nupdate, ob.touched[j]) {
							this.m_update[this.m_nupdate] = ob.touched[j]
							this.m_nupdate++
						}
						ob.pending[ob.npending] = ob.touched[j]
						ob.npending++
					}
				}
			} else if req.action == REQUEST_REMOVE {
				// Prepare to remove obstacle.
				ob.state = uint8(DT_OBSTACLE_REMOVING)
				// Add tiles to update list.
				ob.npending = 0
				for j := int32(0); j < int32(ob.ntouched); j++ {
					if this.m_nupdate < MAX_UPDATE {
						if !contains(this.m_update[:], this.m_nupdate, ob.touched[j]) {
							this.m_update[this.m_nupdate] = ob.touched[j]
							this.m_nupdate++
						}
						ob.pending[ob.npending] = ob.touched[j]
						ob.npending++
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
		status = this.buildNavMeshTile(ref, navmesh)
		this.m_nupdate--
		if this.m_nupdate > 0 {
			for i := int32(0); i < this.m_nupdate; i++ {
				this.m_update[i] = this.m_update[i+1]
			}
		}

		// Update obstacle states.
		for i := int32(0); i < this.m_params.maxObstacles; i++ {
			ob := &this.m_obstacles[i]
			if ob.state == uint8(DT_OBSTACLE_PROCESSING) || ob.state == uint8(DT_OBSTACLE_REMOVING) {
				// Remove handled tile from pending list.
				for j := int32(0); j < int32(ob.npending); j++ {
					if ob.pending[j] == ref {
						ob.pending[j] = ob.pending[int32(ob.npending-1)]
						ob.npending--
						break
					}
				}

				// If all pending tiles processed, change state.
				if ob.npending == 0 {
					if ob.state == uint8(DT_OBSTACLE_PROCESSING) {
						ob.state = uint8(DT_OBSTACLE_PROCESSED)
					} else if ob.state == uint8(DT_OBSTACLE_REMOVING) {
						ob.state = uint8(DT_OBSTACLE_EMPTY)
						// Update salt, salt should never be zero.
						ob.salt = (ob.salt + 1) & ((1 << 16) - 1)
						if ob.salt == 0 {
							ob.salt++
						}
						// Return obstacle to free list.
						ob.next = this.m_nextFreeObstacle
						this.m_nextFreeObstacle = ob
					}
				}
			}
		}
	}

	if upToDate != nil {
		*upToDate = this.m_nupdate == 0 && this.m_nreqs == 0
	}

	return status
}

func (this *DtTileCache) buildNavMeshTilesAt(tx, ty int32, navmesh *detour.DtNavMesh) detour.DtStatus {
	const MAX_TILES = int32(32)
	var tiles [MAX_TILES]DtCompressedTileRef
	ntiles := this.getTilesAt(tx, ty, tiles[:], MAX_TILES)

	for i := int32(0); i < ntiles; i++ {
		status := this.buildNavMeshTile(tiles[i], navmesh)
		if detour.DtStatusFailed(status) {
			return status
		}
	}

	return detour.DT_SUCCESS
}

func (this *DtTileCache) buildNavMeshTile(ref DtCompressedTileRef, navmesh *detour.DtNavMesh) detour.DtStatus {
	detour.DtAssert(this.m_tcomp != nil)

	idx := this.decodeTileIdTile(ref)
	if idx > uint32(this.m_params.maxTiles) {
		return detour.DT_FAILURE | detour.DT_INVALID_PARAM
	}
	tile := &this.m_tiles[idx]
	salt := this.decodeTileIdSalt(ref)
	if tile.salt != salt {
		return detour.DT_FAILURE | detour.DT_INVALID_PARAM
	}

	var bc NavMeshTileBuildContext
	walkableClimbVx := int32(this.m_params.walkableClimb / this.m_params.ch)
	var status detour.DtStatus

	// Decompress tile layer data.
	status, bc.layer = dtDecompressTileCacheLayer(this.m_tcomp, tile.data, tile.dataSize)
	if detour.DtStatusFailed(status) {
		return status
	}

	// Rasterize obstacles.
	for i := int32(0); i < this.m_params.maxObstacles; i++ {
		ob := &this.m_obstacles[i]
		if ob.state == uint8(DT_OBSTACLE_EMPTY) || ob.state == uint8(DT_OBSTACLE_REMOVING) {
			continue
		}
		if contains(ob.touched[:], int32(ob.ntouched), ref) {
			if ob._type == uint8(DT_OBSTACLE_CYLINDER) {
				dtMarkCylinderArea(bc.layer, tile.header.bmin[:], this.m_params.cs, this.m_params.ch,
					ob.cylinder.pos[:], ob.cylinder.radius, ob.cylinder.height, 0)
			} else if ob._type == uint8(DT_OBSTACLE_BOX) {
				dtMarkBoxArea1(bc.layer, tile.header.bmin[:], this.m_params.cs, this.m_params.ch,
					ob.box.bmin[:], ob.box.bmax[:], 0)
			} else if ob._type == uint8(DT_OBSTACLE_ORIENTED_BOX) {
				dtMarkBoxArea2(bc.layer, tile.header.bmin[:], this.m_params.cs, this.m_params.ch,
					ob.orientedBox.center[:], ob.orientedBox.halfExtents[:], ob.orientedBox.rotAux[:], 0)
			}
		}
	}

	// Build navmesh
	status = dtBuildTileCacheRegions(bc.layer, walkableClimbVx)
	if detour.DtStatusFailed(status) {
		return status
	}

	bc.lcset = DtAllocTileCacheContourSet()
	if bc.lcset == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}
	status = dtBuildTileCacheContours(bc.layer, walkableClimbVx,
		this.m_params.maxSimplificationError, bc.lcset)
	if detour.DtStatusFailed(status) {
		return status
	}

	bc.lmesh = DtAllocTileCachePolyMesh()
	if bc.lmesh == nil {
		return detour.DT_FAILURE | detour.DT_OUT_OF_MEMORY
	}
	status = dtBuildTileCachePolyMesh(bc.lcset, bc.lmesh)
	if detour.DtStatusFailed(status) {
		return status
	}

	// Early out if the mesh tile is empty.
	if bc.lmesh.npolys == 0 {
		// Remove existing tile.
		navmesh.RemoveTile(navmesh.GetTileRefAt(tile.header.tx, tile.header.ty, tile.header.tlayer), nil, nil)
		return detour.DT_SUCCESS
	}

	var params detour.DtNavMeshCreateParams

	params.Verts = bc.lmesh.verts
	params.VertCount = bc.lmesh.nverts
	params.Polys = bc.lmesh.polys
	params.PolyAreas = bc.lmesh.areas
	params.PolyFlags = bc.lmesh.flags
	params.PolyCount = bc.lmesh.npolys
	params.Nvp = detour.DT_VERTS_PER_POLYGON
	params.WalkableHeight = this.m_params.walkableHeight
	params.WalkableRadius = this.m_params.walkableRadius
	params.WalkableClimb = this.m_params.walkableClimb
	params.TileX = tile.header.tx
	params.TileY = tile.header.ty
	params.TileLayer = tile.header.tlayer
	params.Cs = this.m_params.cs
	params.Ch = this.m_params.ch
	params.BuildBvTree = false
	detour.DtVcopy(params.Bmin[:], tile.header.bmin[:])
	detour.DtVcopy(params.Bmax[:], tile.header.bmax[:])

	if this.m_tmproc != nil {
		this.m_tmproc.process(&params, bc.lmesh.areas[:], bc.lmesh.flags[:])
	}

	var navData []byte
	var navDataSize int
	if !detour.DtCreateNavMeshData(&params, &navData, &navDataSize) {
		return detour.DT_FAILURE
	}

	// Remove existing tile.
	navmesh.RemoveTile(navmesh.GetTileRefAt(tile.header.tx, tile.header.ty, tile.header.tlayer), nil, nil)

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

func (this *DtTileCache) calcTightTileBounds(header *DtTileCacheLayerHeader, bmin, bmax []float32) {
	cs := this.m_params.cs
	bmin[0] = header.bmin[0] + float32(header.minx)*cs
	bmin[1] = header.bmin[1]
	bmin[2] = header.bmin[2] + float32(header.miny)*cs
	bmax[0] = header.bmin[0] + float32(header.maxx+1)*cs
	bmax[1] = header.bmax[1]
	bmax[2] = header.bmin[2] + float32(header.maxy+1)*cs
}

func (this *DtTileCache) getObstacleBounds(ob *DtTileCacheObstacle, bmin, bmax []float32) {
	if ob._type == uint8(DT_OBSTACLE_CYLINDER) {
		cl := &ob.cylinder

		bmin[0] = cl.pos[0] - cl.radius
		bmin[1] = cl.pos[1]
		bmin[2] = cl.pos[2] - cl.radius
		bmax[0] = cl.pos[0] + cl.radius
		bmax[1] = cl.pos[1] + cl.height
		bmax[2] = cl.pos[2] + cl.radius
	} else if ob._type == uint8(DT_OBSTACLE_BOX) {
		detour.DtVcopy(bmin, ob.box.bmin[:])
		detour.DtVcopy(bmax, ob.box.bmax[:])
	} else if ob._type == uint8(DT_OBSTACLE_ORIENTED_BOX) {
		orientedBox := &ob.orientedBox

		maxr := 1.41 * detour.DtMaxFloat32(orientedBox.halfExtents[0], orientedBox.halfExtents[2])
		bmin[0] = orientedBox.center[0] - maxr
		bmax[0] = orientedBox.center[0] + maxr
		bmin[1] = orientedBox.center[1] - orientedBox.halfExtents[1]
		bmax[1] = orientedBox.center[1] + orientedBox.halfExtents[1]
		bmin[2] = orientedBox.center[2] - maxr
		bmax[2] = orientedBox.center[2] + maxr
	}
}
