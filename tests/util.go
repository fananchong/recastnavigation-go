package tests

import (
	"fmt"
	"io/ioutil"
	"math"
	"unsafe"

	"github.com/fananchong/recastnavigation-go/Detour"
)

func IsEquals(a, b float32) bool {
	return math.Abs(float64(a-b)) < 0.00001
}

type NavMeshSetHeader struct {
	magic      int32
	version    int32
	numTiles   int32
	params     detour.DtNavMeshParams
	boundsMinX float32
	boundsMinY float32
	boundsMinZ float32
	boundsMaxX float32
	boundsMaxY float32
	boundsMaxZ float32
}

type NavMeshTileHeader struct {
	tileRef  detour.DtTileRef
	dataSize int32
}

const NAVMESHSET_MAGIC int32 = int32('M')<<24 | int32('S')<<16 | int32('E')<<8 | int32('T')
const NAVMESHSET_VERSION int32 = 1

func LoadStaticMesh(path string) *detour.DtNavMesh {
	meshData, err := ioutil.ReadFile(path)
	detour.DtAssert(err == nil)

	header := (*NavMeshSetHeader)(unsafe.Pointer(&(meshData[0])))
	detour.DtAssert(header.magic == NAVMESHSET_MAGIC)
	detour.DtAssert(header.version == NAVMESHSET_VERSION)

	fmt.Printf("boundsMin: %f, %f, %f\n", header.boundsMinX, header.boundsMinY, header.boundsMinZ)
	fmt.Printf("boundsMax: %f, %f, %f\n", header.boundsMaxX, header.boundsMaxY, header.boundsMaxZ)

	navMesh := detour.DtAllocNavMesh()
	state := navMesh.Init(&header.params)
	detour.DtAssert(detour.DtStatusSucceed(state))

	d := int32(unsafe.Sizeof(*header))
	for i := 0; i < int(header.numTiles); i++ {
		tileHeader := (*NavMeshTileHeader)(unsafe.Pointer(&(meshData[d])))
		if tileHeader.tileRef == 0 || tileHeader.dataSize == 0 {
			break
		}
		d += int32(unsafe.Sizeof(*tileHeader))

		data := meshData[d : d+tileHeader.dataSize]
		state = navMesh.AddTile(data, int(tileHeader.dataSize), detour.DT_TILE_FREE_DATA, tileHeader.tileRef, nil)
		detour.DtAssert(detour.DtStatusSucceed(state))
		d += tileHeader.dataSize
	}
	return navMesh
}

func CreateQuery(mesh *detour.DtNavMesh, maxNode int) *detour.DtNavMeshQuery {
	query := detour.DtAllocNavMeshQuery()
	detour.DtAssert(query != nil)
	status := query.Init(mesh, maxNode)
	detour.DtAssert(detour.DtStatusSucceed(status))
	return query
}
