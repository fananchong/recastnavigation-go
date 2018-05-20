package tests

import (
	"testing"

	"github.com/fananchong/recastnavigation-go/Detour"
)

func Test_dtNavMeshParams(t *testing.T) {
	params1 := &detour.DtNavMeshParams{}
	params1.MaxPolys = 1
	params1.MaxTiles = 2
	params1.Orig[0] = 3
	params1.Orig[1] = 4
	params1.Orig[2] = 5
	params1.TileHeight = 6
	params1.TileWidth = 7

	params2 := *params1
	detour.DtAssert(params2.MaxPolys == params1.MaxPolys)
	detour.DtAssert(params2.MaxTiles == params1.MaxTiles)
	detour.DtAssert(params2.Orig[0] == params1.Orig[0])
	detour.DtAssert(params2.Orig[1] == params1.Orig[1])
	detour.DtAssert(params2.Orig[2] == params1.Orig[2])
	detour.DtAssert(params2.TileHeight == params1.TileHeight)
	detour.DtAssert(params2.TileWidth == params1.TileWidth)
}

func Test_dtNavMesh1(t *testing.T) {
	navMesh := detour.DtAllocNavMesh()
	params := &detour.DtNavMeshParams{}
	params.MaxPolys = 3000
	params.MaxTiles = 16
	params.Orig[0] = 3
	params.Orig[1] = 4
	params.Orig[2] = 5
	params.TileHeight = 6
	params.TileWidth = 7
	navMesh.Init(params)

	var salt1, it1, ip1, salt2, it2, ip2 uint32
	salt1, it1, ip1 = 1, 2, 3
	id := navMesh.EncodePolyId(salt1, it1, ip1)
	navMesh.DecodePolyId(id, &salt2, &it2, &ip2)
	detour.DtAssert(salt1 == salt2)
	detour.DtAssert(it1 == it2)
	detour.DtAssert(ip1 == ip2)
}
