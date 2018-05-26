// +build !debug

//
// Copyright (c) 2009-2010 Mikko Mononen memon@inside.org
//
// This software is provided 'as-is', without any express or implied
// warranty.  In no event will the authors be held liable for any damages
// arising from the use of this software.
// Permission is granted to anyone to use this software for any purpose,
// including commercial applications, and to alter it and redistribute it
// freely, subject to the following restrictions:
// 1. The origin of this software must not be misrepresented; you must not
//    claim that you wrote the original software. If you use this software
//    in a product, an acknowledgment in the product documentation would be
//    appreciated but is not required.
// 2. Altered source versions must be plainly marked as such, and must not be
//    misrepresented as being the original software.
// 3. This notice may not be removed or altered from any source distribution.
//

// Note: This header file's only purpose is to include define assert.
// Feel free to change the file and include your own implementation instead.

package detour

/// An assertion failure function.
//  @param[in]		expression  asserted expression.
//  @param[in]		file  Filename of the failed assertion.
//  @param[in]		line  Line number of the failed assertion.
///  @see dtAssertFailSetCustom
type DtAssertFailFunc func(expression bool)

/// Sets the base custom assertion failure function to be used by Detour.
///  @param[in]		assertFailFunc	The function to be invoked in case of failure of #dtAssert
func DtAssertFailSetCustom(assertFailFunc DtAssertFailFunc) {
}

/// Gets the base custom assertion failure function to be used by Detour.
func DtAssertFailGetCustom() DtAssertFailFunc {
	return nil
}

func DtAssert(expression bool) {
}
