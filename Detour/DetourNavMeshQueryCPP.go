package detour

/// @class dtQueryFilter
///
/// <b>The Default Implementation</b>
///
/// At construction: All area costs default to 1.0.  All flags are included
/// and none are excluded.
///
/// If a polygon has both an include and an exclude flag, it will be excluded.
///
/// The way filtering works, a navigation mesh polygon must have at least one flag
/// set to ever be considered by a query. So a polygon with no flags will never
/// be considered.
///
/// Setting the include flags to 0 will result in all polygons being excluded.
///
/// <b>Custom Implementations</b>
///
/// DT_VIRTUAL_QUERYFILTER must be defined in order to extend this class.
///
/// Implement a custom query filter by overriding the virtual passFilter()
/// and getCost() functions. If this is done, both functions should be as
/// fast as possible. Use cached local copies of data rather than accessing
/// your own objects where possible.
///
/// Custom implementations do not need to adhere to the flags or cost logic
/// used by the default implementation.
///
/// In order for A* searches to work properly, the cost should be proportional to
/// the travel distance. Implementing a cost modifier less than 1.0 is likely
/// to lead to problems during pathfinding.
///
/// @see dtNavMeshQuery

func (this *DtQueryFilter) constructor() {
	this.m_includeFlags = 0xffff
	this.m_excludeFlags = 0
	for i := 0; i < DT_MAX_AREAS; i++ {
		this.m_areaCost[i] = 1.0
	}
}

func (this *DtQueryFilter) destructor() {
}

func (this *DtQueryFilter) PassFilter(_ DtPolyRef, _ *DtMeshTile, poly *DtPoly) bool {
	return (poly.Flags&this.m_includeFlags) != 0 && (poly.Flags&this.m_excludeFlags) == 0
}

func (this *DtQueryFilter) GetCost(pa, pb []float32,
	_ DtPolyRef, _ *DtMeshTile, _ *DtPoly,
	_ DtPolyRef, _ *DtMeshTile, curPoly *DtPoly,
	_ DtPolyRef, _ *DtMeshTile, _ *DtPoly) float32 {
	return DtVdist(pa, pb) * this.m_areaCost[curPoly.GetArea()]
}

const H_SCALE float32 = 0.999 // Search heuristic scale.

//////////////////////////////////////////////////////////////////////////////////////////

/// @class dtNavMeshQuery
///
/// For methods that support undersized buffers, if the buffer is too small
/// to hold the entire result set the return status of the method will include
/// the #DT_BUFFER_TOO_SMALL flag.
///
/// Constant member functions can be used by multiple clients without side
/// effects. (E.g. No change to the closed list. No impact on an in-progress
/// sliced path query. Etc.)
///
/// Walls and portals: A @e wall is a polygon segment that is
/// considered impassable. A @e portal is a passable segment between polygons.
/// A portal may be treated as a wall based on the dtQueryFilter used for a query.
///
/// @see dtNavMesh, dtQueryFilter, #dtAllocNavMeshQuery(), #dtAllocNavMeshQuery()

func (this *DtNavMeshQuery) constructor() {

}

func (this *DtNavMeshQuery) destructor() {
	if this.m_tinyNodePool != nil {
		DtFreeNodePool(this.m_tinyNodePool)
		this.m_tinyNodePool = nil
	}
	if this.m_nodePool != nil {
		DtFreeNodePool(this.m_nodePool)
		this.m_nodePool = nil
	}
	if this.m_openList != nil {
		DtFreeNodeQueue(this.m_openList)
		this.m_openList = nil
	}
}
