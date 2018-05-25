package mesh

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sync/atomic"
)

type MeshLoaderObj struct {
	fileName  string
	scale     float32
	verts     []float32
	vertCount int
	tris      []int
	triCount  int
	normals   []float32

	bmin [3]float32
	bmax [3]float32

	loadOk int32
}

func NewMeshLoaderObj() *MeshLoaderObj {
	return &MeshLoaderObj{
		scale: 1.0,
	}
}

func (ml *MeshLoaderObj) Load(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	var row []byte
	var x, y, z float32
	for len(data) > 0 {
		row, data = ml.parseRow(data)
		// Skip comments
		if row[0] == '#' {
			continue
		}
		if row[0] == 'v' && row[1] != 'n' && row[1] != 't' {
			// vertex pos
			fmt.Sscanf(string(row[1:]), "%f %f %f", &x, &y, &z)
			ml.addVertex(x, y, z)
		}
		if row[0] == 'f' {
			face := ml.parseFace(row[1:], 32, ml.vertCount)
			for i := 2; i < len(face); i++ {
				a := face[0]
				b := face[i-1]
				c := face[i]
				if a < 0 || a >= ml.vertCount || b < 0 || b >= ml.vertCount || c < 0 || c >= ml.vertCount {
					continue
				}
				ml.addTriangle(a, b, c)
			}
		}
	}

	for i := 0; i < ml.triCount*3; i += 3 {
		v0 := ml.verts[ml.tris[i]*3 : ml.tris[i]*3+3]
		v1 := ml.verts[ml.tris[i+1]*3 : ml.tris[i+1]*3+3]
		v2 := ml.verts[ml.tris[i+2]*3 : ml.tris[i+2]*3+3]
		var e0, e1 [3]float32
		for j := 0; j < 3; j++ {
			e0[j] = v1[j] - v0[j]
			e1[j] = v2[j] - v0[j]
		}
		n0 := e0[1]*e1[2] - e0[2]*e1[1]
		n1 := e0[2]*e1[0] - e0[0]*e1[2]
		n2 := e0[0]*e1[1] - e0[1]*e1[0]
		d := math.Sqrt(float64(n0*n0 + n1*n1 + n2*n2))
		if d > 0 {
			d = 1.0 / d
			n0 *= float32(d)
			n1 *= float32(d)
			n2 *= float32(d)
		}
		ml.normals = append(ml.normals, n0, n1, n2)
	}

	ml.fileName = fileName

	ml.calcBounds()

	fmt.Println("AABB:", ml.bmin, ml.bmax)
	// ml.Save()
	atomic.StoreInt32(&ml.loadOk, 1)
	return nil
}

func (ml *MeshLoaderObj) calcBounds() {
	copy(ml.bmin[0:], ml.verts)
	copy(ml.bmax[0:], ml.verts)
	for i := 1; i < ml.vertCount; i++ {
		v := ml.verts[i*3 : i*3+3]
		ml.bmin[0] = float32(math.Min(float64(ml.bmin[0]), float64(v[0])))
		ml.bmin[1] = float32(math.Min(float64(ml.bmin[1]), float64(v[1])))
		ml.bmin[2] = float32(math.Min(float64(ml.bmin[2]), float64(v[2])))
		ml.bmax[0] = float32(math.Max(float64(ml.bmax[0]), float64(v[0])))
		ml.bmax[1] = float32(math.Max(float64(ml.bmax[1]), float64(v[1])))
		ml.bmax[2] = float32(math.Max(float64(ml.bmax[2]), float64(v[2])))
	}
}

func (ml *MeshLoaderObj) addVertex(x, y, z float32) {
	x, y, z = x*ml.scale, y*ml.scale, z*ml.scale
	ml.verts = append(ml.verts, x, y, z)
	ml.vertCount++
}

func (ml *MeshLoaderObj) addTriangle(a, b, c int) {
	ml.tris = append(ml.tris, a, b, c)
	ml.triCount++
}

func (ml *MeshLoaderObj) parseRow(data []byte) ([]byte, []byte) {
	var row []byte
	start := true
	done := false
	for !done && len(data) > 0 {
		c := data[0]
		data = data[1:]
		switch c {
		case '\\':
		case '\n':
			if !start {
				done = true
			}
		case '\r':
		case '\t':
			start = false
			row = append(row, c)
		case ' ':
			if !start {
				start = false
				row = append(row, c)
			}
		default:
			start = false
			row = append(row, c)
		}
	}
	return row, data
}

func (ml *MeshLoaderObj) parseFace(row []byte, n, vcnt int) []int {
	// fmt.Println(string(row))
	var data []int
	for len(row) > 0 {
		for len(row) > 0 && (row[0] == ' ' || row[0] == '\t') {
			row = row[1:]
		}
		s := row
		for len(row) > 0 && row[0] != ' ' && row[0] != '\t' {
			if row[0] == '/' {
				row[0] = 0
			}
			row = row[1:]
		}
		if len(s) == 0 || s[0] == 0 {
			continue
		}
		var vi int
		// fmt.Println(string(s))
		fmt.Sscanf(string(s), "%d", &vi)
		if vi < 0 {
			data = append(data, vi+vcnt)
		} else {
			data = append(data, vi-1)
		}
		if len(data) >= n {
			return data
		}
	}
	return data
}

func (ml *MeshLoaderObj) Save() {
	f, _ := os.OpenFile("verify.txt", os.O_CREATE|os.O_TRUNC, 0)
	defer f.Close()
	f.WriteString(fmt.Sprintf("vertex count: %v\n", ml.vertCount))
	f.WriteString(fmt.Sprintf("triangle count: %v\n", ml.triCount))

	for i := 0; i < ml.vertCount; i++ {
		f.WriteString(fmt.Sprintf("%v, %v, %v\n", ml.verts[i*3], ml.verts[i*3+1], ml.verts[i*3+2]))
	}

	f.WriteString("\n\n")
	for i := 0; i < ml.triCount; i++ {
		f.WriteString(fmt.Sprintf("%v, %v, %v\n", ml.tris[i*3], ml.tris[i*3+1], ml.tris[i*3+2]))
	}

	f.WriteString("\n\n")
	for i := 0; i < ml.triCount; i++ {
		f.WriteString(fmt.Sprintf("%v, %v, %v\n", ml.normals[i*3], ml.normals[i*3+1], ml.normals[i*3+2]))
	}
}

func (ml *MeshLoaderObj) LoadDone() bool {
	return atomic.LoadInt32(&ml.loadOk) > 0
}

func (ml *MeshLoaderObj) GetVertexs() ([]float32, int) {
	return ml.verts, ml.vertCount
}

func (ml *MeshLoaderObj) GetTriangles() ([]int, int) {
	return ml.tris, ml.triCount
}

func (ml *MeshLoaderObj) GetNormals() []float32 {
	return ml.normals
}

func (ml *MeshLoaderObj) GetAABB() (*[3]float32, *[3]float32) {
	return &ml.bmin, &ml.bmax
}
