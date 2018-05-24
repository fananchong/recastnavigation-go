package detour

import "unsafe"

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
