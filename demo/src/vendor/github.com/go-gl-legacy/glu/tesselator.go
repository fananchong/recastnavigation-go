// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glu

// #ifdef __APPLE__
//   #include <OpenGL/glu.h>
// #else
//   #include <GL/glu.h>
// #endif
import "C"
import "github.com/go-gl/gl/v2.1/gl"
import "unsafe"

// Opaque object used for book keeping on the go side.
type Tesselator struct {
	tess *C.GLUtesselator

	polyData interface{}

	// vertData keeps references to the vertices specifed by TessVertex
	// so that the garbage collector does not invalidate them.
	vertData []*vertexDataWrapper

	// vertLocs stores a copy of the vertices' locations as specified
	// to TessVertex. Again, so the garbage collector doesn't get them.
	vertLocs [][3]float64

	beginData    TessBeginHandler
	vertexData   TessVertexHandler
	endData      TessEndHandler
	errorData    TessErrorHandler
	edgeFlagData TessEdgeFlagHandler
	combineData  TessCombineHandler
}

// Wrapper around an interface. Does go not support (*interface{})(ptr)?
type vertexDataWrapper struct {
	data interface{}
}

// Create a new tesselator.
func NewTess() (tess *Tesselator) {
	tess = new(Tesselator)
	tess.tess = C.gluNewTess()

	if tess.tess == nil {
		panic("Out of memory.")
	}

	return
}

// Clean up resources held by the tesselator. Go's garbage collector cannot
// do this automatically.
func (tess *Tesselator) Delete() {
	C.gluDeleteTess(tess.tess)
	tess.tess = nil
}

// Begin the drawing of the polygon, with the data parameter that will
// be provided to callbacks.
func (tess *Tesselator) BeginPolygon(data interface{}) {
	tess.polyData = data
	C.gluTessBeginPolygon(tess.tess, unsafe.Pointer(tess))
}

// End the drawing of the polygon.
func (tess *Tesselator) EndPolygon() {
	C.gluTessEndPolygon(tess.tess)

	// Free memory that we were safeguarding on the go side.
	tess.vertData = []*vertexDataWrapper{}
	tess.vertLocs = [][3]float64{}
}

// Begin a contour within the polygon.
func (tess *Tesselator) BeginContour() {
	C.gluTessBeginContour(tess.tess)
}

// End a contour within the polygon.
func (tess *Tesselator) EndContour() {
	C.gluTessEndContour(tess.tess)
}

// Add a vertex to the polygon, with the data parameter that will be
// provided to callbacks.
func (tess *Tesselator) Vertex(location [3]float64, data interface{}) {
	// Wrap and safeguard data pointer.
	_data := &vertexDataWrapper{data}
	tess.vertData = append(tess.vertData, _data)

	// Copy location to a safe memory location.
	tess.vertLocs = append(tess.vertLocs, location)
	_location := unsafe.Pointer(&tess.vertLocs[len(tess.vertLocs)-1])

	C.gluTessVertex(tess.tess, (*C.GLdouble)(_location), unsafe.Pointer(_data))
}

// Set the normal of the plane onto which points are projected onto before tesselation.
func (tess *Tesselator) Normal(valueX, valueY, valueZ float64) {
	cx := C.GLdouble(valueX)
	cy := C.GLdouble(valueY)
	cz := C.GLdouble(valueZ)
	C.gluTessNormal(tess.tess, cx, cy, cz)
}

// Set a property of the tesselator.
func (tess *Tesselator) Property(which gl.GLenum, data float64) {
	C.gluTessProperty(tess.tess, C.GLenum(which), C.GLdouble(data))
}
