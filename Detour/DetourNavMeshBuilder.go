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

// Note: This header file's only purpose is to include define assert.
// Feel free to change the file and include your own implementation instead.

package detour

/// Represents the source data used to build an navigation mesh tile.
/// @ingroup detour
type DtNavMeshCreateParams struct {

	/// @name Polygon Mesh Attributes
	/// Used to create the base navigation graph.
	/// See #rcPolyMesh for details related to these attributes.
	/// @{

	Verts     []uint16 ///< The polygon mesh vertices. [(x, y, z) * #vertCount] [Unit: vx]
	VertCount int32    ///< The number vertices in the polygon mesh. [Limit: >= 3]
	Polys     []uint16 ///< The polygon data. [Size: #polyCount * 2 * #nvp]
	PolyFlags []uint16 ///< The user defined flags assigned to each polygon. [Size: #polyCount]
	PolyAreas []uint8  ///< The user defined area ids assigned to each polygon. [Size: #polyCount]
	PolyCount int32    ///< Number of polygons in the mesh. [Limit: >= 1]
	Nvp       int32    ///< Number maximum number of vertices per polygon. [Limit: >= 3]

	/// @}
	/// @name Height Detail Attributes (Optional)
	/// See #rcPolyMeshDetail for details related to these attributes.
	/// @{

	DetailMeshes     []uint32  ///< The height detail sub-mesh data. [Size: 4 * #polyCount]
	DetailVerts      []float32 ///< The detail mesh vertices. [Size: 3 * #detailVertsCount] [Unit: wu]
	DetailVertsCount int32     ///< The number of vertices in the detail mesh.
	DetailTris       []uint8   ///< The detail mesh triangles. [Size: 4 * #detailTriCount]
	DetailTriCount   int32     ///< The number of triangles in the detail mesh.

	/// @}
	/// @name Off-Mesh Connections Attributes (Optional)
	/// Used to define a custom point-to-point edge within the navigation graph, an
	/// off-mesh connection is a user defined traversable connection made up to two vertices,
	/// at least one of which resides within a navigation mesh polygon.
	/// @{

	/// Off-mesh connection vertices. [(ax, ay, az, bx, by, bz) * #offMeshConCount] [Unit: wu]
	OffMeshConVerts []float32
	/// Off-mesh connection radii. [Size: #offMeshConCount] [Unit: wu]
	OffMeshConRad []float32
	/// User defined flags assigned to the off-mesh connections. [Size: #offMeshConCount]
	OffMeshConFlags []uint16
	/// User defined area ids assigned to the off-mesh connections. [Size: #offMeshConCount]
	OffMeshConAreas []uint8
	/// The permitted travel direction of the off-mesh connections. [Size: #offMeshConCount]
	///
	/// 0 = Travel only from endpoint A to endpoint B.<br/>
	/// #DT_OFFMESH_CON_BIDIR = Bidirectional travel.
	OffMeshConDir []uint8
	/// The user defined ids of the off-mesh connection. [Size: #offMeshConCount]
	OffMeshConUserID []uint32
	/// The number of off-mesh connections. [Limit: >= 0]
	OffMeshConCount int32

	/// @}
	/// @name Tile Attributes
	/// @note The tile grid/layer data can be left at zero if the destination is a single tile mesh.
	/// @{

	UserId    uint32     ///< The user defined id of the tile.
	TileX     int32      ///< The tile's x-grid location within the multi-tile destination mesh. (Along the x-axis.)
	TileY     int32      ///< The tile's y-grid location within the multi-tile desitation mesh. (Along the z-axis.)
	TileLayer int32      ///< The tile's layer within the layered destination mesh. [Limit: >= 0] (Along the y-axis.)
	Bmin      [3]float32 ///< The minimum bounds of the tile. [(x, y, z)] [Unit: wu]
	Bmax      [3]float32 ///< The maximum bounds of the tile. [(x, y, z)] [Unit: wu]

	/// @}
	/// @name General Configuration Attributes
	/// @{

	WalkableHeight float32 ///< The agent height. [Unit: wu]
	WalkableRadius float32 ///< The agent radius. [Unit: wu]
	WalkableClimb  float32 ///< The agent maximum traversable ledge. (Up/Down) [Unit: wu]
	Cs             float32 ///< The xz-plane cell size of the polygon mesh. [Limit: > 0] [Unit: wu]
	Ch             float32 ///< The y-axis cell height of the polygon mesh. [Limit: > 0] [Unit: wu]

	/// True if a bounding volume tree should be built for the tile.
	/// @note The BVTree is not normally needed for layered navigation meshes.
	BuildBvTree bool

	/// @}
}

// This section contains detailed documentation for members that don't have
// a source file. It reduces clutter in the main section of the header.

/**

@struct dtNavMeshCreateParams
@par

This structure is used to marshal data between the Recast mesh generation pipeline and Detour navigation components.

See the rcPolyMesh and rcPolyMeshDetail documentation for detailed information related to mesh structure.

Units are usually in voxels (vx) or world units (wu). The units for voxels, grid size, and cell size
are all based on the values of #cs and #ch.

The standard navigation mesh build process is to create tile data using dtCreateNavMeshData, then add the tile
to a navigation mesh using either the dtNavMesh single tile <tt>init()</tt> function or the dtNavMesh::addTile()
function.

@see dtCreateNavMeshData

*/
