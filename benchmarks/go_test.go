package benchmarks

import (
	"io/ioutil"
	"reflect"
	"testing"
	"unsafe"

	"github.com/fananchong/recastnavigation-go/Detour"
	"github.com/fananchong/recastnavigation-go/tests"
)

const RAND_MAX_COUNT int = 20000000
const PATH_MAX_NODE int = 2048

var tempdata, err = ioutil.ReadFile("../tests/randpos.bin")
var mesh = tests.LoadStaticMesh("../tests/nav_test.obj.tile.bin")
var randPosValue []float32

func Benchmark_GO_FindPath(t *testing.B) {
	var randPosIndex int = 0

	getPos := func(ref *detour.DtPolyRef, pos []float32) {
		*ref = detour.DtPolyRef(randPosValue[randPosIndex*4+0])
		pos[0] = randPosValue[randPosIndex*4+1]
		pos[1] = randPosValue[randPosIndex*4+2]
		pos[2] = randPosValue[randPosIndex*4+3]
		randPosIndex++
	}

	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&randPosValue)))
	sliceHeader.Cap = int(len(tempdata) / int(unsafe.Sizeof(float32(1.0))))
	sliceHeader.Len = int(len(tempdata) / int(unsafe.Sizeof(float32(1.0))))
	sliceHeader.Data = uintptr(unsafe.Pointer(&(tempdata[0])))

	query := tests.CreateQuery(mesh, 2048)
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

func Benchmark_GO_MoveAlongSurface(t *testing.B) {
	var randPosValue [RAND_MAX_COUNT * 4]float32
	var randPosIndex int = 0

	getPos := func(ref *detour.DtPolyRef, pos []float32) {
		*ref = detour.DtPolyRef(randPosValue[randPosIndex*4+0])
		pos[0] = randPosValue[randPosIndex*4+1]
		pos[1] = randPosValue[randPosIndex*4+2]
		pos[2] = randPosValue[randPosIndex*4+3]
		randPosIndex++
	}

	detour.DtAssert(err == nil)
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&randPosValue)))
	sliceHeader.Cap = int(len(tempdata) / int(unsafe.Sizeof(float32(1.0))))
	sliceHeader.Len = int(len(tempdata) / int(unsafe.Sizeof(float32(1.0))))
	sliceHeader.Data = uintptr(unsafe.Pointer(&(tempdata[0])))

	query := tests.CreateQuery(mesh, 2048)
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
