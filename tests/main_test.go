package tests

import (
	"testing"

	"github.com/fananchong/recastnavigation-go/Detour"
)

var randValue = []float32{
	0.001,
	0.564,
	0.193,
	0.809,
	0.585,
	0.480,
	0.350,
	0.896,
	0.823,
	0.747,
	0.174,
	0.859,
	0.711,
	0.514,
	0.304,
	0.015,
	0.091,
	0.364,
	0.147,
	0.166,
	0.989,
	0.446,
	0.119,
	0.005,
	0.009,
	0.378,
	0.532,
	0.571,
	0.602,
	0.607,
	0.166,
	0.663,
	0.451,
	0.352,
	0.057,
	0.608,
	0.783,
	0.803,
	0.520,
	0.302,
	0.876,
	0.727,
	0.956,
	0.926,
	0.539,
	0.142,
	0.462,
	0.235,
	0.862,
	0.210,
	0.780,
	0.844,
	0.997,
	1.000,
	0.611,
	0.392,
	0.266,
	0.297,
	0.840,
	0.024,
	0.376,
	0.093,
	0.677,
	0.056,
	0.009,
	0.919,
	0.276,
	0.273,
	0.588,
	0.691,
	0.838,
	0.726,
	0.485,
	0.205,
	0.744,
	0.468,
	0.458,
	0.949,
	0.744,
	0.108,
	0.599,
	0.385,
	0.735,
	0.609,
	0.572,
	0.361,
	0.152,
	0.225,
	0.425,
	0.803,
	0.517,
	0.990,
	0.752,
	0.346,
	0.169,
	0.657,
	0.492,
	0.064,
	0.700,
	0.505,
	0.147,
	0.950,
	0.142,
	0.905,
	0.693,
	0.303,
	0.427,
	0.070,
	0.967,
	0.683,
	0.153,
	0.877,
	0.822,
	0.582,
	0.191,
	0.178,
	0.817,
	0.475,
	0.156,
	0.504,
	0.732,
	0.406,
	0.280,
	0.569,
	0.682,
	0.756,
	0.722,
	0.475,
	0.123,
	0.368,
	0.835,
	0.035,
	0.517,
	0.663,
	0.426,
	0.105,
	0.949,
	0.921,
	0.550,
}

var randIndex = 0

func frand() float32 {
	v := randValue[randIndex]
	randIndex++
	return v
}

const PATH_MAX_NODE int = 2048

func Test_main(t *testing.T) {
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

	t.Logf("================================================ findNearestPoly ================================================\n")
	tempPos := [3]float32{0, 0, 0}
	var nearestPos [3]float32
	var nearestRef detour.DtPolyRef
	stat = query.FindNearestPoly(tempPos[:], halfExtents[:], filter, &nearestRef, nearestPos[:])
	detour.DtAssert(detour.DtStatusSucceed(stat))
	t.Logf("nearestPos: %.2f %.2f %.2f\n", nearestPos[0], nearestPos[1], nearestPos[2])
	t.Logf("nearestRef: %d\n", nearestRef)
	t.Logf("\n")

	t.Logf("================================================ findPath ================================================\n")
	var path [PATH_MAX_NODE]detour.DtPolyRef
	var pathCount int
	stat = query.FindPath(startRef, endRef, startPos[:], endPos[:], filter, path[:], &pathCount, PATH_MAX_NODE)
	detour.DtAssert(detour.DtStatusSucceed(stat))
	t.Logf("pathCount: %d\n", pathCount)
	for i := 0; i < pathCount; i++ {
		t.Logf("%d\n", path[i])
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
		for i := 0; i < straightPathCount; i++ {
			t.Logf("straightPath: %.3f %.3f %.3f, straightPathFlags: %d, straightPathRefs: %d\n",
				straightPath[i*3+0], straightPath[i*3+1], straightPath[i*3+2],
				straightPathFlags[i], straightPathRefs[i])
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
		for i := 0; i < straightPathCount; i++ {
			t.Logf("straightPath: %.3f %.3f %.3f, straightPathFlags: %d, straightPathRefs: %d\n",
				straightPath[i*3+0], straightPath[i*3+1], straightPath[i*3+2],
				straightPathFlags[i], straightPathRefs[i])
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
	for i := 0; i < visitedCount; i++ {
		t.Logf("%d\n", visited[i])
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

}
