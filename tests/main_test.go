package tests

import (
	"io/ioutil"
	"reflect"
	"testing"
	"unsafe"

	"github.com/fananchong/recastnavigation-go/Detour"
)

var randValue []float32
var randIndex = 0

func frand() float32 {
	v := randValue[randIndex]
	randIndex++
	return v
}

const PATH_MAX_NODE int = 2048

func Test_main(t *testing.T) {
	tempdata, err := ioutil.ReadFile("rand.bin")
	detour.DtAssert(err == nil)
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&randValue)))
	sliceHeader.Cap = int(len(tempdata) / int(unsafe.Sizeof(float32(1.0))))
	sliceHeader.Len = int(len(tempdata) / int(unsafe.Sizeof(float32(1.0))))
	sliceHeader.Data = uintptr(unsafe.Pointer(&(tempdata[0])))

	var resultValue []float32
	var resultIndex = 0
	tempdata2, err2 := ioutil.ReadFile("result.bin")
	detour.DtAssert(err2 == nil)
	sliceHeader = (*reflect.SliceHeader)((unsafe.Pointer(&resultValue)))
	sliceHeader.Cap = int(len(tempdata2) / int(unsafe.Sizeof(float32(1.0))))
	sliceHeader.Len = int(len(tempdata2) / int(unsafe.Sizeof(float32(1.0))))
	sliceHeader.Data = uintptr(unsafe.Pointer(&(tempdata2[0])))

	mesh := LoadStaticMesh("nav_test.obj.tile.bin")
	query := CreateQuery(mesh, 2048)
	filter := detour.DtAllocDtQueryFilter()

	var stat detour.DtStatus
	halfExtents := [3]float32{2, 4, 2}
	var startPos, endPos [3]float32
	var startRef, endRef detour.DtPolyRef

	t.Logf("================================================ findRandomPoint ================================================\n")
	stat = query.FindRandomPoint(filter, frand, &startRef, startPos[:])
	detour.DtAssert(detour.DtStatusSucceed(stat))
	stat = query.FindRandomPoint(filter, frand, &endRef, endPos[:])
	detour.DtAssert(detour.DtStatusSucceed(stat))
	t.Logf("startPos: %.2f %.2f %.2f", startPos[0], startPos[1], startPos[2])
	t.Logf("endPos: %.2f %.2f %.2f", endPos[0], endPos[1], endPos[2])
	t.Logf("startRef: %d", startRef)
	t.Logf("endRef: %d", endRef)
	t.Logf("\n")
	detour.DtAssert(IsEquals(resultValue[resultIndex], startPos[0]))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], startPos[1]))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], startPos[2]))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], endPos[0]))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], endPos[1]))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], endPos[2]))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], float32(startRef)))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], float32(endRef)))
	resultIndex++

	t.Logf("================================================ findNearestPoly ================================================\n")
	tempPos := [3]float32{0, 0, 0}
	var nearestPos [3]float32
	var nearestRef detour.DtPolyRef
	stat = query.FindNearestPoly(tempPos[:], halfExtents[:], filter, &nearestRef, nearestPos[:])
	detour.DtAssert(detour.DtStatusSucceed(stat))
	t.Logf("nearestPos: %.2f %.2f %.2f\n", nearestPos[0], nearestPos[1], nearestPos[2])
	t.Logf("nearestRef: %d\n", nearestRef)
	t.Logf("\n")
	detour.DtAssert(IsEquals(resultValue[resultIndex], nearestPos[0]))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], nearestPos[1]))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], nearestPos[2]))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], float32(nearestRef)))
	resultIndex++

	t.Logf("================================================ findPath ================================================\n")
	var path [PATH_MAX_NODE]detour.DtPolyRef
	var pathCount int
	stat = query.FindPath(startRef, endRef, startPos[:], endPos[:], filter, path[:], &pathCount, PATH_MAX_NODE)
	detour.DtAssert(detour.DtStatusSucceed(stat))
	t.Logf("pathCount: %d\n", pathCount)
	detour.DtAssert(IsEquals(resultValue[resultIndex], float32(pathCount)))
	resultIndex++
	for i := 0; i < pathCount; i++ {
		t.Logf("%d\n", path[i])
		detour.DtAssert(IsEquals(resultValue[resultIndex], float32(path[i])))
		resultIndex++
	}
	t.Logf("\n")

	{
		t.Logf("================================================ findStraightPath # DT_STRAIGHTPATH_AREA_CROSSINGS ================================================\n")
		var straightPath [PATH_MAX_NODE * 3]float32
		var straightPathFlags [PATH_MAX_NODE]detour.DtStraightPathFlags
		var straightPathRefs [PATH_MAX_NODE]detour.DtPolyRef
		var straightPathCount int
		stat = query.FindStraightPath(startPos[:], endPos[:], path[:], pathCount,
			straightPath[:], straightPathFlags[:], straightPathRefs[:],
			&straightPathCount, PATH_MAX_NODE, detour.DT_STRAIGHTPATH_AREA_CROSSINGS)
		detour.DtAssert(detour.DtStatusSucceed(stat))
		t.Logf("straightPathCount: %d\n", straightPathCount)
		detour.DtAssert(IsEquals(resultValue[resultIndex], float32(straightPathCount)))
		resultIndex++
		for i := 0; i < straightPathCount; i++ {
			t.Logf("straightPath: %.3f %.3f %.3f, straightPathFlags: %d, straightPathRefs: %d\n",
				straightPath[i*3+0], straightPath[i*3+1], straightPath[i*3+2],
				straightPathFlags[i], straightPathRefs[i])
			detour.DtAssert(IsEquals(resultValue[resultIndex], float32(straightPath[i*3+0])))
			resultIndex++
			detour.DtAssert(IsEquals(resultValue[resultIndex], float32(straightPath[i*3+1])))
			resultIndex++
			detour.DtAssert(IsEquals(resultValue[resultIndex], float32(straightPath[i*3+2])))
			resultIndex++
			detour.DtAssert(IsEquals(resultValue[resultIndex], float32(straightPathFlags[i])))
			resultIndex++
			detour.DtAssert(IsEquals(resultValue[resultIndex], float32(straightPathRefs[i])))
			resultIndex++
		}
		t.Logf("\n")
	}

	{
		t.Logf("================================================ findStraightPath # DT_STRAIGHTPATH_ALL_CROSSINGS ================================================\n")
		var straightPath [PATH_MAX_NODE * 3]float32
		var straightPathFlags [PATH_MAX_NODE]detour.DtStraightPathFlags
		var straightPathRefs [PATH_MAX_NODE]detour.DtPolyRef
		var straightPathCount int
		stat = query.FindStraightPath(startPos[:], endPos[:], path[:], pathCount,
			straightPath[:], straightPathFlags[:], straightPathRefs[:],
			&straightPathCount, PATH_MAX_NODE, detour.DT_STRAIGHTPATH_ALL_CROSSINGS)
		detour.DtAssert(detour.DtStatusSucceed(stat))
		t.Logf("straightPathCount: %d\n", straightPathCount)
		detour.DtAssert(IsEquals(resultValue[resultIndex], float32(straightPathCount)))
		resultIndex++
		for i := 0; i < straightPathCount; i++ {
			t.Logf("straightPath: %.3f %.3f %.3f, straightPathFlags: %d, straightPathRefs: %d\n",
				straightPath[i*3+0], straightPath[i*3+1], straightPath[i*3+2],
				straightPathFlags[i], straightPathRefs[i])
			detour.DtAssert(IsEquals(resultValue[resultIndex], float32(straightPath[i*3+0])))
			resultIndex++
			detour.DtAssert(IsEquals(resultValue[resultIndex], float32(straightPath[i*3+1])))
			resultIndex++
			detour.DtAssert(IsEquals(resultValue[resultIndex], float32(straightPath[i*3+2])))
			resultIndex++
			detour.DtAssert(IsEquals(resultValue[resultIndex], float32(straightPathFlags[i])))
			resultIndex++
			detour.DtAssert(IsEquals(resultValue[resultIndex], float32(straightPathRefs[i])))
			resultIndex++
		}
		t.Logf("\n")
	}

	t.Logf("================================================ moveAlongSurface ================================================\n")
	var resultPos [3]float32
	var visited [PATH_MAX_NODE]detour.DtPolyRef
	var visitedCount int
	bHit := false
	stat = query.MoveAlongSurface(startRef, startPos[:], endPos[:], filter, resultPos[:], visited[:], &visitedCount, PATH_MAX_NODE, &bHit)
	detour.DtAssert(detour.DtStatusSucceed(stat))
	t.Logf("resultPos: %.2f %.2f %.2f\n", resultPos[0], resultPos[1], resultPos[2])
	t.Log("bHit: ", bHit)
	t.Logf("visitedCount: %d\n", visitedCount)
	detour.DtAssert(IsEquals(resultValue[resultIndex], float32(resultPos[0])))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], float32(resultPos[1])))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], float32(resultPos[2])))
	resultIndex++
	if bHit {
		detour.DtAssert(IsEquals(resultValue[resultIndex], float32(1)))
	} else {
		detour.DtAssert(IsEquals(resultValue[resultIndex], float32(0)))
	}
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], float32(visitedCount)))
	resultIndex++
	for i := 0; i < visitedCount; i++ {
		t.Logf("%d\n", visited[i])
		detour.DtAssert(IsEquals(resultValue[resultIndex], float32(visited[i])))
		resultIndex++
	}
	t.Logf("\n")

	t.Logf("================================================ findDistanceToWall ================================================\n")
	var hitDist float32
	var hitPos [3]float32
	var hitNormal [3]float32
	stat = query.FindDistanceToWall(startRef, startPos[:], 30, filter, &hitDist, hitPos[:], hitNormal[:])
	detour.DtAssert(detour.DtStatusSucceed(stat))
	t.Logf("hitPos: %.2f %.2f %.2f\n", hitPos[0], hitPos[1], hitPos[2])
	t.Logf("hitDist: %f\n", hitDist)
	t.Logf("hitNormal: %.2f %.2f %.2f\n", hitNormal[0], hitNormal[1], hitNormal[2])
	t.Logf("\n")
	detour.DtAssert(IsEquals(resultValue[resultIndex], float32(hitPos[0])))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], float32(hitPos[1])))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], float32(hitPos[2])))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], float32(hitDist)))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], float32(hitNormal[0])))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], float32(hitNormal[1])))
	resultIndex++
	detour.DtAssert(IsEquals(resultValue[resultIndex], float32(hitNormal[2])))
	resultIndex++

	{
		t.Logf("================================================ SlicedFindPath # 0================================================\n")
		stat = query.InitSlicedFindPath(startRef, endRef, startPos[:], endPos[:], filter, 0)
		detour.DtAssert(detour.DtStatusInProgress(stat) || detour.DtStatusSucceed(stat))
		for {
			if detour.DtStatusInProgress(stat) {
				var doneIters int
				stat = query.UpdateSlicedFindPath(4, &doneIters)
			}
			if detour.DtStatusSucceed(stat) {
				var path [PATH_MAX_NODE]detour.DtPolyRef
				var pathCount int
				stat = query.FinalizeSlicedFindPath(path[:], &pathCount, PATH_MAX_NODE)
				t.Logf("pathCount: %d\n", pathCount)
				detour.DtAssert(IsEquals(resultValue[resultIndex], float32(pathCount)))
				resultIndex++
				for i := 0; i < pathCount; i++ {
					t.Logf("%d\n", path[i])
					detour.DtAssert(IsEquals(resultValue[resultIndex], float32(path[i])))
					resultIndex++
				}
				break
			}
		}
		t.Logf("\n")
	}

	{
		t.Logf("================================================ SlicedFindPath # DT_FINDPATH_ANY_ANGLE ================================================\n")
		stat = query.InitSlicedFindPath(startRef, endRef, startPos[:], endPos[:], filter, detour.DT_FINDPATH_ANY_ANGLE)
		detour.DtAssert(detour.DtStatusInProgress(stat) || detour.DtStatusSucceed(stat))
		for {
			if detour.DtStatusInProgress(stat) {
				var doneIters int
				stat = query.UpdateSlicedFindPath(4, &doneIters)
			}
			if detour.DtStatusSucceed(stat) {
				var path [PATH_MAX_NODE]detour.DtPolyRef
				var pathCount int
				stat = query.FinalizeSlicedFindPath(path[:], &pathCount, PATH_MAX_NODE)
				t.Logf("pathCount: %d\n", pathCount)
				detour.DtAssert(IsEquals(resultValue[resultIndex], float32(pathCount)))
				resultIndex++
				for i := 0; i < pathCount; i++ {
					t.Logf("%d\n", path[i])
					detour.DtAssert(IsEquals(resultValue[resultIndex], float32(path[i])))
					resultIndex++
				}
				break
			}
		}
		t.Logf("\n")
	}

}
