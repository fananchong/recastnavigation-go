package tests

import (
	"math/rand"
	"testing"
	"time"

	"github.com/fananchong/recastnavigation-go/Detour"
)

func Test_dtNodePool(t *testing.T) {
	detour.DtAssert(uint16(detour.DT_NULL_IDX) == 0xffff)
	detour.DtAssert(detour.DT_NODE_PARENT_MASK == 0x00ffffff)
	detour.DtAssert(detour.DT_NODE_STATE_MASK == 0x03000000)
	detour.DtAssert(detour.DT_NODE_FLAGS_MASK == 0x1C000000)

	node := &detour.DtNode{}
	node.SetPidx(7)
	node.SetState(2)
	node.SetFlags(detour.DT_NODE_PARENT_DETACHED)
	detour.DtAssert(node.GetPidx() == 7 && node.GetState() == 2 && node.GetFlags() == detour.DT_NODE_PARENT_DETACHED)

	pool := detour.DtAllocNodePool(50, 4)
	ns1 := make([]*detour.DtNode, 0)
	for i := 0; i < 25; i++ {
		ns1 = append(ns1, pool.GetNode(detour.DtPolyRef(i), 1))
	}
	ns2 := make([]*detour.DtNode, 0)
	for i := 0; i < 25; i++ {
		ns2 = append(ns2, pool.GetNode(detour.DtPolyRef(i), 2))
	}
	detour.DtAssert(pool.GetNode(51, 3) == nil)
	for i := 0; i < 25; i++ {
		detour.DtAssert(pool.GetNodeAtIdx(uint32(i+1)) == ns1[i])
		detour.DtAssert(pool.GetNodeAtIdx(uint32(25+i+1)) == ns2[i])
		detour.DtAssert(pool.GetNodeIdx(ns1[i]) == uint32(i+1))
		detour.DtAssert(pool.GetNodeIdx(ns2[i]) == uint32(25+i+1))
		detour.DtAssert(pool.FindNode(detour.DtPolyRef(i), 1) == ns1[i])
		detour.DtAssert(pool.FindNode(detour.DtPolyRef(i), 2) == ns2[i])
		temps := [4]*detour.DtNode{}
		tempn := pool.FindNodes(detour.DtPolyRef(i), temps[:], 4)
		detour.DtAssert(tempn == 2)
		detour.DtAssert(temps[0] == ns2[i])
		detour.DtAssert(temps[1] == ns1[i])
	}
	detour.DtAssert(pool.GetNodeCount() == 50)
	detour.DtFreeNodePool(pool)
}

func Test_dtNodeQueue(t *testing.T) {
	rand.Seed(time.Now().Unix())
	n := 32
	queue := detour.DtAllocNodeQueue(n)
	for i := 0; i < n; i++ {
		node := &detour.DtNode{}
		node.Total = float32(rand.Intn(100))
		queue.Push(node)
	}
	nodes := make([]*detour.DtNode, 0)
	for i := 0; i < n; i++ {
		nodes = append(nodes, queue.Pop())
	}
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			detour.DtAssert(nodes[i].Total <= nodes[j].Total)
		}
	}
}
