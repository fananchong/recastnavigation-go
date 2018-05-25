package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"unsafe"

	detour "github.com/fananchong/recastnavigation-go/Detour"
)

const (
	//copy from sample.h
	SAMPLE_POLYAREA_GROUND int32 = 0
	SAMPLE_POLYAREA_WATER  int32 = 1
	SAMPLE_POLYAREA_ROAD   int32 = 2
	SAMPLE_POLYAREA_DOOR   int32 = 3
	SAMPLE_POLYAREA_GRASS  int32 = 4
	SAMPLE_POLYAREA_JUMP   int32 = 5

	SAMPLE_POLYFLAGS_WALK     uint16 = 0x01   // Ability to walk (ground, grass, road)
	SAMPLE_POLYFLAGS_SWIM     uint16 = 0x02   // Ability to swim (water).
	SAMPLE_POLYFLAGS_DOOR     uint16 = 0x04   // Ability to move through doors.
	SAMPLE_POLYFLAGS_JUMP     uint16 = 0x08   // Ability to jump.
	SAMPLE_POLYFLAGS_DISABLED uint16 = 0x10   // Disabled polygon
	SAMPLE_POLYFLAGS_ALL      uint16 = 0xffff // All abilities.
)

var (
	navMesh     *detour.DtNavMesh
	navQuery    *detour.DtNavMeshQuery
	filter      *detour.DtQueryFilter
	polyPickExt [3]float32
)

func init() {
	navMesh = detour.DtAllocNavMesh()
	navQuery = detour.DtAllocNavMeshQuery()
	filter = detour.DtAllocDtQueryFilter()
	filter.SetIncludeFlags(SAMPLE_POLYFLAGS_ALL ^ SAMPLE_POLYFLAGS_DISABLED)
	filter.SetExcludeFlags(0)

	polyPickExt[0] = 2
	polyPickExt[1] = 4
	polyPickExt[2] = 2

	f, err := os.Open("navmesh.data")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	params := (*detour.DtNavMeshParams)(unsafe.Pointer(&buf[0]))
	if detour.DtStatusFailed(navMesh.Init(params)) {
		panic("buildTiledNavigation: Could not init navmesh.")
	}

	if detour.DtStatusFailed(navQuery.Init(navMesh, 2048)) {
		panic("buildTiledNavigation: Could not init Detour navmesh query.")
	}

	paramSize := detour.DtAlign4(int(unsafe.Sizeof(detour.DtNavMeshParams{})))
	buf = buf[paramSize:]
	var x, y, dataSize *int32
	var count int32
	for {
		if len(buf) < 4 {
			break
		}
		x = (*int32)(unsafe.Pointer(&buf[0]))
		buf = buf[4:]

		if len(buf) < 4 {
			panic("init 1")
		}
		y = (*int32)(unsafe.Pointer(&buf[0]))
		buf = buf[4:]

		if len(buf) < 4 {
			panic("init 2")
		}
		dataSize = (*int32)(unsafe.Pointer(&buf[0]))
		buf = buf[4:]
		if len(buf) < int(*dataSize) {
			panic("init 3")
		}
		data := buf[:*dataSize]
		buf = buf[*dataSize:]

		if *dataSize > 0 {
			navMesh.RemoveTile(navMesh.GetTileRefAt(*x, *y, 0), nil, nil)
			status := navMesh.AddTile(data, int(*dataSize), detour.DT_TILE_FREE_DATA, 0, nil)
			if !detour.DtStatusFailed(status) {
				count++
			}
		}
	}
	fmt.Println("success count:", count)
}

func GoFindPath(start, end, ptlst []float32, ptCount *int, maxPolys int) {
	var startRef detour.DtPolyRef
	var endRef detour.DtPolyRef
	status := navQuery.FindNearestPoly(start, polyPickExt[:], filter, &startRef, nil)
	if detour.DtStatusFailed(status) {
		fmt.Println("startref falied")
		return
	}
	status = navQuery.FindNearestPoly(end, polyPickExt[:], filter, &endRef, nil)
	if detour.DtStatusFailed(status) {
		fmt.Println("startref falied")
		return
	}

	fmt.Println("startRef:", startRef, " endRef:", endRef)

	polys := make([]detour.DtPolyRef, maxPolys)
	var npolys int
	navQuery.FindPath(startRef, endRef, start, end, filter, polys, &npolys, maxPolys)
	if npolys > 0 {
		fmt.Println("findPath npolys:", npolys)
		for i := 0; i < npolys; i++ {
			fmt.Println(polys[i])
		}
		fmt.Println()

		var epos [3]float32
		detour.DtVcopy(epos[:], end)
		if polys[npolys-1] != endRef {
			navQuery.ClosestPointOnPoly(polys[npolys-1], end, epos[:], nil)
		}

		straightPathFlags := make([]detour.DtStraightPathFlags, maxPolys)
		straightPathPolys := make([]detour.DtPolyRef, maxPolys)
		navQuery.FindStraightPath(start, epos[:], polys, npolys, ptlst, straightPathFlags,
			straightPathPolys, ptCount, maxPolys, 0)
	} else {
		fmt.Println("find path failed")
	}
}
