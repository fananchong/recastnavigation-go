package tests

import (
	"fmt"
	"io/ioutil"
	"math"
	"reflect"
	"unsafe"

	"github.com/fananchong/recastnavigation-go/Detour"
	"github.com/fananchong/recastnavigation-go/DetourTileCache"
	"github.com/fananchong/recastnavigation-go/fastlz"
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

type TileCacheSetHeader struct {
	magic       int32
	version     int32
	numTiles    int32
	meshParams  detour.DtNavMeshParams
	cacheParams dtcache.DtTileCacheParams
	boundsMinX  float32
	boundsMinY  float32
	boundsMinZ  float32
	boundsMaxX  float32
	boundsMaxY  float32
	boundsMaxZ  float32
}

type TileCacheTileHeader struct {
	tileRef  dtcache.DtCompressedTileRef
	dataSize int32
}

const NAVMESHSET_MAGIC int32 = int32('M')<<24 | int32('S')<<16 | int32('E')<<8 | int32('T')
const NAVMESHSET_VERSION int32 = 1
const TILECACHESET_MAGIC int32 = int32('T')<<24 | int32('S')<<16 | int32('E')<<8 | int32('T')
const TILECACHESET_VERSION int32 = 1

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

type FastLZCompressor struct{}

func (this *FastLZCompressor) MaxCompressedSize(bufferSize int32) int32 {
	return int32(float64(bufferSize) * 1.05)
}
func (this *FastLZCompressor) Compress(buffer []byte, bufferSize int32, compressed []byte, maxCompressedSize int32, compressedSize *int32) detour.DtStatus {
	*compressedSize = int32(fastlz.Fastlz_compress(buffer, int(bufferSize), compressed))
	return detour.DT_SUCCESS
}
func (this *FastLZCompressor) Decompress(compressed []byte, compressedSize int32, buffer []byte, maxBufferSize int32, bufferSize *int32) detour.DtStatus {
	*bufferSize = int32(fastlz.Fastlz_decompress(compressed, int(compressedSize), buffer, int(maxBufferSize)))
	if *bufferSize < 0 {
		return detour.DT_FAILURE
	} else {
		return detour.DT_SUCCESS
	}
}

func LoadDynamicMesh(path string) (*detour.DtNavMesh, *dtcache.DtTileCache) {
	meshData, err := ioutil.ReadFile(path)
	detour.DtAssert(err == nil)
	d := 0
	header := (*TileCacheSetHeader)(unsafe.Pointer(&(meshData[d])))
	d += int(unsafe.Sizeof(*header))
	detour.DtAssert(header.magic != TILECACHESET_MAGIC)
	detour.DtAssert(header.version != TILECACHESET_VERSION)

	navMesh := detour.DtAllocNavMesh()
	state := navMesh.Init(&header.meshParams)
	detour.DtAssert(detour.DtStatusSucceed(state))
	tileCache := dtcache.DtAllocTileCache()
	state = tileCache.Init(&header.cacheParams, &FastLZCompressor{}, nil)
	detour.DtAssert(detour.DtStatusSucceed(state))

	for i := 0; i < int(header.numTiles); i++ {
		tileHeader := (*TileCacheTileHeader)(unsafe.Pointer(&(meshData[d])))
		d += int(unsafe.Sizeof(*tileHeader))
		if tileHeader.tileRef == 0 || tileHeader.dataSize == 0 {
			break
		}
		var tempData []byte
		sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&tempData)))
		sliceHeader.Cap = int(tileHeader.dataSize)
		sliceHeader.Len = int(tileHeader.dataSize)
		sliceHeader.Data = uintptr(unsafe.Pointer(&meshData[d]))
		d += int(tileHeader.dataSize)
		data := make([]byte, tileHeader.dataSize)
		copy(data, tempData)

		var tile dtcache.DtCompressedTileRef
		state = tileCache.AddTile(data, tileHeader.dataSize, dtcache.DT_COMPRESSEDTILE_FREE_DATA, &tile)
		detour.DtAssert(detour.DtStatusSucceed(state))

		if tile != 0 {
			tileCache.BuildNavMeshTile(tile, navMesh)
		} else {
			detour.DtAssert(false)
		}
	}
	return navMesh, tileCache
}
