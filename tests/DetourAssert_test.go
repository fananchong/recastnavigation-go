package tests

import (
	"testing"

	"github.com/fananchong/recastnavigation-go/Detour"
)

func Test_dtAssert(t *testing.T) {
	detour.DtAssert(true)

	detour.DtAssertFailSetCustom(
		func(expression bool) {
			if !expression {
				t.Logf("here!\n")
			}
		})
	detour.DtAssert(false)
}
