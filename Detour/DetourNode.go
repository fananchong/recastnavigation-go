package detour

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
