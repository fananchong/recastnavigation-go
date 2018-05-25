// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glu

//#include "callback.h"
import "C"
import "github.com/go-gl/gl/v2.1/gl"
import "unsafe"

// ===========================================================================

type TessBeginHandler func(tessType gl.GLenum, polygonData interface{})

//export goTessBeginData
func goTessBeginData(tessType C.GLenum, tessPtr unsafe.Pointer) {
	var tess *Tesselator = (*Tesselator)(tessPtr)
	if tess == nil || tess.beginData == nil {
		return
	}
	tess.beginData(gl.GLenum(tessType), tess.polyData)
}

// ===========================================================================

type TessVertexHandler func(vertexData interface{}, polygonData interface{})

//export goTessVertexData
func goTessVertexData(vertexDataPtr, tessPtr unsafe.Pointer) {
	var tess *Tesselator = (*Tesselator)(tessPtr)
	if tess == nil || tess.vertexData == nil {
		return
	}
	var wrapper *vertexDataWrapper = (*vertexDataWrapper)(vertexDataPtr)
	tess.vertexData(wrapper.data, tess.polyData)
}

// ===========================================================================

type TessEndHandler func(polygonData interface{})

//export goTessEndData
func goTessEndData(tessPtr unsafe.Pointer) {
	var tess *Tesselator = (*Tesselator)(tessPtr)
	if tess == nil || tess.endData == nil {
		return
	}
	tess.endData(tess.polyData)
}

// ===========================================================================

type TessErrorHandler func(errorNumber gl.GLenum, polygonData interface{})

//export goTessErrorData
func goTessErrorData(errorNumber C.GLenum, tessPtr unsafe.Pointer) {
	var tess *Tesselator = (*Tesselator)(tessPtr)
	if tess == nil || tess.errorData == nil {
		return
	}
	tess.errorData(gl.GLenum(errorNumber), tess.polyData)
}

// ===========================================================================

type TessEdgeFlagHandler func(flag bool, polygonData interface{})

//export goTessEdgeFlagData
func goTessEdgeFlagData(flag C.GLboolean, tessPtr unsafe.Pointer) {
	var tess *Tesselator = (*Tesselator)(tessPtr)
	if tess == nil || tess.edgeFlagData == nil {
		return
	}
	var goFlag bool
	if C.GLboolean(0) == flag {
		goFlag = false
	} else {
		goFlag = true
	}

	tess.edgeFlagData(goFlag, tess.polyData)
}

// ===========================================================================

type TessCombineHandler func(coords [3]float64,
	vertexData [4]interface{},
	weight [4]float32,
	polygonData interface{}) (outData interface{})

//export goTessCombineData
func goTessCombineData(coords, vertexData, weight, outData, tessPtr unsafe.Pointer) {
	var tess *Tesselator = (*Tesselator)(tessPtr)
	if tess == nil || tess.combineData == nil {
		return
	}

	var _coords *[3]float64 = (*[3]float64)(coords)
	var _weight *[4]float32 = (*[4]float32)(weight)

	var wrappers *[4]*vertexDataWrapper = (*[4]*vertexDataWrapper)(vertexData)
	var _vertexData [4]interface{}

	for i, wrapper := range *wrappers {
		// Work around for https://bugs.freedesktop.org/show_bug.cgi?id=51641
		// According to documentation, all vertex pointers should be valid.
		if wrapper == nil {
			_vertexData[i] = _vertexData[0]
		} else {
			_vertexData[i] = wrapper.data
		}
	}

	out := tess.combineData(*_coords, _vertexData, *_weight, tess.polyData)
	outWrapper := &vertexDataWrapper{out}

	tess.vertData = append(tess.vertData, outWrapper)
	_outData := (**vertexDataWrapper)(outData)
	*_outData = outWrapper
}

// =============================================================================

// Sets the callback for TESS_BEGIN_DATA.
func (tess *Tesselator) SetBeginCallback(f TessBeginHandler) {
	if tess.tess == nil {
		panic("Uninitialised Tesselator. @see glu.NewTess.")
	}
	tess.beginData = f
	C.setGluTessCallback(tess.tess, C.GLenum(TESS_BEGIN_DATA))
}

// Sets the callback for TESS_VERTEX_DATA.
func (tess *Tesselator) SetVertexCallback(f TessVertexHandler) {
	if tess.tess == nil {
		panic("Uninitialised Tesselator. @see glu.NewTess.")
	}
	tess.vertexData = f
	C.setGluTessCallback(tess.tess, C.GLenum(TESS_VERTEX_DATA))
}

// Sets the callback for TESS_END_DATA.
func (tess *Tesselator) SetEndCallback(f TessEndHandler) {
	if tess.tess == nil {
		panic("Uninitialised Tesselator. @see glu.NewTess.")
	}
	tess.endData = f
	C.setGluTessCallback(tess.tess, C.GLenum(TESS_END_DATA))
}

// Sets the callback for TESS_ERROR_DATA.
func (tess *Tesselator) SetErrorCallback(f TessErrorHandler) {
	if tess.tess == nil {
		panic("Uninitialised Tesselator. @see glu.NewTess.")
	}
	tess.errorData = f
	C.setGluTessCallback(tess.tess, C.GLenum(TESS_ERROR_DATA))
}

// Sets the callback for TESS_EDGE_FLAG_DATA.
func (tess *Tesselator) SetEdgeFlagCallback(f TessEdgeFlagHandler) {
	if tess.tess == nil {
		panic("Uninitialised Tesselator. @see glu.NewTess.")
	}
	tess.edgeFlagData = f
	C.setGluTessCallback(tess.tess, C.GLenum(TESS_EDGE_FLAG_DATA))
}

// Sets the callback for TESS_COMBINE_DATA.
func (tess *Tesselator) SetCombineCallback(f TessCombineHandler) {
	if tess.tess == nil {
		panic("Uninitialised Tesselator. @see glu.NewTess.")
	}
	tess.combineData = f
	C.setGluTessCallback(tess.tess, C.GLenum(TESS_COMBINE_DATA))
}
