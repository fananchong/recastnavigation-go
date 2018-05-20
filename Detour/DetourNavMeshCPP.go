package detour

import (
	"bytes"
	"encoding/binary"
)

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
	this.m_maxTiles = params.MaxTiles
	this.m_tileLutSize = DtNextPow2(params.MaxTiles / 4)
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
	return DT_SUCCESS
}
