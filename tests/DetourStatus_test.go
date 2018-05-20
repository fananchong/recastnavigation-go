package tests

import (
	"testing"

	"github.com/fananchong/recastnavigation-go/Detour"
)

func Test_dtStatus(t *testing.T) {
	state1 := detour.DT_SUCCESS | detour.DT_IN_PROGRESS
	state2 := detour.DT_FAILURE | detour.DT_INVALID_PARAM
	detour.DtAssert(detour.DtStatusSucceed(state1))
	detour.DtAssert(detour.DtStatusFailed(state1) == false)
	detour.DtAssert(detour.DtStatusFailed(state2))
	detour.DtAssert(detour.DtStatusInProgress(state1))
	detour.DtAssert(detour.DtStatusDetail(state2, detour.DT_INVALID_PARAM))
}
