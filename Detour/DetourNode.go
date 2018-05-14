package detour

import (
	"unsafe"
)

type DtNodeFlags int

const (
	DT_NODE_OPEN            DtNodeFlags = 0x01
	DT_NODE_CLOSED          DtNodeFlags = 0x02
	DT_NODE_PARENT_DETACHED DtNodeFlags = 0x04 // parent of the node is not adjacent. Found using raycast.
)

type DtNodeIndex uint16

const DT_NULL_IDX DtNodeIndex = ^DtNodeIndex(0)

const DT_NODE_PARENT_BITS uint32 = 24
const DT_NODE_STATE_BITS uint32 = 2
const DT_NODE_FLAGS_BITS uint32 = 3

type DtNode struct {
	Pos   [3]float32 ///< Position of the node.
	Cost  float32    ///< Cost from previous node to current node.
	Total float32    ///< Cost up to the node.

	//	unsigned int pidx : DT_NODE_PARENT_BITS;	///< Index to parent node.
	//	unsigned int state : DT_NODE_STATE_BITS;	///< extra state information. A polyRef can have multiple nodes with different extra info. see DT_MAX_STATES_PER_NODE
	//	unsigned int flags : 3;						///< Node flags. A combination of dtNodeFlags.
	mixture uint32

	Id DtPolyRef ///< Polygon ref the node corresponds to.
}

/// go no support bitfields. so see GetPidx、SetPidx、GetState、SetState、GetFlags、SetFlags
const DT_NODE_PARENT_MASK = (uint32(1) << DT_NODE_PARENT_BITS) - 1
const DT_NODE_STATE_MASK = ((uint32(1) << DT_NODE_STATE_BITS) - 1) << DT_NODE_PARENT_BITS
const DT_NODE_FLAGS_MASK = ((uint32(1) << DT_NODE_FLAGS_BITS) - 1) << (DT_NODE_PARENT_BITS + DT_NODE_STATE_BITS)
const DT_NODE_PARENT_MASK2 = ^DT_NODE_PARENT_MASK
const DT_NODE_STATE_MASK2 = ^DT_NODE_STATE_MASK
const DT_NODE_FLAGS_MASK2 = ^DT_NODE_FLAGS_MASK

func (this *DtNode) GetPidx() uint32 {
	return this.mixture & DT_NODE_PARENT_MASK
}
func (this *DtNode) SetPidx(pidx uint32) {
	this.mixture &= DT_NODE_PARENT_MASK2
	this.mixture |= pidx
}
func (this *DtNode) GetState() uint32 {
	return (this.mixture & DT_NODE_STATE_MASK) >> DT_NODE_PARENT_BITS
}
func (this *DtNode) SetState(state uint32) {
	this.mixture &= DT_NODE_STATE_MASK2
	this.mixture |= (state << DT_NODE_PARENT_BITS)
}
func (this *DtNode) GetFlags() DtNodeFlags {
	return DtNodeFlags((this.mixture & DT_NODE_FLAGS_MASK) >> (DT_NODE_PARENT_BITS + DT_NODE_STATE_BITS))
}
func (this *DtNode) SetFlags(flags DtNodeFlags) {
	this.mixture &= DT_NODE_FLAGS_MASK2
	this.mixture |= (uint32(flags) << (DT_NODE_PARENT_BITS + DT_NODE_STATE_BITS))
}

const DT_MAX_STATES_PER_NODE int = 1 << DT_NODE_STATE_BITS // number of extra states per node. See dtNode::state

type DtNodePool struct {
	m_nodes     []DtNode
	m_first     []DtNodeIndex
	m_next      []DtNodeIndex
	m_maxNodes  int32
	m_hashSize  int32
	m_nodeCount int32
}

func (this *DtNodePool) GetNodeIdx(node *DtNode) uint32 {
	if node == nil {
		return 0
	}
	base := uintptr(unsafe.Pointer(&this.m_nodes))
	current := uintptr(unsafe.Pointer(node))
	return (uint32)((current-base)/unsafe.Sizeof(*node)) + 1
}

func (this *DtNodePool) GetNodeAtIdx(idx uint32) *DtNode {
	if idx == 0 {
		return nil
	}
	return &this.m_nodes[idx-1]
}

func (this *DtNodePool) getNodeAtIdx(idx uint32) *DtNode {
	if idx == 0 {
		return nil
	}
	return &this.m_nodes[idx-1]
}

func (this *DtNodePool) GetMemUsed() int32 {
	return int32(unsafe.Sizeof(*this)) +
		int32(unsafe.Sizeof(&this.m_nodes[0]))*this.m_maxNodes +
		int32(unsafe.Sizeof(&this.m_next[0]))*this.m_maxNodes +
		int32(unsafe.Sizeof(&this.m_first[0]))*this.m_hashSize
}

func (this *DtNodePool) GetMaxNodes() int32 { return this.m_maxNodes }

func (this *DtNodePool) GetHashSize() int32              { return this.m_hashSize }
func (this *DtNodePool) GetFirst(bucket int) DtNodeIndex { return this.m_first[bucket] }
func (this *DtNodePool) GetNext(i int) DtNodeIndex       { return this.m_next[i] }
func (this *DtNodePool) GetNodeCount() int32             { return this.m_nodeCount }

func DtAllocNodePool(maxNodes, hashSize int32) *DtNodePool {
	pool := &DtNodePool{}
	pool.constructor(maxNodes, hashSize)
	return pool
}

func DtFreeNodePool(pool *DtNodePool) {
	if pool == nil {
		return
	}
	pool.destructor()
}
