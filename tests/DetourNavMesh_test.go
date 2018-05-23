package tests

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"reflect"
	"testing"
	"unsafe"

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

const NAVMESHSET_MAGIC int32 = int32('M')<<24 | int32('S')<<16 | int32('E')<<8 | int32('T')
const NAVMESHSET_VERSION int32 = 1

type NavMeshTileHeader struct {
	tileRef  detour.DtTileRef
	dataSize int32
}

func Test_dtNavMesh2(t *testing.T) {
	meshData, err := ioutil.ReadFile("./nav_test.obj.tile.bin")
	detour.DtAssert(err == nil)

	header := (*NavMeshSetHeader)(unsafe.Pointer(&(meshData[0])))
	detour.DtAssert(header.magic == NAVMESHSET_MAGIC)
	detour.DtAssert(header.version == NAVMESHSET_VERSION)

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

func Test_size(t *testing.T) {

	/*
	   const int headerSize = dtAlign4(sizeof(dtMeshHeader));
	   const int vertsSize = dtAlign4(sizeof(float));
	   const int polysSize = dtAlign4(sizeof(dtPoly));
	   const int linksSize = dtAlign4(sizeof(dtLink));
	   const int detailMeshesSize = dtAlign4(sizeof(dtPolyDetail));
	   const int detailVertsSize = dtAlign4(sizeof(float));
	   const int detailTrisSize = dtAlign4(sizeof(unsigned char));
	   const int bvtreeSize = dtAlign4(sizeof(dtBVNode));
	   const int offMeshLinksSize = dtAlign4(sizeof(dtOffMeshConnection));

	   printf("headerSize: %d\n", headerSize);
	   printf("vertsSize: %d\n", vertsSize);
	   printf("polysSize: %d\n", polysSize);
	   printf("linksSize: %d\n", linksSize);
	   printf("detailMeshesSize: %d\n", detailMeshesSize);
	   printf("detailVertsSize: %d\n", detailVertsSize);
	   printf("detailTrisSize: %d\n", detailTrisSize);
	   printf("bvtreeSize: %d\n", bvtreeSize);
	   printf("offMeshLinksSize: %d\n", offMeshLinksSize);
	*/

	/*
	   headerSize: 100
	   vertsSize: 4
	   polysSize: 32
	   linksSize: 12
	   detailMeshesSize: 12
	   detailVertsSize: 4
	   detailTrisSize: 4
	   bvtreeSize: 16
	   offMeshLinksSize: 36
	*/

	headerSize := detour.DtAlign4(int(unsafe.Sizeof(detour.DtMeshHeader{})))
	detour.DtAssert(headerSize == 100)

	vertsSize := detour.DtAlign4(int(unsafe.Sizeof(float32(1.2))))
	detour.DtAssert(vertsSize == 4)

	polysSize := detour.DtAlign4(int(unsafe.Sizeof(detour.DtPoly{})))
	detour.DtAssert(polysSize == 32)

	linksSize := detour.DtAlign4(int(unsafe.Sizeof(detour.DtLink{})))
	detour.DtAssert(linksSize == 12)

	detailMeshesSize := detour.DtAlign4(int(unsafe.Sizeof(detour.DtPolyDetail{})))
	detour.DtAssert(detailMeshesSize == 12)

	detailVertsSize := detour.DtAlign4(int(unsafe.Sizeof(float32(1.2))))
	detour.DtAssert(detailVertsSize == 4)

	detailTrisSize := detour.DtAlign4(int(unsafe.Sizeof(uint8(1))))
	detour.DtAssert(detailTrisSize == 4)

	bvtreeSize := detour.DtAlign4(int(unsafe.Sizeof(detour.DtBVNode{})))
	detour.DtAssert(bvtreeSize == 16)

	offMeshLinksSize := detour.DtAlign4(int(unsafe.Sizeof(detour.DtOffMeshConnection{})))
	detour.DtAssert(offMeshLinksSize == 36)
}

func Test_data(t *testing.T) {
	a := detour.DtPoly{}
	b := detour.DtPoly{}
	a.FirstLink = 100000
	a.Verts[0] = 200
	a.Verts[1] = 300
	a.Verts[2] = 400
	a.Verts[3] = 500
	a.Verts[4] = 600
	a.Verts[5] = 700
	a.Neis[0] = 8000
	a.Neis[1] = 9000
	a.Neis[2] = 10000
	a.Neis[3] = 11000
	a.Neis[4] = 12000
	a.Neis[5] = 13000
	a.Flags = 14000
	a.VertCount = 200
	a.AreaAndtype = 100
	b.FirstLink = 100001
	b.Verts[0] = 201
	b.Verts[1] = 301
	b.Verts[2] = 401
	b.Verts[3] = 501
	b.Verts[4] = 601
	b.Verts[5] = 701
	b.Neis[0] = 8001
	b.Neis[1] = 9001
	b.Neis[2] = 10001
	b.Neis[3] = 11001
	b.Neis[4] = 12001
	b.Neis[5] = 13001
	b.Flags = 14001
	b.VertCount = 201
	b.AreaAndtype = 101

	detour.DtAssert(detour.DtAlign4(int(unsafe.Sizeof(a))) == 32)
	detour.DtAssert(detour.DtAlign4(int(unsafe.Sizeof(b))) == 32)

	var data []byte
	for i := 0; i < detour.DtAlign4(int(unsafe.Sizeof(a))); i++ {
		data = append(data, *(*uint8)(unsafe.Pointer((uintptr(unsafe.Pointer(&a)) + uintptr(i)))))
	}
	for i := 0; i < detour.DtAlign4(int(unsafe.Sizeof(b))); i++ {
		data = append(data, *(*uint8)(unsafe.Pointer((uintptr(unsafe.Pointer(&b)) + uintptr(i)))))
	}

	s := int(unsafe.Sizeof(a))
	d := 0
	a2 := (*detour.DtPoly)(unsafe.Pointer(&(data[d])))
	d += s
	b2 := (*detour.DtPoly)(unsafe.Pointer(&(data[d])))

	detour.DtAssert(a2.FirstLink == 100000)
	detour.DtAssert(a2.Verts[0] == 200)
	detour.DtAssert(a2.Verts[1] == 300)
	detour.DtAssert(a2.Verts[2] == 400)
	detour.DtAssert(a2.Verts[3] == 500)
	detour.DtAssert(a2.Verts[4] == 600)
	detour.DtAssert(a2.Verts[5] == 700)
	detour.DtAssert(a2.Neis[0] == 8000)
	detour.DtAssert(a2.Neis[1] == 9000)
	detour.DtAssert(a2.Neis[2] == 10000)
	detour.DtAssert(a2.Neis[3] == 11000)
	detour.DtAssert(a2.Neis[4] == 12000)
	detour.DtAssert(a2.Neis[5] == 13000)
	detour.DtAssert(a2.Flags == 14000)
	detour.DtAssert(a2.VertCount == 200)
	detour.DtAssert(a2.AreaAndtype == 100)

	detour.DtAssert(b2.FirstLink == 100001)
	detour.DtAssert(b2.Verts[0] == 201)
	detour.DtAssert(b2.Verts[1] == 301)
	detour.DtAssert(b2.Verts[2] == 401)
	detour.DtAssert(b2.Verts[3] == 501)
	detour.DtAssert(b2.Verts[4] == 601)
	detour.DtAssert(b2.Verts[5] == 701)
	detour.DtAssert(b2.Neis[0] == 8001)
	detour.DtAssert(b2.Neis[1] == 9001)
	detour.DtAssert(b2.Neis[2] == 10001)
	detour.DtAssert(b2.Neis[3] == 11001)
	detour.DtAssert(b2.Neis[4] == 12001)
	detour.DtAssert(b2.Neis[5] == 13001)
	detour.DtAssert(b2.Flags == 14001)
	detour.DtAssert(b2.VertCount == 201)
	detour.DtAssert(b2.AreaAndtype == 101)

	var cc []detour.DtPoly
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&cc)))
	sliceHeader.Cap = 2
	sliceHeader.Len = 2
	sliceHeader.Data = uintptr(unsafe.Pointer(&(data[0])))
	detour.DtAssert(cc[0].FirstLink == 100000)
	detour.DtAssert(cc[0].Verts[0] == 200)
	detour.DtAssert(cc[1].FirstLink == 100001)
	detour.DtAssert(cc[1].Verts[0] == 201)
}
