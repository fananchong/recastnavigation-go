package detour

import (
	"bytes"
	"encoding/binary"
	"unsafe"
)

func computeTileHash(x, y, mask int32) int32 {
	h1 := uint32(0x8da6b343) // Large multiplicative constants;
	h2 := uint32(0xd8163841) // here arbitrarily chosen primes
	n := h1*uint32(x) + h2*uint32(y)
	return int32(n & uint32(mask))
}

/**
@class dtNavMesh

The navigation mesh consists of one or more tiles defining three primary types of structural data:

A polygon mesh which defines most of the navigation graph. (See rcPolyMesh for its structure.)
A detail mesh used for determining surface height on the polygon mesh. (See rcPolyMeshDetail for its structure.)
Off-mesh connections, which define custom point-to-point edges within the navigation graph.

The general build process is as follows:

-# Create rcPolyMesh and rcPolyMeshDetail data using the Recast build pipeline.
-# Optionally, create off-mesh connection data.
-# Combine the source data into a dtNavMeshCreateParams structure.
-# Create a tile data array using dtCreateNavMeshData().
-# Allocate at dtNavMesh object and initialize it. (For single tile navigation meshes,
   the tile data is loaded during this step.)
-# For multi-tile navigation meshes, load the tile data using dtNavMesh::addTile().

Notes:

- This class is usually used in conjunction with the dtNavMeshQuery class for pathfinding.
- Technically, all navigation meshes are tiled. A 'solo' mesh is simply a navigation mesh initialized
  to have only a single tile.
- This class does not implement any asynchronous methods. So the ::dtStatus result of all methods will
  always contain either a success or failure flag.

@see dtNavMeshQuery, dtCreateNavMeshData, dtNavMeshCreateParams, #dtAllocNavMesh, #dtFreeNavMesh
*/
func (this *DtNavMesh) constructor() {

}

func (this *DtNavMesh) destructor() {
	for i := 0; i < int(this.m_maxTiles); i++ {
		if (this.m_tiles[i].Flags & DT_TILE_FREE_DATA) != 0 {
			this.m_tiles[i].Data = nil
			this.m_tiles[i].DataSize = 0
		}
	}
	this.m_posLookup = nil
	this.m_tiles = nil
}

/// @{
/// @name Initialization and Tile Management

/// Initializes the navigation mesh for tiled use.
///  @param[in]	params		Initialization parameters.
/// @return The status flags for the operation.
func (this *DtNavMesh) Init(params *DtNavMeshParams) DtStatus {
	this.m_params = *params
	DtVcopy(this.m_orig[:], params.Orig[:])
	this.m_tileWidth = params.TileWidth
	this.m_tileHeight = params.TileHeight

	// Init tiles
	this.m_maxTiles = int32(params.MaxTiles)
	this.m_tileLutSize = int32(DtNextPow2(params.MaxTiles / 4))
	if this.m_tileLutSize == 0 {
		this.m_tileLutSize = 1
	}
	this.m_tileLutMask = this.m_tileLutSize - 1

	this.m_tiles = make([]DtMeshTile, this.m_maxTiles)
	if this.m_tiles == nil {
		return DT_FAILURE | DT_OUT_OF_MEMORY
	}
	this.m_posLookup = make([]*DtMeshTile, this.m_tileLutSize)
	if this.m_posLookup == nil {
		return DT_FAILURE | DT_OUT_OF_MEMORY
	}

	this.m_nextFree = nil
	for i := int(this.m_maxTiles - 1); i >= 0; i-- {
		this.m_tiles[i].Salt = 1
		this.m_tiles[i].Next = this.m_nextFree
		this.m_nextFree = &this.m_tiles[i]
	}

	// Init ID generator values.
	this.m_tileBits = DtIlog2(DtNextPow2(params.MaxTiles))
	this.m_polyBits = DtIlog2(DtNextPow2(params.MaxPolys))
	// Only allow 31 salt bits, since the salt mask is calculated using 32bit uint and it will overflow.
	this.m_saltBits = DtMinUInt32(31, 32-this.m_tileBits-this.m_polyBits)

	if this.m_saltBits < 10 {
		return DT_FAILURE | DT_INVALID_PARAM
	}
	return DT_SUCCESS
}

/// Initializes the navigation mesh for single tile use.
///  @param[in]	data		Data of the new tile. (See: #dtCreateNavMeshData)
///  @param[in]	dataSize	The data size of the new tile.
///  @param[in]	flags		The tile flags. (See: #dtTileFlags)
/// @return The status flags for the operation.
///  @see dtCreateNavMeshData
func (this *DtNavMesh) Init2(data []byte, dataSize int, flags DtTileFlags) DtStatus {
	// Make sure the data is in right format.
	reader := bytes.NewReader(data)
	header := &DtMeshHeader{}
	if err := binary.Read(reader, binary.LittleEndian, header); err != nil {
		return DT_FAILURE | DT_INVALID_PARAM
	}
	if header.Magic != DT_NAVMESH_MAGIC {
		return DT_FAILURE | DT_WRONG_MAGIC
	}
	if header.Version != DT_NAVMESH_VERSION {
		return DT_FAILURE | DT_WRONG_VERSION
	}
	var params DtNavMeshParams
	DtVcopy(params.Orig[:], header.Bmin[:])
	params.TileWidth = header.Bmax[0] - header.Bmin[0]
	params.TileHeight = header.Bmax[2] - header.Bmin[2]
	params.MaxTiles = 1
	params.MaxPolys = uint32(header.PolyCount)

	status := this.Init(&params)
	if DtStatusFailed(status) {
		return status
	}
	return this.AddTile(data, dataSize, flags, 0, nil)
}

/// Adds a tile to the navigation mesh.
///  @param[in]		data		Data for the new tile mesh. (See: #dtCreateNavMeshData)
///  @param[in]		dataSize	Data size of the new tile mesh.
///  @param[in]		flags		Tile flags. (See: #dtTileFlags)
///  @param[in]		lastRef		The desired reference for the tile. (When reloading a tile.) [opt] [Default: 0]
///  @param[out]	result		The tile reference. (If the tile was succesfully added.) [opt]
/// @return The status flags for the operation.

/// @par
///
/// The add operation will fail if the data is in the wrong format, the allocated tile
/// space is full, or there is a tile already at the specified reference.
///
/// The lastRef parameter is used to restore a tile with the same tile
/// reference it had previously used.  In this case the #dtPolyRef's for the
/// tile will be restored to the same values they were before the tile was
/// removed.
///
/// The nav mesh assumes exclusive access to the data passed and will make
/// changes to the dynamic portion of the data. For that reason the data
/// should not be reused in other nav meshes until the tile has been successfully
/// removed from this nav mesh.
///
/// @see dtCreateNavMeshData, #removeTile
func (this *DtNavMesh) AddTile(data []byte, dataSize int, flags DtTileFlags, lastRef DtTileRef, result *DtTileRef) DtStatus {
	// Make sure the data is in right format.
	reader := bytes.NewReader(data)
	header := &DtMeshHeader{}
	if err := binary.Read(reader, binary.LittleEndian, header); err != nil {
		return DT_FAILURE | DT_INVALID_PARAM
	}

	if header.Magic != DT_NAVMESH_MAGIC {
		return DT_FAILURE | DT_WRONG_MAGIC
	}
	if header.Version != DT_NAVMESH_VERSION {
		return DT_FAILURE | DT_WRONG_VERSION
	}

	// Make sure the location is free.
	if this.GetTileAt(header.X, header.Y, header.Layer) != nil {
		return DT_FAILURE | DT_ALREADY_OCCUPIED
	}

	// Allocate a tile.
	var tile *DtMeshTile = nil
	if lastRef == 0 {
		if this.m_nextFree != nil {
			tile = this.m_nextFree
			this.m_nextFree = tile.Next
			tile.Next = nil
		}
	} else {
		// Try to relocate the tile to specific index with same salt.
		tileIndex := this.DecodePolyIdTile(DtPolyRef(lastRef))
		if tileIndex >= uint32(this.m_maxTiles) {
			return DT_FAILURE | DT_OUT_OF_MEMORY
		}
		// Try to find the specific tile id from the free list.
		target := &this.m_tiles[tileIndex]
		var prev *DtMeshTile = nil
		tile = this.m_nextFree
		for tile != nil && tile != target {
			prev = tile
			tile = tile.Next
		}
		// Could not find the correct location.
		if tile != target {
			return DT_FAILURE | DT_OUT_OF_MEMORY
		}
		// Remove from freelist
		if prev == nil {
			this.m_nextFree = tile.Next
		} else {
			prev.Next = tile.Next
		}
		// Restore salt.
		tile.Salt = this.DecodePolyIdSalt(DtPolyRef(lastRef))
	}

	// Make sure we could allocate a tile.
	if tile == nil {
		return DT_FAILURE | DT_OUT_OF_MEMORY
	}
	// Insert tile into the position lut.
	h := computeTileHash(header.X, header.Y, int32(this.m_tileLutMask))
	tile.Next = this.m_posLookup[h]
	this.m_posLookup[h] = tile

	// Patch header pointers.
	headerSize := DtAlign4(int(unsafe.Sizeof(DtMeshHeader{})))
	vertsSize := DtAlign4(int(unsafe.Sizeof(float32(1.0))) * 3 * int(header.VertCount))
	polysSize := DtAlign4(int(unsafe.Sizeof(DtPoly{})) * int(header.PolyCount))
	linksSize := DtAlign4(int(unsafe.Sizeof(DtLink{})) * int(header.MaxLinkCount))
	detailMeshesSize := DtAlign4(int(unsafe.Sizeof(DtPolyDetail{})) * int(header.DetailMeshCount))
	detailVertsSize := DtAlign4(int(unsafe.Sizeof(float32(1.0))) * 3 * int(header.DetailVertCount))
	detailTrisSize := DtAlign4(int(unsafe.Sizeof(uint8(1))) * 4 * int(header.DetailTriCount))
	bvtreeSize := DtAlign4(int(unsafe.Sizeof(DtBVNode{})) * int(header.BvNodeCount))
	offMeshLinksSize := DtAlign4(int(unsafe.Sizeof(DtOffMeshConnection{})) * int(header.OffMeshConCount))

	d := 0 + headerSize
	s := int(unsafe.Sizeof(float32(1.0)))
	for i := 0; i < 3*int(header.VertCount); i++ {
		reader := bytes.NewReader(data[d:])
		v := float32(1.0)
		if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
			return DT_FAILURE | DT_INVALID_PARAM
		}
		tile.Verts = append(tile.Verts, v)
		d += s
	}

	//	tile->polys = dtGetThenAdvanceBufferPointer<dtPoly>(d, polysSize);
	//	tile->links = dtGetThenAdvanceBufferPointer<dtLink>(d, linksSize);
	//	tile->detailMeshes = dtGetThenAdvanceBufferPointer<dtPolyDetail>(d, detailMeshesSize);
	//	tile->detailVerts = dtGetThenAdvanceBufferPointer<float>(d, detailVertsSize);
	//	tile->detailTris = dtGetThenAdvanceBufferPointer<unsigned char>(d, detailTrisSize);
	//	tile->bvTree = dtGetThenAdvanceBufferPointer<dtBVNode>(d, bvtreeSize);
	//	tile->offMeshCons = dtGetThenAdvanceBufferPointer<dtOffMeshConnection>(d, offMeshLinksSize);

	//	// If there are no items in the bvtree, reset the tree pointer.
	//	if (!bvtreeSize)
	//		tile->bvTree = 0;

	//	// Build links freelist
	//	tile->linksFreeList = 0;
	//	tile->links[header->maxLinkCount-1].next = DT_NULL_LINK;
	//	for (int i = 0; i < header->maxLinkCount-1; ++i)
	//		tile->links[i].next = i+1;

	//	// Init tile.
	//	tile->header = header;
	//	tile->data = data;
	//	tile->dataSize = dataSize;
	//	tile->flags = flags;

	//	connectIntLinks(tile);

	//	// Base off-mesh connections to their starting polygons and connect connections inside the tile.
	//	baseOffMeshLinks(tile);
	//	connectExtOffMeshLinks(tile, tile, -1);

	//	// Create connections with neighbour tiles.
	//	static const int MAX_NEIS = 32;
	//	dtMeshTile* neis[MAX_NEIS];
	//	int nneis;

	//	// Connect with layers in current tile.
	//	nneis = getTilesAt(header->x, header->y, neis, MAX_NEIS);
	//	for (int j = 0; j < nneis; ++j)
	//	{
	//		if (neis[j] == tile)
	//			continue;

	//		connectExtLinks(tile, neis[j], -1);
	//		connectExtLinks(neis[j], tile, -1);
	//		connectExtOffMeshLinks(tile, neis[j], -1);
	//		connectExtOffMeshLinks(neis[j], tile, -1);
	//	}

	//	// Connect with neighbour tiles.
	//	for (int i = 0; i < 8; ++i)
	//	{
	//		nneis = getNeighbourTilesAt(header->x, header->y, i, neis, MAX_NEIS);
	//		for (int j = 0; j < nneis; ++j)
	//		{
	//			connectExtLinks(tile, neis[j], i);
	//			connectExtLinks(neis[j], tile, dtOppositeTile(i));
	//			connectExtOffMeshLinks(tile, neis[j], i);
	//			connectExtOffMeshLinks(neis[j], tile, dtOppositeTile(i));
	//		}
	//	}

	//	if (result)
	//		*result = getTileRef(tile);

	return DT_SUCCESS
}

/// Gets the tile at the specified grid location.
///  @param[in]	x		The tile's x-location. (x, y, layer)
///  @param[in]	y		The tile's y-location. (x, y, layer)
///  @param[in]	layer	The tile's layer. (x, y, layer)
/// @return The tile, or null if the tile does not exist.
func (this *DtNavMesh) GetTileAt(x, y, layer int32) *DtMeshTile {
	// Find tile based on hash.
	h := computeTileHash(x, y, int32(this.m_tileLutMask))
	tile := this.m_posLookup[h]
	for tile != nil {
		if tile.Header != nil &&
			tile.Header.X == int32(x) &&
			tile.Header.Y == int32(y) &&
			tile.Header.Layer == int32(layer) {
			return tile
		}
		tile = tile.Next
	}
	return nil
}
