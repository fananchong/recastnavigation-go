package tests

import (
	"testing"

	"github.com/fananchong/recastnavigation-go/Detour"
)

func Test_DT_NULL_IDX(t *testing.T) {
	detour.DtAssert(uint16(detour.DT_NULL_IDX) == 0xffff)
	detour.DtAssert(detour.DT_NODE_PARENT_MASK == 0x00ffffff)
	detour.DtAssert(detour.DT_NODE_STATE_MASK == 0x03000000)
	detour.DtAssert(detour.DT_NODE_FLAGS_MASK == 0x1C000000)

	node := &detour.DtNode{}
	node.SetPidx(7)
	node.SetState(2)
	node.SetFlags(detour.DT_NODE_PARENT_DETACHED)
	detour.DtAssert(node.GetPidx() == 7 && node.GetState() == 2 && node.GetFlags() == detour.DT_NODE_PARENT_DETACHED)
}
