package detour

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
