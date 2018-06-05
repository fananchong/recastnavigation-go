package tests

import (
	"fmt"
	"io/ioutil"
	"math"
	"reflect"
	"unsafe"

	detour "github.com/fananchong/recastnavigation-go/Detour"
	dtcache "github.com/fananchong/recastnavigation-go/DetourTileCache"
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

const (
	POLYAREA_GROUND uint8 = 0
	POLYAREA_WATER  uint8 = 1
	POLYAREA_ROAD   uint8 = 2
	POLYAREA_DOOR   uint8 = 3
	POLYAREA_GRASS  uint8 = 4
	POLYAREA_JUMP   uint8 = 5
)

const (
	POLYFLAGS_WALK     uint16 = 0x01   // Ability to walk (ground, grass, road)
	POLYFLAGS_SWIM     uint16 = 0x02   // Ability to swim (water).
	POLYFLAGS_DOOR     uint16 = 0x04   // Ability to move through doors.
	POLYFLAGS_JUMP     uint16 = 0x08   // Ability to jump.
	POLYFLAGS_DISABLED uint16 = 0x10   // Disabled polygon
	POLYFLAGS_ALL      uint16 = 0xffff // All abilities.
)

type MeshProcess struct{}

func (this *MeshProcess) Process(params *detour.DtNavMeshCreateParams, polyAreas []uint8, polyFlags []uint16) {
	// Update poly flags from areas.
	for i := 0; i < int(params.PolyCount); i++ {
		if polyAreas[i] == dtcache.DT_TILECACHE_WALKABLE_AREA {
			polyAreas[i] = POLYAREA_GROUND
		}
		if polyAreas[i] == POLYAREA_GROUND ||
			polyAreas[i] == POLYAREA_GRASS ||
			polyAreas[i] == POLYAREA_ROAD {
			polyFlags[i] = POLYFLAGS_WALK
		} else if polyAreas[i] == POLYAREA_WATER {
			polyFlags[i] = POLYFLAGS_SWIM
		} else if polyAreas[i] == POLYAREA_DOOR {
			polyFlags[i] = POLYFLAGS_WALK | POLYFLAGS_DOOR
		}
	}

	// TODO: Pass in off-mesh connections.
}

func LoadDynamicMesh(path string) (*detour.DtNavMesh, *dtcache.DtTileCache) {
	meshData, err := ioutil.ReadFile(path)
	detour.DtAssert(err == nil)
	d := 0
	header := (*TileCacheSetHeader)(unsafe.Pointer(&(meshData[d])))
	d += int(unsafe.Sizeof(*header))
	detour.DtAssert(header.magic == TILECACHESET_MAGIC)
	detour.DtAssert(header.version == TILECACHESET_VERSION)

	navMesh := detour.DtAllocNavMesh()
	state := navMesh.Init(&header.meshParams)
	detour.DtAssert(detour.DtStatusSucceed(state))
	tileCache := dtcache.DtAllocTileCache()
	state = tileCache.Init(&header.cacheParams, &FastLZCompressor{}, &MeshProcess{})
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

func FindRandomPoint(query *detour.DtNavMeshQuery, filter *detour.DtQueryFilter, frand func() float32,
	randomRef *detour.DtPolyRef, randomPt []float32) detour.DtStatus {
	m_nav := query.GetAttachedNavMesh()
	detour.DtAssert(m_nav != nil)

	// Randomly pick one tile. Assume that all tiles cover roughly the same area.
	tileIndex := int(frand() * float32(m_nav.GetMaxTiles()))
	var tile *detour.DtMeshTile
	for i := tileIndex; true; i++ {
		i = i % int(m_nav.GetMaxTiles())
		tile = m_nav.GetTile(i)
		if tile != nil && tile.Header != nil {
			break
		}
	}
	if tile == nil {
		return detour.DT_FAILURE
	}
	// Randomly pick one polygon weighted by polygon area.
	var poly *detour.DtPoly
	var polyRef detour.DtPolyRef
	base := m_nav.GetPolyRefBase(tile)

	var areaSum float32
	for i := 0; i < int(tile.Header.PolyCount); i++ {
		p := &tile.Polys[i]
		// Do not return off-mesh connection polygons.
		if p.GetType() != detour.DT_POLYTYPE_GROUND {
			continue
		}
		// Must pass filter
		ref := base | (detour.DtPolyRef)(i)
		if !filter.PassFilter(ref, tile, p) {
			continue
		}
		// Calc area of the polygon.
		var polyArea float32
		for j := 2; j < int(p.VertCount); j++ {
			va := tile.Verts[p.Verts[0]*3:]
			vb := tile.Verts[p.Verts[j-1]*3:]
			vc := tile.Verts[p.Verts[j]*3:]
			polyArea += detour.DtTriArea2D(va, vb, vc)
		}

		// Choose random polygon weighted by area, using reservoi sampling.
		areaSum += polyArea
		u := frand()
		if u*areaSum <= polyArea {
			poly = p
			polyRef = ref
		}
	}

	if poly == nil {
		return detour.DT_FAILURE
	}
	// Randomly pick point on polygon.
	v := tile.Verts[poly.Verts[0]*3:]
	var verts [3 * detour.DT_VERTS_PER_POLYGON]float32
	var areas [detour.DT_VERTS_PER_POLYGON]float32
	detour.DtVcopy(verts[0*3:], v)
	for j := 1; j < int(poly.VertCount); j++ {
		v = tile.Verts[poly.Verts[j]*3:]
		detour.DtVcopy(verts[j*3:], v)
	}

	s := frand()
	t := frand()

	var pt [3]float32
	detour.DtRandomPointInConvexPoly(verts[:], int(poly.VertCount), areas[:], s, t, pt[:])

	var h float32
	status := query.GetPolyHeight(polyRef, pt[:], &h)
	if detour.DtStatusFailed(status) {
		return status
	}
	pt[1] = h

	detour.DtVcopy(randomPt, pt[:])
	*randomRef = polyRef

	return detour.DT_SUCCESS
}
