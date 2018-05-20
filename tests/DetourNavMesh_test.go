package tests

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
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

func Test_dtNavMesh2(t *testing.T) {
	data, err := ioutil.ReadFile("./data.dtMeshHeader")
	detour.DtAssert(err == nil)
	navMesh := detour.DtAllocNavMesh()
	state := navMesh.Init2(data, len(data), 0)
	detour.DtAssert(detour.DtStatusSucceed(state))
}

func Test_dtMeshHeader(t *testing.T) {
	/* data.dtMeshHeader
	   dtMeshHeader header;
	   header.magic = DT_NAVMESH_MAGIC;
	   header.version = DT_NAVMESH_VERSION;
	   header.x = 300;
	   header.y = 4000;
	   header.layer = 50000;
	   header.userId = 2;
	   header.polyCount = 3;
	   header.vertCount = 10002;
	   header.maxLinkCount = 9999;
	   header.detailMeshCount = 8888;
	   header.detailVertCount = 9879;
	   header.detailTriCount = 10923;
	   header.bvNodeCount = 9083;
	   header.offMeshConCount = 908;
	   header.offMeshBase = 102;
	   header.walkableHeight = 90;
	   header.walkableRadius = 765;
	   header.walkableClimb = 9;
	   header.bmin[0] = 1.5f;
	   header.bmin[1] = 2.5f;
	   header.bmin[2] = 3.5f;
	   header.bmax[0] = 2.6f;
	   header.bmax[1] = 3.6f;
	   header.bmax[2] = 4.6f;
	   header.bvQuantFactor = 998;
	   auto f = fopen("./data.dtMeshHeader", "wb");
	   fwrite(&header, sizeof(dtMeshHeader), 1, f);
	   fclose(f);
	*/

	data, err := ioutil.ReadFile("./data.dtMeshHeader")
	detour.DtAssert(err == nil)
	reader := bytes.NewReader(data)
	header := &detour.DtMeshHeader{}
	err = binary.Read(reader, binary.LittleEndian, header)
	detour.DtAssert(err == nil)
	detour.DtAssert(header.Magic == detour.DT_NAVMESH_MAGIC)
	detour.DtAssert(header.Version == detour.DT_NAVMESH_VERSION)
	detour.DtAssert(header.X == 300)
	detour.DtAssert(header.Y == 4000)
	detour.DtAssert(header.Layer == 50000)
	detour.DtAssert(header.UserId == 2)
	detour.DtAssert(header.PolyCount == 3)
	detour.DtAssert(header.VertCount == 10002)
	detour.DtAssert(header.MaxLinkCount == 9999)
	detour.DtAssert(header.DetailMeshCount == 8888)
	detour.DtAssert(header.DetailVertCount == 9879)
	detour.DtAssert(header.DetailTriCount == 10923)
	detour.DtAssert(header.BvNodeCount == 9083)
	detour.DtAssert(header.OffMeshConCount == 908)
	detour.DtAssert(header.OffMeshBase == 102)
	detour.DtAssert(header.WalkableHeight == 90)
	detour.DtAssert(header.WalkableRadius == 765)
	detour.DtAssert(header.WalkableClimb == 9)
	detour.DtAssert(IsEquals(header.Bmin[0], 1.5))
	detour.DtAssert(IsEquals(header.Bmin[1], 2.5))
	detour.DtAssert(IsEquals(header.Bmin[2], 3.5))
	detour.DtAssert(IsEquals(header.Bmax[0], 2.6))
	detour.DtAssert(IsEquals(header.Bmax[1], 3.6))
	detour.DtAssert(IsEquals(header.Bmax[2], 4.6))
	detour.DtAssert(IsEquals(header.BvQuantFactor, 998))
}
