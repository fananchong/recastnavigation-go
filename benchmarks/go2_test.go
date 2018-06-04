package benchmarks

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/fananchong/recastnavigation-go/Detour"
	"github.com/fananchong/recastnavigation-go/tests"
)

func Benchmark_TileCache_FindPath(t *testing.B) {
	var randPosValue []float32
	var randPosIndex int = 0

	getPos := func(ref *detour.DtPolyRef, pos []float32) {
		*ref = detour.DtPolyRef(randPosValue[randPosIndex*4+0])
		pos[0] = randPosValue[randPosIndex*4+1]
		pos[1] = randPosValue[randPosIndex*4+2]
		pos[2] = randPosValue[randPosIndex*4+3]
		randPosIndex++
	}

	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&randPosValue)))
	sliceHeader.Cap = int(len(tempdata2) / int(unsafe.Sizeof(float32(1.0))))
	sliceHeader.Len = int(len(tempdata2) / int(unsafe.Sizeof(float32(1.0))))
	sliceHeader.Data = uintptr(unsafe.Pointer(&(tempdata2[0])))

	query := tests.CreateQuery(mesh2, 2048)
	filter := detour.DtAllocDtQueryFilter()

	for i := 0; i < t.N; i++ {
		var stat detour.DtStatus
		startPos := [3]float32{0, 0, 0}
		endPos := [3]float32{0, 0, 0}
		var startRef detour.DtPolyRef
		var endRef detour.DtPolyRef
		getPos(&startRef, startPos[:])
		getPos(&endRef, endPos[:])
		var path [PATH_MAX_NODE]detour.DtPolyRef
		pathCount := 0
		stat = query.FindPath(startRef, endRef, startPos[:], endPos[:], filter, path[:], &pathCount, PATH_MAX_NODE)
		detour.DtAssert(detour.DtStatusSucceed(stat))
	}
}

func Benchmark_TileCache_MoveAlongSurface(t *testing.B) {
	var randPosValue []float32
	var randPosIndex int = 0

	getPos := func(ref *detour.DtPolyRef, pos []float32) {
		*ref = detour.DtPolyRef(randPosValue[randPosIndex*4+0])
		pos[0] = randPosValue[randPosIndex*4+1]
		pos[1] = randPosValue[randPosIndex*4+2]
		pos[2] = randPosValue[randPosIndex*4+3]
		randPosIndex++
	}

	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&randPosValue)))
	sliceHeader.Cap = int(len(tempdata2) / int(unsafe.Sizeof(float32(1.0))))
	sliceHeader.Len = int(len(tempdata2) / int(unsafe.Sizeof(float32(1.0))))
	sliceHeader.Data = uintptr(unsafe.Pointer(&(tempdata2[0])))

	query := tests.CreateQuery(mesh2, 2048)
	filter := detour.DtAllocDtQueryFilter()

	for i := 0; i < t.N; i++ {
		var stat detour.DtStatus
		startPos := [3]float32{0, 0, 0}
		endPos := [3]float32{0, 0, 0}
		var startRef detour.DtPolyRef
		var endRef detour.DtPolyRef
		getPos(&startRef, startPos[:])
		getPos(&endRef, endPos[:])
		resultPos := [3]float32{0, 0, 0}
		var visited [PATH_MAX_NODE]detour.DtPolyRef
		visitedCount := 0
		bHit := false
		stat = query.MoveAlongSurface(startRef, startPos[:], endPos[:], filter, resultPos[:], visited[:], &visitedCount, PATH_MAX_NODE, &bHit)
		detour.DtAssert(detour.DtStatusSucceed(stat))
	}
}

func Benchmark_TileCache_FindNearestPoly(t *testing.B) {
	var randPosValue []float32
	var randPosIndex int = 0

	getPos := func(ref *detour.DtPolyRef, pos []float32) {
		*ref = detour.DtPolyRef(randPosValue[randPosIndex*4+0])
		pos[0] = randPosValue[randPosIndex*4+1]
		pos[1] = randPosValue[randPosIndex*4+2]
		pos[2] = randPosValue[randPosIndex*4+3]
		randPosIndex++
	}

	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&randPosValue)))
	sliceHeader.Cap = int(len(tempdata2) / int(unsafe.Sizeof(float32(1.0))))
	sliceHeader.Len = int(len(tempdata2) / int(unsafe.Sizeof(float32(1.0))))
	sliceHeader.Data = uintptr(unsafe.Pointer(&(tempdata2[0])))

	query := tests.CreateQuery(mesh2, 2048)
	filter := detour.DtAllocDtQueryFilter()

	for i := 0; i < t.N; i++ {
		var stat detour.DtStatus
		halfExtents := [3]float32{0.6, 2.0, 0.6}
		startPos := [3]float32{0, 0, 0}
		var startRef detour.DtPolyRef
		getPos(&startRef, startPos[:])
		var nearestRef detour.DtPolyRef
		nearestPos := [3]float32{0, 0, 0}
		stat = query.FindNearestPoly(startPos[:], halfExtents[:], filter, &nearestRef, nearestPos[:])
		detour.DtAssert(detour.DtStatusSucceed(stat))
	}
}
