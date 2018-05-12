// +build debug

package detour

type DtAssertFailFunc func(expression bool)

var sAssertFailFunc DtAssertFailFunc = nil

func DtAssertFailSetCustom(assertFailFunc DtAssertFailFunc) {
	sAssertFailFunc = assertFailFunc
}

func DtAssertFailGetCustom() DtAssertFailFunc {
	return sAssertFailFunc
}

func DtAssert(expression bool) {
	failFunc := DtAssertFailGetCustom()
	if failFunc == nil {
		if !expression {
			panic("DtAssert")
		}
	} else if !expression {
		failFunc(expression)
	}
}
