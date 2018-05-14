package detour

/// A handle to a polygon within a navigation mesh tile.
/// @ingroup detour
type DtPolyRef uint32

/// A handle to a tile within a navigation mesh.
/// @ingroup detour
type DtTileRef uint32

/// The maximum number of vertices per navigation polygon.
/// @ingroup detour
const DT_VERTS_PER_POLYGON int32 = 6

/// @{
/// @name Tile Serialization Constants
/// These constants are used to detect whether a navigation tile's data
/// and state format is compatible with the current build.
///

/// A magic number used to detect compatibility of navigation tile data.
const DT_NAVMESH_MAGIC int32 = 'D'<<24 | 'N'<<16 | 'A'<<8 | 'V'

/// A version number used to detect compatibility of navigation tile data.
const DT_NAVMESH_VERSION int32 = 7

/// A magic number used to detect the compatibility of navigation tile states.
const DT_NAVMESH_STATE_MAGIC int32 = 'D'<<24 | 'N'<<16 | 'M'<<8 | 'S'

/// A version number used to detect compatibility of navigation tile states.
const DT_NAVMESH_STATE_VERSION int32 = 1

/// @}

/// A flag that indicates that an entity links to an external entity.
/// (E.g. A polygon edge is a portal that links to another polygon.)
const DT_EXT_LINK uint16 = 0x8000

/// A value that indicates the entity does not link to anything.
const DT_NULL_LINK uint32 = 0xffffffff

/// A flag that indicates that an off-mesh connection can be traversed in both directions. (Is bidirectional.)
const DT_OFFMESH_CON_BIDIR uint32 = 1

/// The maximum number of user defined area ids.
/// @ingroup detour
const DT_MAX_AREAS int32 = 64

/// Tile flags used for various functions and fields.
/// For an example, see dtNavMesh::addTile().
type DtTileFlags int

const (
	/// The navigation mesh owns the tile memory and is responsible for freeing it.
	DT_TILE_FREE_DATA DtTileFlags = 0x01
)

/// Vertex flags returned by dtNavMeshQuery::findStraightPath.
type DtStraightPathFlags int

const (
	DT_STRAIGHTPATH_START              DtStraightPathFlags = 0x01 ///< The vertex is the start position in the path.
	DT_STRAIGHTPATH_END                DtStraightPathFlags = 0x02 ///< The vertex is the end position in the path.
	DT_STRAIGHTPATH_OFFMESH_CONNECTION DtStraightPathFlags = 0x04 ///< The vertex is the start of an off-mesh connection.
)

/// Options for dtNavMeshQuery::findStraightPath.
type DtStraightPathOptions int

const (
	DT_STRAIGHTPATH_AREA_CROSSINGS DtStraightPathOptions = 0x01 ///< Add a vertex at every polygon edge crossing where area changes.
	DT_STRAIGHTPATH_ALL_CROSSINGS  DtStraightPathOptions = 0x02 ///< Add a vertex at every polygon edge crossing.
)

/// Options for dtNavMeshQuery::initSlicedFindPath and updateSlicedFindPath
type DtFindPathOptions int

const (
	DT_FINDPATH_ANY_ANGLE DtFindPathOptions = 0x02 ///< use raycasts during pathfind to "shortcut" (raycast still consider costs)
)

/// Options for dtNavMeshQuery::raycast
type DtRaycastOptions int

const (
	DT_RAYCAST_USE_COSTS DtRaycastOptions = 0x01 ///< Raycast should calculate movement cost along the ray and fill RaycastHit::cost
)

/// Limit raycasting during any angle pahfinding
/// The limit is given as a multiple of the character radius
const DT_RAY_CAST_LIMIT_PROPORTIONS float32 = 50.0

/// Flags representing the type of a navigation mesh polygon.
type DtPolyTypes int

const (
	/// The polygon is a standard convex polygon that is part of the surface of the mesh.
	DT_POLYTYPE_GROUND DtPolyTypes = 0
	/// The polygon is an off-mesh connection consisting of two vertices.
	DT_POLYTYPE_OFFMESH_CONNECTION DtPolyTypes = 1
)

/// Defines a polygon within a dtMeshTile object.
/// @ingroup detour
type DtPoly struct {
	/// Index to first link in linked list. (Or #DT_NULL_LINK if there is no link.)
	FirstLink uint32

	/// The indices of the polygon's vertices.
	/// The actual vertices are located in dtMeshTile::verts.
	Verts [DT_VERTS_PER_POLYGON]uint16

	/// Packed data representing neighbor polygons references and flags for each edge.
	Neis [DT_VERTS_PER_POLYGON]uint16

	/// The user defined polygon flags.
	Flags uint16

	/// The number of vertices in the polygon.
	VertCount uint8

	/// The bit packed area id and polygon type.
	/// @note Use the structure's set and get methods to acess this value.
	AreaAndtype uint8
}

/// Sets the user defined area id. [Limit: < #DT_MAX_AREAS]
func (this *DtPoly) SetArea(a uint8) { this.AreaAndtype = (this.AreaAndtype & 0xc0) | (a & 0x3f) }

/// Sets the polygon type. (See: #dtPolyTypes.)
func (this *DtPoly) SetType(t uint8) { this.AreaAndtype = (this.AreaAndtype & 0x3f) | (t << 6) }

/// Gets the user defined area id.
func (this *DtPoly) GetArea() uint8 { return this.AreaAndtype & 0x3f }

/// Gets the polygon type. (See: #dtPolyTypes)
func (this *DtPoly) GetType() uint8 { return this.AreaAndtype >> 6 }

/// Defines the location of detail sub-mesh data within a dtMeshTile.
type DtPolyDetail struct {
	VertBase  uint32 ///< The offset of the vertices in the dtMeshTile::detailVerts array.
	TriBase   uint32 ///< The offset of the triangles in the dtMeshTile::detailTris array.
	VertCount uint8  ///< The number of vertices in the sub-mesh.
	TriCount  uint8  ///< The number of triangles in the sub-mesh.
}

/// Defines a link between polygons.
/// @note This structure is rarely if ever used by the end user.
/// @see dtMeshTile
type DtLink struct {
	Ref  DtPolyRef ///< Neighbour reference. (The neighbor that is linked to.)
	Next uint32    ///< Index of the next link.
	Edge uint8     ///< Index of the polygon edge that owns this link.
	Side uint8     ///< If a boundary link, defines on which side the link is.
	Bmin uint8     ///< If a boundary link, defines the minimum sub-edge area.
	Bmax uint8     ///< If a boundary link, defines the maximum sub-edge area.
}

/// Bounding volume node.
/// @note This structure is rarely if ever used by the end user.
/// @see dtMeshTile
type DtBVNode struct {
	Bmin [3]uint8 ///< Minimum bounds of the node's AABB. [(x, y, z)]
	Bmax [3]uint8 ///< Maximum bounds of the node's AABB. [(x, y, z)]
	I    int32    ///< The node's index. (Negative for escape sequence.)
}

/// Defines an navigation mesh off-mesh connection within a dtMeshTile object.
/// An off-mesh connection is a user defined traversable connection made up to two vertices.
type DtOffMeshConnection struct {
	/// The endpoints of the connection. [(ax, ay, az, bx, by, bz)]
	Pos [6]float32

	/// The radius of the endpoints. [Limit: >= 0]
	Rad float32

	/// The polygon reference of the connection within the tile.
	Poly uint16

	/// Link flags.
	/// @note These are not the connection's user defined flags. Those are assigned via the
	/// connection's dtPoly definition. These are link flags used for internal purposes.
	Flags uint8

	/// End point side.
	Side uint8

	/// The id of the offmesh connection. (User assigned when the navigation mesh is built.)
	UserId uint32
}

/// Provides high level information related to a dtMeshTile object.
/// @ingroup detour
type DtMeshHeader struct {
	Magic           int32  ///< Tile magic number. (Used to identify the data format.)
	Version         int32  ///< Tile data format version number.
	X               int32  ///< The x-position of the tile within the dtNavMesh tile grid. (x, y, layer)
	Y               int32  ///< The y-position of the tile within the dtNavMesh tile grid. (x, y, layer)
	Layer           int32  ///< The layer of the tile within the dtNavMesh tile grid. (x, y, layer)
	UserId          uint32 ///< The user defined id of the tile.
	PolyCount       int32  ///< The number of polygons in the tile.
	VertCount       int32  ///< The number of vertices in the tile.
	MaxLinkCount    int32  ///< The number of allocated links.
	DetailMeshCount int32  ///< The number of sub-meshes in the detail mesh.

	/// The number of unique vertices in the detail mesh. (In addition to the polygon vertices.)
	DetailVertCount int32

	DetailTriCount  int32      ///< The number of triangles in the detail mesh.
	BvNodeCount     int32      ///< The number of bounding volume nodes. (Zero if bounding volumes are disabled.)
	OffMeshConCount int32      ///< The number of off-mesh connections.
	OffMeshBase     int32      ///< The index of the first polygon which is an off-mesh connection.
	WalkableHeight  float32    ///< The height of the agents using the tile.
	WalkableRadius  float32    ///< The radius of the agents using the tile.
	WalkableClimb   float32    ///< The maximum climb height of the agents using the tile.
	Bmin            [3]float32 ///< The minimum bounds of the tile's AABB. [(x, y, z)]
	Bmax            [3]float32 ///< The maximum bounds of the tile's AABB. [(x, y, z)]

	/// The bounding volume quantization factor.
	BvQuantFactor float32
}

/// Defines a navigation mesh tile.
/// @ingroup detour
type DtMeshTile struct {
	Salt uint32 ///< Counter describing modifications to the tile.

	LinksFreeList uint32         ///< Index to the next free link.
	Header        *DtMeshHeader  ///< The tile header.
	Polys         []DtPoly       ///< The tile polygons. [Size: dtMeshHeader::polyCount]
	Verts         []float32      ///< The tile vertices. [Size: dtMeshHeader::vertCount]
	Links         []DtLink       ///< The tile links. [Size: dtMeshHeader::maxLinkCount]
	DetailMeshes  []DtPolyDetail ///< The tile's detail sub-meshes. [Size: dtMeshHeader::detailMeshCount]

	/// The detail mesh's unique vertices. [(x, y, z) * dtMeshHeader::detailVertCount]
	DetailVerts []float32

	/// The detail mesh's triangles. [(vertA, vertB, vertC) * dtMeshHeader::detailTriCount]
	DetailTris []uint8

	/// The tile bounding volume nodes. [Size: dtMeshHeader::bvNodeCount]
	/// (Will be null if bounding volumes are disabled.)
	BvTree []DtBVNode

	OffMeshCons []DtOffMeshConnection ///< The tile off-mesh connections. [Size: dtMeshHeader::offMeshConCount]

	Data     []byte      ///< The tile data. (Not directly accessed under normal situations.)
	DataSize int32       ///< Size of the tile data.
	Flags    int32       ///< Tile flags. (See: #dtTileFlags)
	Next     *DtMeshTile ///< The next free tile, or the next tile in the spatial grid.
}

/// Configuration parameters used to define multi-tile navigation meshes.
/// The values are used to allocate space during the initialization of a navigation mesh.
/// @see dtNavMesh::init()
/// @ingroup detour
type DtNavMeshParams struct {
	Orig       [3]float32 ///< The world space origin of the navigation mesh's tile space. [(x, y, z)]
	TileWidth  float32    ///< The width of each tile. (Along the x-axis.)
	TileHeight float32    ///< The height of each tile. (Along the z-axis.)
	MaxTiles   int32      ///< The maximum number of tiles the navigation mesh can contain.
	MaxPolys   int32      ///< The maximum number of polygons each tile can contain.
}

/// A navigation mesh based on tiles of convex polygons.
/// @ingroup detour
type DtNavMesh struct {
	m_params                  DtNavMeshParams ///< Current initialization params. TODO: do not store this info twice.
	m_orig                    [3]float32      ///< Origin of the tile (0,0)
	m_tileWidth, m_tileHeight float32         ///< Dimensions of each tile.
	m_maxTiles                int32           ///< Max number of tiles.
	m_tileLutSize             int32           ///< Tile hash lookup size (must be pot).
	m_tileLutMask             int32           ///< Tile hash lookup mask.

	m_posLookup []*DtMeshTile ///< Tile hash lookup.
	m_nextFree  *DtMeshTile   ///< Freelist of tiles.
	m_tiles     []DtMeshTile  ///< List of tiles.

	m_saltBits uint32 ///< Number of salt bits in the tile ID.
	m_tileBits uint32 ///< Number of tile bits in the tile ID.
	m_polyBits uint32 ///< Number of poly bits in the tile ID.
}

/// Allocates a navigation mesh object using the Detour allocator.
/// @return A navigation mesh that is ready for initialization, or null on failure.
///  @ingroup detour
func DtAllocNavMesh() *DtNavMesh {
	navmesh := &DtNavMesh{}
	navmesh.constructor()
	return navmesh
}

/// Frees the specified navigation mesh object using the Detour allocator.
///  @param[in]	navmesh		A navigation mesh allocated using #dtAllocNavMesh
///  @ingroup detour
func DtFreeNavMesh(navmesh *DtNavMesh) {
	if navmesh == nil {
		return
	}
	navmesh.destructor()
}
