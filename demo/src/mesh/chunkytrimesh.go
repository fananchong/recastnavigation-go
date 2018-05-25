package mesh

import (
	"fmt"
	"math"
	"sort"
)

type ChunkyTriMeshNode struct {
	bmin [2]float32
	bmax [2]float32
	i    int
	n    int
}

func (n *ChunkyTriMeshNode) I() int {
	return n.i
}

func (n *ChunkyTriMeshNode) N() int {
	return n.n
}

type ChunkyTriMesh struct {
	nodes           []*ChunkyTriMeshNode
	nnodes          int
	tris            []int
	ntris           int
	maxTrisPerChunk int
}

func NewChunkyTriMesh() *ChunkyTriMesh {
	return &ChunkyTriMesh{}
}

func (cm *ChunkyTriMesh) GetNode(i int) *ChunkyTriMeshNode {
	return cm.nodes[i]
}

func (cm *ChunkyTriMesh) GetTris(i int) []int {
	return cm.tris[i:]
}

func (cm *ChunkyTriMesh) PrintTris() {

	for k := 0; k < cm.ntris; k++ {
		if k%16 == 0 {
			fmt.Println()
		}
		fmt.Print(cm.tris[k], " ")
	}
	fmt.Println()
}

func (cm *ChunkyTriMesh) CreateChunkyTriMesh(verts []float32, tris []int, ntris, trisPerChunk int) bool {
	nchunks := (ntris + trisPerChunk - 1) / trisPerChunk
	cm.nodes = make([]*ChunkyTriMeshNode, nchunks*4)
	for i := 0; i < nchunks*4; i++ {
		cm.nodes[i] = &ChunkyTriMeshNode{}
	}
	cm.tris = make([]int, ntris*3)

	cm.ntris = ntris

	// build tree
	items := make([]BoundsItem, ntris)
	// for i := 0; i < ntris; i++ {
	// 	items[i] = &BoundsItem{}
	// }
	for i := 0; i < ntris; i++ {
		t := tris[i*3 : i*3+3]
		it := &items[i]
		it.i = i
		it.bmin[0] = verts[t[0]*3+0]
		it.bmax[0] = verts[t[0]*3+0]
		it.bmin[1] = verts[t[0]*3+2]
		it.bmax[1] = verts[t[0]*3+2]
		for j := 0; j < 3; j++ {
			v := verts[t[j]*3 : t[j]*3+3]
			if v[0] < it.bmin[0] {
				it.bmin[0] = v[0]
			}
			if v[2] < it.bmin[1] {
				it.bmin[1] = v[2]
			}

			if v[0] > it.bmax[0] {
				it.bmax[0] = v[0]
			}
			if v[2] > it.bmax[1] {
				it.bmax[1] = v[2]
			}
		}
	}

	var curTri, curNode int
	subdivide(items, ntris, 0, ntris, trisPerChunk, &curNode, cm.nodes, nchunks*4, &curTri, cm.tris, tris)
	cm.nnodes = curNode

	cm.maxTrisPerChunk = 0
	for i := 0; i < cm.nnodes; i++ {
		node := cm.nodes[i]
		isLeaf := node.i >= 0
		if !isLeaf {
			continue
		}
		if node.n > cm.maxTrisPerChunk {
			cm.maxTrisPerChunk = node.n
		}
	}
	return true
}

func (cm *ChunkyTriMesh) GetChunksOverlappingSegment(p, q *[2]float32, ids []int, maxIds int) int {
	i, n := 0, 0
	for i < cm.nnodes {
		node := cm.nodes[i]
		overlap := checkOverlapSegment(p, q, &node.bmin, &node.bmax)
		isLeafNode := node.i >= 0

		if isLeafNode && overlap {
			if n < maxIds {
				ids[n] = i
				n++
			}
		}

		if overlap || isLeafNode {
			i++
		} else {
			escapeIndex := -node.i
			i += escapeIndex
		}

	}
	return n
}

func calcExtends(items []BoundsItem, nitems, imin, imax int, bmin, bmax *[2]float32) {
	bmin[0] = items[imin].bmin[0]
	bmin[1] = items[imin].bmin[1]

	bmax[0] = items[imin].bmax[0]
	bmax[1] = items[imin].bmax[1]

	for i := imin + 1; i < imax; i++ {
		it := &items[i]
		if it.bmin[0] < bmin[0] {
			bmin[0] = it.bmin[0]
		}
		if it.bmin[1] < bmin[1] {
			bmin[1] = it.bmin[1]
		}

		if it.bmax[0] > bmax[0] {
			bmax[0] = it.bmax[0]
		}
		if it.bmax[1] > bmax[1] {
			bmax[1] = it.bmax[1]
		}
	}
}

func subdivide(items []BoundsItem, nitems, imin, imax, trisPerChunk int,
	curNode *int, nodes []*ChunkyTriMeshNode, maxNodes int, curTri *int, outTris, inTris []int) {

	inum := imax - imin
	icur := *curNode
	if *curNode > maxNodes {
		return
	}

	node := nodes[*curNode]
	(*curNode)++

	if inum <= trisPerChunk {

		// leaf
		calcExtends(items, nitems, imin, imax, &node.bmin, &node.bmax)

		// copy triangles
		node.i = *curTri
		node.n = inum

		for i := imin; i < imax; i++ {
			src := inTris[items[i].i*3:]
			dst := outTris[*curTri*3:]
			(*curTri)++
			dst[0], dst[1], dst[2] = src[0], src[1], src[2]
		}
	} else {
		// split
		calcExtends(items, nitems, imin, imax, &node.bmin, &node.bmax)

		axis := longestAxis(node.bmax[0]-node.bmin[0], node.bmax[1]-node.bmin[1])
		if axis == 0 {
			sort.Sort(NewBoundsSorter(items[imin:imax], func(a, b *BoundsItem) bool {
				return a.bmin[0] < b.bmin[0]
			}))
		} else if axis == 1 {
			sort.Sort(NewBoundsSorter(items[imin:imax], func(a, b *BoundsItem) bool {
				return a.bmin[1] < b.bmin[1]
			}))
		}

		isplit := imin + inum/2

		//left
		subdivide(items, nitems, imin, isplit, trisPerChunk, curNode, nodes, maxNodes, curTri, outTris, inTris)
		//right
		subdivide(items, nitems, isplit, imax, trisPerChunk, curNode, nodes, maxNodes, curTri, outTris, inTris)

		iescape := *curNode - icur
		node.i = -iescape
	}
}

func longestAxis(x, y float32) int {
	if y > x {
		return 1
	}
	return 0
}

func checkOverlapSegment(p, q, bmin, bmax *[2]float32) bool {
	EPSILON := 1e-6
	var tmin, tmax float32 = 0, 1
	var d = [2]float32{q[0] - p[0], q[1] - p[1]}

	for i := 0; i < 2; i++ {
		// Ray is parallel to slab. No hit if origin not within slab
		if math.Abs(float64(d[i])) < EPSILON {
			if p[i] < bmin[i] || p[i] > bmax[i] {

				return false
			}
		} else {
			// Ray is parallel to slab. No hit if origin not within slab
			ood := 1.0 / d[i]
			t1 := (bmin[i] - p[i]) * ood
			t2 := (bmax[i] - p[i]) * ood

			if t1 > t2 {
				t1, t2 = t2, t1
			}

			if t1 > tmin {
				tmin = t1
			}
			if t2 < tmax {
				tmax = t2
			}
			if tmin > tmax {
				return false
			}
		}
	}
	return true
}
