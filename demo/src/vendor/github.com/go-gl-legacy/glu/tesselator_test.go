// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glu

import (
	"github.com/go-gl/gl"
	"testing"
)

type PolygonData struct {
	BeginCount    int
	VertexCount   int
	EndCount      int
	ErrorCount    int
	EdgeFlagCount int
	CombineCount  int

	Vertices []VertexData
}

type VertexData struct {
	Location    [3]float64
	VertexHits  int
	CombineHits int
}

// Test shape is a square with a square hole inside.
var OuterContour [4][3]float64 = [4][3]float64{[3]float64{-2, 2, 0},
	[3]float64{-2, -2, 0},
	[3]float64{2, -2, 0},
	[3]float64{2, 2, 0}}

var InnerContour [4][3]float64 = [4][3]float64{[3]float64{-1, 1, 0},
	[3]float64{1, 1, 0},
	[3]float64{1, -1, 0},
	[3]float64{-1, -1, 0}}

// Pentagram with crossing edges. Invokes the combine callback.
var StarContour [5][3]float64 = [5][3]float64{[3]float64{0, 1, 0},
	[3]float64{-1, -1, 0},
	[3]float64{1, 0, 0},
	[3]float64{-1, 0, 0},
	[3]float64{1, -1, 0}}

func TestTesselatorData(t *testing.T) {
	poly := new(PolygonData)

	for _, v := range OuterContour {
		poly.Vertices = append(poly.Vertices, VertexData{Location: v})
	}
	for _, v := range InnerContour {
		poly.Vertices = append(poly.Vertices, VertexData{Location: v})
	}

	tess := NewTess()

	tess.SetBeginCallback(tessBeginDataHandler)
	tess.SetVertexCallback(tessVertexDataHandler)
	tess.SetEndCallback(tessEndDataHandler)
	tess.SetErrorCallback(tessErrorDataHandler)
	tess.SetEdgeFlagCallback(tessEdgeFlagDataHandler)
	tess.SetCombineCallback(tessCombineDataHandler)

	tess.Normal(0, 0, 1)

	tess.BeginPolygon(poly)
	tess.BeginContour()

	for v := 0; v < 4; v += 1 {
		tess.Vertex(poly.Vertices[v].Location, &poly.Vertices[v])
	}

	tess.EndContour()
	tess.BeginContour()

	for v := 4; v < 8; v += 1 {
		tess.Vertex(poly.Vertices[v].Location, &poly.Vertices[v])
	}

	tess.EndContour()
	tess.EndPolygon()

	expectedTriangles := 8
	// There are a total of 24 edges, 8 of which are not edges. This means
	// the EdgeFlag must be toggled to true 8 times.
	expectedEdges := 8

	checkPoly(t, poly, 1, expectedTriangles*3, 1, 0, expectedEdges, 0)

	tess.Delete()
}

func TestTesselatorNil(t *testing.T) {
	poly := new(PolygonData)

	for _, v := range OuterContour {
		poly.Vertices = append(poly.Vertices, VertexData{Location: v})
	}
	for _, v := range InnerContour {
		poly.Vertices = append(poly.Vertices, VertexData{Location: v})
	}

	tess := NewTess()

	tess.SetBeginCallback(tessBeginNilHandler)
	tess.SetVertexCallback(tessVertexNilHandler)
	tess.SetEndCallback(tessEndNilHandler)
	tess.SetErrorCallback(tessErrorNilHandler)
	tess.SetEdgeFlagCallback(tessEdgeFlagNilHandler)
	tess.SetCombineCallback(tessCombineNilHandler)

	tess.Normal(0, 0, 1)

	tess.BeginPolygon(nil)
	tess.BeginContour()

	for v := 0; v < 4; v += 1 {
		tess.Vertex(poly.Vertices[v].Location, nil)
	}

	tess.EndContour()
	tess.BeginContour()

	for v := 4; v < 8; v += 1 {
		tess.Vertex(poly.Vertices[v].Location, nil)
	}

	tess.EndContour()
	tess.EndPolygon()

	tess.Delete()
}

func TestTesselatorStar(t *testing.T) {
	poly := new(PolygonData)

	for _, v := range StarContour {
		poly.Vertices = append(poly.Vertices, VertexData{Location: v})
	}

	tess := NewTess()

	tess.SetBeginCallback(tessBeginDataHandler)
	tess.SetVertexCallback(tessVertexDataHandler)
	tess.SetEndCallback(tessEndDataHandler)
	tess.SetErrorCallback(tessErrorDataHandler)
	tess.SetEdgeFlagCallback(tessEdgeFlagDataHandler)
	tess.SetCombineCallback(tessCombineDataHandler)

	tess.Normal(0, 0, 1)

	tess.BeginPolygon(poly)
	tess.BeginContour()

	for v := range poly.Vertices {
		tess.Vertex(poly.Vertices[v].Location, &poly.Vertices[v])
	}

	tess.EndContour()
	tess.EndPolygon()

	expectedTriangles := 5
	// All edges lie on the edge of the polygon, so this is only set once.
	expectedEdges := 1
	expectedCombines := 5

	checkPoly(t, poly, 1, expectedTriangles*3, 1, 0, expectedEdges, expectedCombines)

	tess.Delete()
}

func tessBeginDataHandler(tessType gl.GLenum, polygonData interface{}) {
	polygonDataPtr := polygonData.(*PolygonData)
	polygonDataPtr.BeginCount += 1
}

func tessVertexDataHandler(vertexData interface{}, polygonData interface{}) {
	polygonDataPtr := polygonData.(*PolygonData)
	polygonDataPtr.VertexCount += 1

	vertexDataPtr := vertexData.(*VertexData)
	vertexDataPtr.VertexHits += 1
}

func tessEndDataHandler(polygonData interface{}) {
	polygonDataPtr := polygonData.(*PolygonData)
	polygonDataPtr.EndCount += 1
}

func tessErrorDataHandler(errno gl.GLenum, polygonData interface{}) {
	polygonDataPtr := polygonData.(*PolygonData)
	polygonDataPtr.ErrorCount += 1
}

func tessEdgeFlagDataHandler(flag bool, polygonData interface{}) {
	polygonDataPtr := polygonData.(*PolygonData)

	if flag {
		polygonDataPtr.EdgeFlagCount += 1
	}
}

func tessCombineDataHandler(coords [3]float64,
	vertexData [4]interface{},
	weight [4]float32,
	polygonData interface{}) (outData interface{}) {

	polygonDataPtr := polygonData.(*PolygonData)
	polygonDataPtr.CombineCount += 1

	for _, v := range vertexData {
		vertexDataPtr := v.(*VertexData)
		vertexDataPtr.CombineHits += 1
	}

	newVertex := VertexData{Location: coords}
	polygonDataPtr.Vertices = append(polygonDataPtr.Vertices, newVertex)

	return &polygonDataPtr.Vertices[len(polygonDataPtr.Vertices)-1]
}

func tessBeginNilHandler(tessType gl.GLenum, polygonData interface{}) {
}

func tessVertexNilHandler(vertexData interface{}, polygonData interface{}) {
}

func tessEndNilHandler(polygonData interface{}) {
}

func tessErrorNilHandler(errno gl.GLenum, polygonData interface{}) {
}

func tessEdgeFlagNilHandler(flag bool, polygonData interface{}) {
}

func tessCombineNilHandler(coords [3]float64,
	vertexData [4]interface{},
	weight [4]float32,
	polygonData interface{}) (outData interface{}) {
	return nil
}

func checkPoly(t *testing.T, poly *PolygonData, expectedBegins, expectedVertices,
	expectedEnds, expectedErrors, expectedEdges, expectedCombines int) {
	if poly.BeginCount != expectedBegins {
		t.Errorf("Expected BeginCount == %v, got %v\n",
			expectedBegins,
			poly.BeginCount)
	}
	if poly.VertexCount != expectedVertices {
		t.Errorf("Expected VertexCount == %v, got %v\n",
			expectedVertices,
			poly.VertexCount)
	}
	if poly.EndCount != expectedEnds {
		t.Errorf("Expected EndCount == %v, got %v\n",
			expectedEnds,
			poly.EndCount)
	}
	if poly.ErrorCount != expectedErrors {
		t.Errorf("Expected ErrorCount == %v, got %v\n",
			expectedErrors,
			poly.ErrorCount)
	}
	if poly.EdgeFlagCount != expectedEdges {
		t.Errorf("Expected EdgeFlagCount == %v, got %v\n",
			expectedEdges,
			poly.EdgeFlagCount)
	}
	if poly.CombineCount != expectedCombines {
		t.Errorf("Expected CombineCount == %v, got %v\n",
			expectedCombines,
			poly.CombineCount)
	}
}
