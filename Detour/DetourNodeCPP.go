package detour

import "unsafe"

func DtHashRef(polyRef DtPolyRef) uint32 {
	a := uint32(polyRef)
	a += ^(a << 15)
	a ^= (a >> 10)
	a += (a << 3)
	a ^= (a >> 6)
	a += ^(a << 11)
	a ^= (a >> 16)
	return a
}

func (this *DtNodePool) constructor(maxNodes, hashSize uint32) {
	this.m_maxNodes = maxNodes
	this.m_hashSize = hashSize

	DtAssert(DtNextPow2(this.m_hashSize) == this.m_hashSize)
	// pidx is special as 0 means "none" and 1 is the first node. For that reason
	// we have 1 fewer nodes available than the number of values it can contain.
	DtAssert(this.m_maxNodes > 0 && this.m_maxNodes <= uint32(DT_NULL_IDX) && this.m_maxNodes <= (1<<DT_NODE_PARENT_BITS)-1)

	this.m_nodes = make([]DtNode, this.m_maxNodes)
	this.m_next = make([]DtNodeIndex, this.m_maxNodes)
	this.m_first = make([]DtNodeIndex, this.m_hashSize)

	DtAssert(this.m_nodes != nil)
	DtAssert(this.m_next != nil)
	DtAssert(this.m_first != nil)

	for i := 0; i < len(this.m_first); i++ {
		this.m_first[i] = 0xff
	}
	for i := 0; i < len(this.m_next); i++ {
		this.m_next[i] = 0xff
	}

	this.base = uintptr(unsafe.Pointer(&this.m_nodes))
}

func (this *DtNodePool) destructor() {
	this.m_nodes = nil
	this.m_first = nil
	this.m_next = nil
}

func (this *DtNodePool) Clear() {
	for i := 0; i < len(this.m_first); i++ {
		this.m_first[i] = 0xff
	}
	this.m_nodeCount = 0
}

func (this *DtNodePool) FindNodes(id DtPolyRef, nodes []*DtNode, maxNodes uint32) uint32 {
	var n uint32 = 0
	bucket := DtHashRef(id) & (this.m_hashSize - 1)
	i := this.m_first[bucket]
	for i != DT_NULL_IDX {
		if this.m_nodes[i].Id == id {
			if n >= maxNodes {
				return n
			}
			nodes[n] = &this.m_nodes[i]
			n = n + 1
		}
		i = this.m_next[i]
	}

	return n
}

func (this *DtNodePool) FindNode(id DtPolyRef, state uint8) *DtNode {
	bucket := DtHashRef(id) & (this.m_hashSize - 1)
	i := this.m_first[bucket]
	for i != DT_NULL_IDX {
		if this.m_nodes[i].Id == id && this.m_nodes[i].GetState() == state {
			return &this.m_nodes[i]
		}
		i = this.m_next[i]
	}
	return nil
}

func (this *DtNodePool) GetNode(id DtPolyRef, state uint8) *DtNode {
	bucket := DtHashRef(id) & (this.m_hashSize - 1)
	i := this.m_first[bucket]
	var node *DtNode = nil
	for i != DT_NULL_IDX {
		if this.m_nodes[i].Id == id && this.m_nodes[i].GetState() == state {
			return &this.m_nodes[i]
		}
		i = this.m_next[i]
	}

	if this.m_nodeCount >= this.m_maxNodes {
		return nil
	}

	i = DtNodeIndex(this.m_nodeCount)
	this.m_nodeCount++

	// Init node
	node = &this.m_nodes[i]
	node.SetPidx(0)
	node.Cost = 0
	node.Total = 0
	node.Id = id
	node.SetState(state)
	node.SetFlags(0)

	this.m_next[i] = this.m_first[bucket]
	this.m_first[bucket] = i

	return node
}
