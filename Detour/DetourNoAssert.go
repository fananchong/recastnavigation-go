// +build !debug

package detour

type DtAssertFailFunc func(expression bool)

func DtAssertFailSetCustom(assertFailFunc DtAssertFailFunc) {
}

func DtAssertFailGetCustom() DtAssertFailFunc {
	return nil
}

func DtAssert(expression bool) {
}
