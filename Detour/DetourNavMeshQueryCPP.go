package detour

import (
	"math"
	"unsafe"
)

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

/// Initializes the query object.
///  @param[in]		nav			Pointer to the dtNavMesh object to use for all queries.
///  @param[in]		maxNodes	Maximum number of search nodes. [Limits: 0 < value <= 65535]
/// @returns The status flags for the query.
func (this *DtNavMeshQuery) Init(nav *DtNavMesh, maxNodes int) DtStatus {
	/// @par
	///
	/// Must be the first function called after construction, before other
	/// functions are used.
	///
	/// This function can be used multiple times.
	if maxNodes > int(DT_NULL_IDX) || maxNodes > int((1<<DT_NODE_PARENT_BITS)-1) {
		return DT_FAILURE | DT_INVALID_PARAM
	}
	this.m_nav = nav

	if this.m_nodePool == nil || this.m_nodePool.GetMaxNodes() < uint32(maxNodes) {
		if this.m_nodePool != nil {
			DtFreeNodePool(this.m_nodePool)
			this.m_nodePool = nil
		}
		this.m_nodePool = DtAllocNodePool(uint32(maxNodes), DtNextPow2(uint32(maxNodes/4)))
		if this.m_nodePool == nil {
			return DT_FAILURE | DT_OUT_OF_MEMORY
		}
	} else {
		this.m_nodePool.Clear()
	}

	if this.m_tinyNodePool != nil {
		this.m_tinyNodePool = DtAllocNodePool(64, 32)
		if this.m_tinyNodePool == nil {
			return DT_FAILURE | DT_OUT_OF_MEMORY
		}
	} else {
		this.m_tinyNodePool.Clear()
	}

	if this.m_openList == nil || this.m_openList.GetCapacity() < maxNodes {
		if this.m_openList != nil {
			DtFreeNodeQueue(this.m_openList)
			this.m_openList = nil
		}
		this.m_openList = DtAllocNodeQueue(maxNodes)
		if this.m_openList == nil {
			return DT_FAILURE | DT_OUT_OF_MEMORY
		}
	} else {
		this.m_openList.Clear()
	}

	return DT_SUCCESS
}

/// Returns random location on navmesh.
/// Polygons are chosen weighted by area. The search runs in linear related to number of polygon.
///  @param[in]		filter			The polygon filter to apply to the query.
///  @param[in]		frand			Function returning a random number [0..1).
///  @param[out]	randomRef		The reference id of the random location.
///  @param[out]	randomPt		The random location.
/// @returns The status flags for the query.
func (this *DtNavMeshQuery) FindRandomPoint(filter *DtQueryFilter, frand func() float32,
	randomRef *DtPolyRef, randomPt []float32) DtStatus {
	DtAssert(this.m_nav != nil)

	// Randomly pick one tile. Assume that all tiles cover roughly the same area.
	var tile *DtMeshTile
	var tsum float32
	for i := 0; i < int(this.m_nav.GetMaxTiles()); i++ {
		t := this.m_nav.GetTile(i)
		if t == nil || t.Header == nil {
			continue
		}

		// Choose random tile using reservoi sampling.
		const area float32 = 1.0 // Could be tile area too.
		tsum += area
		u := frand()
		if u*tsum <= area {
			tile = t
		}
	}
	if tile == nil {
		return DT_FAILURE
	}
	// Randomly pick one polygon weighted by polygon area.
	var poly *DtPoly
	var polyRef DtPolyRef
	base := this.m_nav.GetPolyRefBase(tile)

	var areaSum float32
	for i := 0; i < int(tile.Header.PolyCount); i++ {
		p := &tile.Polys[i]
		// Do not return off-mesh connection polygons.
		if p.GetType() != DT_POLYTYPE_GROUND {
			continue
		}
		// Must pass filter
		ref := base | (DtPolyRef)(i)
		if !filter.PassFilter(ref, tile, p) {
			continue
		}
		// Calc area of the polygon.
		var polyArea float32
		for j := 2; j < int(p.VertCount); j++ {
			va := tile.Verts[p.Verts[0]*3:]
			vb := tile.Verts[p.Verts[j-1]*3:]
			vc := tile.Verts[p.Verts[j]*3:]
			polyArea += DtTriArea2D(va, vb, vc)
		}

		// Choose random polygon weighted by area, using reservoi sampling.
		areaSum += polyArea
		u := frand()
		if u*areaSum <= polyArea {
			poly = p
			polyRef = ref
		}
	}

	if poly == nil {
		return DT_FAILURE
	}
	// Randomly pick point on polygon.
	v := tile.Verts[poly.Verts[0]*3:]
	var verts [3 * DT_VERTS_PER_POLYGON]float32
	var areas [DT_VERTS_PER_POLYGON]float32
	DtVcopy(verts[0*3:], v)
	for j := 1; j < int(poly.VertCount); j++ {
		v = tile.Verts[poly.Verts[j]*3:]
		DtVcopy(verts[j*3:], v)
	}

	s := frand()
	t := frand()

	var pt [3]float32
	DtRandomPointInConvexPoly(verts[:], int(poly.VertCount), areas[:], s, t, pt[:])

	var h float32
	status := this.GetPolyHeight(polyRef, pt[:], &h)
	if DtStatusFailed(status) {
		return status
	}
	pt[1] = h

	DtVcopy(randomPt, pt[:])
	*randomRef = polyRef

	return DT_SUCCESS
}

/// Returns random location on navmesh within the reach of specified location.
/// Polygons are chosen weighted by area. The search runs in linear related to number of polygon.
/// The location is not exactly constrained by the circle, but it limits the visited polygons.
///  @param[in]		startRef		The reference id of the polygon where the search starts.
///  @param[in]		centerPos		The center of the search circle. [(x, y, z)]
///  @param[in]		filter			The polygon filter to apply to the query.
///  @param[in]		frand			Function returning a random number [0..1).
///  @param[out]	randomRef		The reference id of the random location.
///  @param[out]	randomPt		The random location. [(x, y, z)]
/// @returns The status flags for the query.
func (this *DtNavMeshQuery) FindRandomPointAroundCircle(startRef DtPolyRef, centerPos []float32, maxRadius float32,
	filter *DtQueryFilter, frand func() float32,
	randomRef *DtPolyRef, randomPt []float32) DtStatus {
	DtAssert(this.m_nav != nil)
	DtAssert(this.m_nodePool != nil)
	DtAssert(this.m_openList != nil)

	// Validate input
	if startRef == 0 || !this.m_nav.IsValidPolyRef(startRef) {
		return DT_FAILURE | DT_INVALID_PARAM
	}
	var startTile *DtMeshTile
	var startPoly *DtPoly
	this.m_nav.GetTileAndPolyByRefUnsafe(startRef, &startTile, &startPoly)
	if !filter.PassFilter(startRef, startTile, startPoly) {
		return DT_FAILURE | DT_INVALID_PARAM
	}
	this.m_nodePool.Clear()
	this.m_openList.Clear()

	startNode := this.m_nodePool.GetNode(startRef, 0)
	DtVcopy(startNode.Pos[:], centerPos)
	startNode.SetPidx(0)
	startNode.Cost = 0
	startNode.Total = 0
	startNode.Id = startRef
	startNode.SetFlags(DT_NODE_OPEN)
	this.m_openList.Push(startNode)

	status := DT_SUCCESS

	radiusSqr := DtSqrFloat32(maxRadius)
	var areaSum float32

	var randomTile *DtMeshTile
	var randomPoly *DtPoly
	var randomPolyRef DtPolyRef

	for !this.m_openList.Empty() {
		bestNode := this.m_openList.Pop()
		bestNode.SetFlags(bestNode.GetFlags() & ^DT_NODE_OPEN)
		bestNode.SetFlags(bestNode.GetFlags() | DT_NODE_CLOSED)

		// Get poly and tile.
		// The API input has been cheked already, skip checking internal data.
		bestRef := bestNode.Id
		var bestTile *DtMeshTile
		var bestPoly *DtPoly
		this.m_nav.GetTileAndPolyByRefUnsafe(bestRef, &bestTile, &bestPoly)

		// Place random locations on on ground.
		if bestPoly.GetType() == DT_POLYTYPE_GROUND {
			// Calc area of the polygon.
			var polyArea float32
			for j := 2; j < int(bestPoly.VertCount); j++ {
				va := bestTile.Verts[bestPoly.Verts[0]*3:]
				vb := bestTile.Verts[bestPoly.Verts[j-1]*3:]
				vc := bestTile.Verts[bestPoly.Verts[j]*3:]
				polyArea += DtTriArea2D(va, vb, vc)
			}
			// Choose random polygon weighted by area, using reservoi sampling.
			areaSum += polyArea
			u := frand()
			if u*areaSum <= polyArea {
				randomTile = bestTile
				randomPoly = bestPoly
				randomPolyRef = bestRef
			}
		}

		// Get parent poly and tile.
		var parentRef DtPolyRef
		var parentTile *DtMeshTile
		var parentPoly *DtPoly
		if bestNode.GetPidx() != 0 {
			parentRef = this.m_nodePool.GetNodeAtIdx(bestNode.GetPidx()).Id
		}
		if parentRef != 0 {
			this.m_nav.GetTileAndPolyByRefUnsafe(parentRef, &parentTile, &parentPoly)
		}
		for i := bestPoly.FirstLink; i != DT_NULL_LINK; i = bestTile.Links[i].Next {
			link := &bestTile.Links[i]
			neighbourRef := link.Ref
			// Skip invalid neighbours and do not follow back to parent.
			if neighbourRef == 0 || neighbourRef == parentRef {
				continue
			}
			// Expand to neighbour
			var neighbourTile *DtMeshTile
			var neighbourPoly *DtPoly
			this.m_nav.GetTileAndPolyByRefUnsafe(neighbourRef, &neighbourTile, &neighbourPoly)

			// Do not advance if the polygon is excluded by the filter.
			if !filter.PassFilter(neighbourRef, neighbourTile, neighbourPoly) {
				continue
			}
			// Find edge and calc distance to the edge.
			var va, vb [3]float32
			if stat := this.getPortalPoints2(bestRef, bestPoly, bestTile, neighbourRef, neighbourPoly, neighbourTile, va[:], vb[:]); DtStatusFailed(stat) {
				continue
			}
			// If the circle is not touching the next polygon, skip it.
			var tseg float32
			distSqr := DtDistancePtSegSqr2D(centerPos, va[:], vb[:], &tseg)
			if distSqr > radiusSqr {
				continue
			}
			neighbourNode := this.m_nodePool.GetNode(neighbourRef, 0)
			if neighbourNode == nil {
				status |= DT_OUT_OF_NODES
				continue
			}

			if (neighbourNode.GetFlags() & DT_NODE_CLOSED) != 0 {
				continue
			}
			// Cost
			if neighbourNode.GetFlags() == 0 {
				DtVlerp(neighbourNode.Pos[:], va[:], vb[:], 0.5)
			}
			total := bestNode.Total + DtVdist(bestNode.Pos[:], neighbourNode.Pos[:])

			// The node is already in open list and the new result is worse, skip.
			if (neighbourNode.GetFlags()&DT_NODE_OPEN) != 0 && total >= neighbourNode.Total {
				continue
			}
			neighbourNode.Id = neighbourRef
			neighbourNode.SetFlags(neighbourNode.GetFlags() & ^DT_NODE_CLOSED)
			neighbourNode.SetPidx(this.m_nodePool.GetNodeIdx(bestNode))
			neighbourNode.Total = total

			if (neighbourNode.GetFlags() & DT_NODE_OPEN) != 0 {
				this.m_openList.Modify(neighbourNode)
			} else {
				neighbourNode.SetFlags(DT_NODE_OPEN)
				this.m_openList.Push(neighbourNode)
			}
		}
	}

	if randomPoly == nil {
		return DT_FAILURE
	}
	// Randomly pick point on polygon.
	v := randomTile.Verts[randomPoly.Verts[0]*3:]
	var verts [3 * DT_VERTS_PER_POLYGON]float32
	var areas [DT_VERTS_PER_POLYGON]float32
	DtVcopy(verts[0*3:], v[:])
	for j := 1; j < int(randomPoly.VertCount); j++ {
		v = randomTile.Verts[randomPoly.Verts[j]*3:]
		DtVcopy(verts[j*3:], v[:])
	}

	s := frand()
	t := frand()

	var pt [3]float32
	DtRandomPointInConvexPoly(verts[:], int(randomPoly.VertCount), areas[:], s, t, pt[:])

	var h float32
	stat := this.GetPolyHeight(randomPolyRef, pt[:], &h)
	if DtStatusFailed(status) {
		return stat
	}
	pt[1] = h

	DtVcopy(randomPt, pt[:])
	*randomRef = randomPolyRef

	return DT_SUCCESS
}

//////////////////////////////////////////////////////////////////////////////////////////

/// Finds the closest point on the specified polygon.
///  @param[in]		ref			The reference id of the polygon.
///  @param[in]		pos			The position to check. [(x, y, z)]
///  @param[out]	closest		The closest point on the polygon. [(x, y, z)]
///  @param[out]	posOverPoly	True of the position is over the polygon.
/// @returns The status flags for the query.
func (this *DtNavMeshQuery) ClosestPointOnPoly(ref DtPolyRef, pos, closest []float32, posOverPoly *bool) DtStatus {
	/// @par
	///
	/// Uses the detail polygons to find the surface height. (Most accurate.)
	///
	/// @p pos does not have to be within the bounds of the polygon or navigation mesh.
	///
	/// See closestPointOnPolyBoundary() for a limited but faster option.
	///
	DtAssert(this.m_nav != nil)
	var tile *DtMeshTile
	var poly *DtPoly
	if DtStatusFailed(this.m_nav.GetTileAndPolyByRef(ref, &tile, &poly)) {
		return DT_FAILURE | DT_INVALID_PARAM
	}
	if tile == nil {
		return DT_FAILURE | DT_INVALID_PARAM
	}
	// Off-mesh connections don't have detail polygons.
	if poly.GetType() == DT_POLYTYPE_OFFMESH_CONNECTION {
		v0 := tile.Verts[poly.Verts[0]*3:]
		v1 := tile.Verts[poly.Verts[1]*3:]
		d0 := DtVdist(pos, v0)
		d1 := DtVdist(pos, v1)
		u := d0 / (d0 + d1)
		DtVlerp(closest, v0, v1, u)
		if posOverPoly != nil {
			*posOverPoly = false
		}
		return DT_SUCCESS
	}

	polyBase := uintptr(unsafe.Pointer(&(tile.Polys[0])))
	current := uintptr(unsafe.Pointer(poly))
	ip := (uint32)(current - polyBase)
	pd := &tile.DetailMeshes[ip]

	// Clamp point to be inside the polygon.
	var verts [DT_VERTS_PER_POLYGON * 3]float32
	var edged [DT_VERTS_PER_POLYGON]float32
	var edget [DT_VERTS_PER_POLYGON]float32
	nv := int(poly.VertCount)
	for i := 0; i < nv; i++ {
		DtVcopy(verts[i*3:], tile.Verts[poly.Verts[i]*3:])
	}
	DtVcopy(closest, pos)
	if !DtDistancePtPolyEdgesSqr(pos, verts[:], nv, edged[:], edget[:]) {
		// Point is outside the polygon, dtClamp to nearest edge.
		dmin := edged[0]
		imin := 0
		for i := 1; i < nv; i++ {
			if edged[i] < dmin {
				dmin = edged[i]
				imin = i
			}
		}
		va := verts[imin*3:]
		vb := verts[((imin+1)%nv)*3:]
		DtVlerp(closest, va, vb, edget[imin])

		if posOverPoly != nil {
			*posOverPoly = false
		}
	} else {
		if posOverPoly != nil {
			*posOverPoly = true
		}
	}

	// Find height at the location.
	for j := 0; j < int(pd.TriCount); j++ {
		t := tile.DetailTris[(int(pd.TriBase)+j)*4:]
		var v [3][]float32
		for k := 0; k < 3; k++ {
			if t[k] < poly.VertCount {
				v[k] = tile.Verts[poly.Verts[t[k]]*3:]
			} else {
				v[k] = tile.DetailVerts[(pd.VertBase+uint32(t[k]-poly.VertCount))*3:]
			}
		}
		var h float32
		if DtClosestHeightPointTriangle(closest, v[0], v[1], v[2], &h) {
			closest[1] = h
			break
		}
	}

	return DT_SUCCESS
}

/// Returns a point on the boundary closest to the source point if the source point is outside the
/// polygon's xz-bounds.
///  @param[in]		ref			The reference id to the polygon.
///  @param[in]		pos			The position to check. [(x, y, z)]
///  @param[out]	closest		The closest point. [(x, y, z)]
/// @returns The status flags for the query.
func (this *DtNavMeshQuery) ClosestPointOnPolyBoundary(ref DtPolyRef, pos, closest []float32) DtStatus {
	/// @par
	///
	/// Much faster than closestPointOnPoly().
	///
	/// If the provided position lies within the polygon's xz-bounds (above or below),
	/// then @p pos and @p closest will be equal.
	///
	/// The height of @p closest will be the polygon boundary.  The height detail is not used.
	///
	/// @p pos does not have to be within the bounds of the polybon or the navigation mesh.
	///
	DtAssert(this.m_nav != nil)
	var tile *DtMeshTile
	var poly *DtPoly
	if DtStatusFailed(this.m_nav.GetTileAndPolyByRef(ref, &tile, &poly)) {
		return DT_FAILURE | DT_INVALID_PARAM
	}

	// Collect vertices.
	var verts [DT_VERTS_PER_POLYGON * 3]float32
	var edged [DT_VERTS_PER_POLYGON]float32
	var edget [DT_VERTS_PER_POLYGON]float32
	nv := 0
	for i := 0; i < int(poly.VertCount); i++ {
		DtVcopy(verts[nv*3:], tile.Verts[poly.Verts[i]*3:])
		nv++
	}

	inside := DtDistancePtPolyEdgesSqr(pos, verts[:], nv, edged[:], edget[:])
	if inside {
		// Point is inside the polygon, return the point.
		DtVcopy(closest, pos)
	} else {
		// Point is outside the polygon, dtClamp to nearest edge.
		dmin := edged[0]
		imin := 0
		for i := 1; i < nv; i++ {
			if edged[i] < dmin {
				dmin = edged[i]
				imin = i
			}
		}
		va := verts[imin*3:]
		vb := verts[((imin+1)%nv)*3:]
		DtVlerp(closest, va, vb, edget[imin])
	}

	return DT_SUCCESS
}

/// Gets the height of the polygon at the provided position using the height detail. (Most accurate.)
///  @param[in]		ref			The reference id of the polygon.
///  @param[in]		pos			A position within the xz-bounds of the polygon. [(x, y, z)]
///  @param[out]	height		The height at the surface of the polygon.
/// @returns The status flags for the query.
func (this *DtNavMeshQuery) GetPolyHeight(ref DtPolyRef, pos []float32, height *float32) DtStatus {
	/// @par
	///
	/// Will return #DT_FAILURE if the provided position is outside the xz-bounds
	/// of the polygon.
	///
	DtAssert(this.m_nav != nil)

	var tile *DtMeshTile
	var poly *DtPoly
	if DtStatusFailed(this.m_nav.GetTileAndPolyByRef(ref, &tile, &poly)) {
		return DT_FAILURE | DT_INVALID_PARAM
	}
	if poly.GetType() == DT_POLYTYPE_OFFMESH_CONNECTION {
		v0 := tile.Verts[poly.Verts[0]*3:]
		v1 := tile.Verts[poly.Verts[1]*3:]
		d0 := DtVdist2D(pos, v0)
		d1 := DtVdist2D(pos, v1)
		u := d0 / (d0 + d1)
		if height != nil {
			*height = v0[1] + (v1[1]-v0[1])*u
		}
		return DT_SUCCESS
	} else {
		polyBase := uintptr(unsafe.Pointer(&(tile.Polys[0])))
		current := uintptr(unsafe.Pointer(poly))
		ip := (uint32)(current - polyBase)
		pd := &tile.DetailMeshes[ip]
		for j := 0; j < int(pd.TriCount); j++ {
			t := tile.DetailTris[(int(pd.TriBase)+j)*4:]
			var v [3][]float32
			for k := 0; k < 3; k++ {
				if t[k] < poly.VertCount {
					v[k] = tile.Verts[poly.Verts[t[k]]*3:]
				} else {
					v[k] = tile.DetailVerts[(pd.VertBase+uint32(t[k]-poly.VertCount))*3:]
				}
			}
			var h float32
			if DtClosestHeightPointTriangle(pos, v[0], v[1], v[2], &h) {
				if height != nil {
					*height = h
				}
				return DT_SUCCESS
			}
		}
	}

	return DT_FAILURE | DT_INVALID_PARAM
}

type dtFindNearestPolyQuery struct {
	m_query              *DtNavMeshQuery
	m_center             []float32
	m_nearestDistanceSqr float32
	m_nearestRef         DtPolyRef
	m_nearestPoint       [3]float32
}

func (this *dtFindNearestPolyQuery) constructor(query *DtNavMeshQuery, center []float32) {
	this.m_query = query
	this.m_center = center
	this.m_nearestDistanceSqr = float32(math.MaxFloat32)
}

func (this *dtFindNearestPolyQuery) nearestRef() DtPolyRef   { return this.m_nearestRef }
func (this *dtFindNearestPolyQuery) nearestPoint() []float32 { return this.m_nearestPoint[:] }

func (this *dtFindNearestPolyQuery) Process(tile *DtMeshTile, polys []*DtPoly, refs []DtPolyRef, count int) {
	//DtIgnoreUnused(polys);
	for i := 0; i < count; i++ {
		ref := refs[i]
		var closestPtPoly [3]float32
		var diff [3]float32
		posOverPoly := false
		var d float32
		this.m_query.ClosestPointOnPoly(ref, this.m_center, closestPtPoly[:], &posOverPoly)

		// If a point is directly over a polygon and closer than
		// climb height, favor that instead of straight line nearest point.
		DtVsub(diff[:], this.m_center, closestPtPoly[:])
		if posOverPoly {
			d = DtAbsFloat32(diff[1]) - tile.Header.WalkableClimb
			if d > 0 {
				d = d * d
			} else {
				d = 0
			}
		} else {
			d = DtVlenSqr(diff[:])
		}

		if d < this.m_nearestDistanceSqr {
			DtVcopy(this.m_nearestPoint[:], closestPtPoly[:])

			this.m_nearestDistanceSqr = d
			this.m_nearestRef = ref
		}
	}
}

/// Finds the polygon nearest to the specified center point.
///  @param[in]		center		The center of the search box. [(x, y, z)]
///  @param[in]		halfExtents		The search distance along each axis. [(x, y, z)]
///  @param[in]		filter		The polygon filter to apply to the query.
///  @param[out]	nearestRef	The reference id of the nearest polygon.
///  @param[out]	nearestPt	The nearest point on the polygon. [opt] [(x, y, z)]
/// @returns The status flags for the query.
func (this *DtNavMeshQuery) FindNearestPoly(center, halfExtents []float32,
	filter *DtQueryFilter,
	nearestRef *DtPolyRef, nearestPt []float32) DtStatus {
	/// @par
	///
	/// @note If the search box does not intersect any polygons the search will
	/// return #DT_SUCCESS, but @p nearestRef will be zero. So if in doubt, check
	/// @p nearestRef before using @p nearestPt.
	///
	DtAssert(this.m_nav != nil)

	if nearestRef == nil {
		return DT_FAILURE | DT_INVALID_PARAM
	}
	query := dtFindNearestPolyQuery{}
	query.constructor(this, center)

	status := this.QueryPolygons2(center, halfExtents, filter, &query)
	if DtStatusFailed(status) {
		return status
	}
	*nearestRef = query.nearestRef()
	// Only override nearestPt if we actually found a poly so the nearest point
	// is valid.
	if nearestPt != nil && (*nearestRef) != 0 {
		DtVcopy(nearestPt, query.nearestPoint())
	}
	return DT_SUCCESS
}

/// Queries polygons within a tile.
func (this *DtNavMeshQuery) queryPolygonsInTile(tile *DtMeshTile, qmin, qmax []float32,
	filter *DtQueryFilter, query dtPolyQuery) {
	DtAssert(this.m_nav != nil)
	const batchSize int = 32
	var polyRefs [batchSize]DtPolyRef
	var polys [batchSize]*DtPoly
	n := 0

	if tile.BvTree != nil {
		nodeIndex := 0
		endIndex := int(tile.Header.BvNodeCount)
		tbmin := tile.Header.Bmin[:]
		tbmax := tile.Header.Bmax[:]
		qfac := tile.Header.BvQuantFactor

		// Calculate quantized box
		var bmin, bmax [3]uint16
		// dtClamp query box to world box.
		minx := DtClampFloat32(qmin[0], tbmin[0], tbmax[0]) - tbmin[0]
		miny := DtClampFloat32(qmin[1], tbmin[1], tbmax[1]) - tbmin[1]
		minz := DtClampFloat32(qmin[2], tbmin[2], tbmax[2]) - tbmin[2]
		maxx := DtClampFloat32(qmax[0], tbmin[0], tbmax[0]) - tbmin[0]
		maxy := DtClampFloat32(qmax[1], tbmin[1], tbmax[1]) - tbmin[1]
		maxz := DtClampFloat32(qmax[2], tbmin[2], tbmax[2]) - tbmin[2]
		// Quantize
		bmin[0] = (uint16)(qfac*minx) & 0xfffe
		bmin[1] = (uint16)(qfac*miny) & 0xfffe
		bmin[2] = (uint16)(qfac*minz) & 0xfffe
		bmax[0] = (uint16)(qfac*maxx+1) | 1
		bmax[1] = (uint16)(qfac*maxy+1) | 1
		bmax[2] = (uint16)(qfac*maxz+1) | 1

		// Traverse tree
		base := this.m_nav.GetPolyRefBase(tile)
		for nodeIndex < endIndex {
			node := &tile.BvTree[nodeIndex]
			overlap := DtOverlapQuantBounds(bmin[:], bmax[:], node.Bmin[:], node.Bmax[:])
			isLeafNode := (node.I >= 0)

			if isLeafNode && overlap {
				ref := base | (DtPolyRef)(node.I)
				if filter.PassFilter(ref, tile, &tile.Polys[node.I]) {
					polyRefs[n] = ref
					polys[n] = &tile.Polys[node.I]

					if n == batchSize-1 {
						query.Process(tile, polys[:], polyRefs[:], batchSize)
						n = 0
					} else {
						n++
					}
				}
			}

			if overlap || isLeafNode {
				nodeIndex++
			} else {
				escapeIndex := int(-node.I)
				nodeIndex += escapeIndex
			}
		}
	} else {
		var bmin, bmax [3]float32
		base := this.m_nav.GetPolyRefBase(tile)
		for i := 0; i < int(tile.Header.PolyCount); i++ {
			p := &tile.Polys[i]
			// Do not return off-mesh connection polygons.
			if p.GetType() == DT_POLYTYPE_OFFMESH_CONNECTION {
				continue
			}
			// Must pass filter
			ref := base | (DtPolyRef)(i)
			if !filter.PassFilter(ref, tile, p) {
				continue
			}
			// Calc polygon bounds.
			v := tile.Verts[p.Verts[0]*3:]
			DtVcopy(bmin[:], v)
			DtVcopy(bmax[:], v)
			for j := 1; j < int(p.VertCount); j++ {
				v = tile.Verts[p.Verts[j]*3:]
				DtVmin(bmin[:], v)
				DtVmax(bmax[:], v)
			}
			if DtOverlapBounds(qmin, qmax, bmin[:], bmax[:]) {
				polyRefs[n] = ref
				polys[n] = p

				if n == batchSize-1 {
					query.Process(tile, polys[:], polyRefs[:], batchSize)
					n = 0
				} else {
					n++
				}
			}
		}
	}

	// Process the last polygons that didn't make a full batch.
	if n > 0 {
		query.Process(tile, polys[:], polyRefs[:], n)
	}
}

type dtCollectPolysQuery struct {
	m_polys        []DtPolyRef
	m_maxPolys     int
	m_numCollected int
	m_overflow     bool
}

func (this *dtCollectPolysQuery) constructor(polys []DtPolyRef, maxPolys int) {
	this.m_polys = polys
	this.m_maxPolys = maxPolys
}

func (this *dtCollectPolysQuery) numCollected() int { return this.m_numCollected }
func (this *dtCollectPolysQuery) overflowed() bool  { return this.m_overflow }

func (this *dtCollectPolysQuery) Process(tile *DtMeshTile, polys []*DtPoly, refs []DtPolyRef, count int) {
	//dtIgnoreUnused(tile);
	//dtIgnoreUnused(polys);
	numLeft := this.m_maxPolys - this.m_numCollected
	toCopy := count
	if toCopy > numLeft {
		this.m_overflow = true
		toCopy = numLeft
	}
	copy(this.m_polys[this.m_numCollected:], refs[0:toCopy])
	this.m_numCollected += toCopy
}

/// Finds polygons that overlap the search box.
///  @param[in]		center		The center of the search box. [(x, y, z)]
///  @param[in]		halfExtents		The search distance along each axis. [(x, y, z)]
///  @param[in]		filter		The polygon filter to apply to the query.
///  @param[out]	polys		The reference ids of the polygons that overlap the query box.
///  @param[out]	polyCount	The number of polygons in the search result.
///  @param[in]		maxPolys	The maximum number of polygons the search result can hold.
/// @returns The status flags for the query.
func (this *DtNavMeshQuery) QueryPolygons(center, halfExtents []float32,
	filter *DtQueryFilter,
	polys []DtPolyRef, polyCount *int, maxPolys int) DtStatus {
	/// @par
	///
	/// If no polygons are found, the function will return #DT_SUCCESS with a
	/// @p polyCount of zero.
	///
	/// If @p polys is too small to hold the entire result set, then the array will
	/// be filled to capacity. The method of choosing which polygons from the
	/// full set are included in the partial result set is undefined.
	///
	if polys == nil || polyCount == nil || maxPolys < 0 {
		return DT_FAILURE | DT_INVALID_PARAM
	}
	collector := dtCollectPolysQuery{}
	collector.constructor(polys, maxPolys)

	status := this.QueryPolygons2(center, halfExtents, filter, &collector)
	if DtStatusFailed(status) {
		return status
	}
	*polyCount = collector.numCollected()
	if collector.overflowed() {
		return DT_SUCCESS | DT_BUFFER_TOO_SMALL
	} else {
		return DT_SUCCESS
	}
}

/// Finds polygons that overlap the search box.
///  @param[in]		center		The center of the search box. [(x, y, z)]
///  @param[in]		halfExtents		The search distance along each axis. [(x, y, z)]
///  @param[in]		filter		The polygon filter to apply to the query.
///  @param[in]		query		The query. Polygons found will be batched together and passed to this query.
func (this *DtNavMeshQuery) QueryPolygons2(center, halfExtents []float32,
	filter *DtQueryFilter, query dtPolyQuery) DtStatus {
	/// @par
	///
	/// The query will be invoked with batches of polygons. Polygons passed
	/// to the query have bounding boxes that overlap with the center and halfExtents
	/// passed to this function. The dtPolyQuery::process function is invoked multiple
	/// times until all overlapping polygons have been processed.
	///
	DtAssert(this.m_nav != nil)

	if center == nil || halfExtents == nil || filter == nil || query == nil {
		return DT_FAILURE | DT_INVALID_PARAM
	}
	var bmin, bmax [3]float32
	DtVsub(bmin[:], center, halfExtents)
	DtVadd(bmax[:], center, halfExtents)

	// Find tiles the query touches.
	var minx, miny, maxx, maxy int32
	this.m_nav.CalcTileLoc(bmin[:], &minx, &miny)
	this.m_nav.CalcTileLoc(bmax[:], &maxx, &maxy)

	const MAX_NEIS int = 32
	var neis [MAX_NEIS]*DtMeshTile

	for y := miny; y <= maxy; y++ {
		for x := minx; x <= maxx; x++ {
			nneis := this.m_nav.GetTilesAt(x, y, neis[:], MAX_NEIS)
			for j := 0; j < nneis; j++ {
				this.queryPolygonsInTile(neis[j], bmin[:], bmax[:], filter, query)
			}
		}
	}

	return DT_SUCCESS
}

// Returns portal points between two polygons.
func (this *DtNavMeshQuery) getPortalPoints2(from DtPolyRef, fromPoly *DtPoly, fromTile *DtMeshTile,
	to DtPolyRef, toPoly *DtPoly, toTile *DtMeshTile,
	left, right []float32) DtStatus {
	// Find the link that points to the 'to' polygon.
	var link *DtLink
	for i := fromPoly.FirstLink; i != DT_NULL_LINK; i = fromTile.Links[i].Next {
		if fromTile.Links[i].Ref == to {
			link = &fromTile.Links[i]
			break
		}
	}
	if link == nil {
		return DT_FAILURE | DT_INVALID_PARAM
	}
	// Handle off-mesh connections.
	if fromPoly.GetType() == DT_POLYTYPE_OFFMESH_CONNECTION {
		// Find link that points to first vertex.
		for i := fromPoly.FirstLink; i != DT_NULL_LINK; i = fromTile.Links[i].Next {
			if fromTile.Links[i].Ref == to {
				v := fromTile.Links[i].Edge
				DtVcopy(left, fromTile.Verts[fromPoly.Verts[v]*3:])
				DtVcopy(right, fromTile.Verts[fromPoly.Verts[v]*3:])
				return DT_SUCCESS
			}
		}
		return DT_FAILURE | DT_INVALID_PARAM
	}

	if toPoly.GetType() == DT_POLYTYPE_OFFMESH_CONNECTION {
		for i := toPoly.FirstLink; i != DT_NULL_LINK; i = toTile.Links[i].Next {
			if toTile.Links[i].Ref == from {
				v := toTile.Links[i].Edge
				DtVcopy(left, toTile.Verts[toPoly.Verts[v]*3:])
				DtVcopy(right, toTile.Verts[toPoly.Verts[v]*3:])
				return DT_SUCCESS
			}
		}
		return DT_FAILURE | DT_INVALID_PARAM
	}

	// Find portal vertices.
	v0 := fromPoly.Verts[link.Edge]
	v1 := fromPoly.Verts[int(link.Edge+1)%(int)(fromPoly.VertCount)]
	DtVcopy(left, fromTile.Verts[v0*3:])
	DtVcopy(right, fromTile.Verts[v1*3:])

	// If the link is at tile boundary, dtClamp the vertices to
	// the link width.
	if link.Side != 0xff {
		// Unpack portal limits.
		if link.Bmin != 0 || link.Bmax != 255 {
			s := float32(1.0 / 255.0)
			tmin := float32(link.Bmin) * s
			tmax := float32(link.Bmax) * s
			DtVlerp(left, fromTile.Verts[v0*3:], fromTile.Verts[v1*3:], tmin)
			DtVlerp(right, fromTile.Verts[v0*3:], fromTile.Verts[v1*3:], tmax)
		}
	}

	return DT_SUCCESS
}
