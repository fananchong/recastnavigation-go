package mgl

import (
	"context"
	"fmt"
	"math"
	"mesh"
	"unsafe"

	"github.com/go-gl-legacy/glu"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	g_tex GLCheckerTexture
)

const (
	LButtonUp = 0x1
)

func InitSDL() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
}

func QuitSDL() {
	sdl.Quit()
}

type ISceneExtra interface {
	OnEvent(evt sdl.Event)
	BeforeDraw()
	Draw()
	OnClick(*[3]float64)
}

type Scene struct {
	w                 *sdl.Window
	glCtx             sdl.GLContext
	mesh              *mesh.MeshLoaderObj
	chunkyMesh        *mesh.ChunkyTriMesh
	cameraEulers      [2]float32
	cameraPos         [3]float64
	camr              float64
	mousePos          [2]int32
	origMousePos      [2]int32
	origCameraEulers  [2]float32
	moveFront         float32
	moveBack          float32
	moveLeft          float32
	moveRight         float32
	moveUp            float32
	moveDown          float32
	scrollZoom        float32
	rotate            bool
	movedDuringRotate bool
	prevFrameTime     uint32
	dt                float32
	rayStart          [3]float64
	rayEnd            [3]float64
	viewport          [4]int32
	projectionMatrix  [16]float64
	modelviewMatrix   [16]float64

	iExtra ISceneExtra

	mask uint32
}

func NewScene(title string, width, height int32, meshPath string) (*Scene, error) {
	s := &Scene{}

	var err error

	s.w, _, err = sdl.CreateWindowAndRenderer(width, height, sdl.WINDOW_OPENGL)
	if err != nil {
		s.Destroy()
		return nil, err
	}

	s.w.SetTitle(title)
	s.w.SetPosition(sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED)
	s.glCtx, err = s.w.GLCreateContext()
	if err != nil {
		s.Destroy()
		return nil, err
	}

	if err = gl.Init(); err != nil {
		s.Destroy()
		return nil, err
	}

	// Enable depth buffer.
	sdl.GLSetAttribute(sdl.GL_DOUBLEBUFFER, 1)
	sdl.GLSetAttribute(sdl.GL_DEPTH_SIZE, 24)

	// Set color channel depth.
	sdl.GLSetAttribute(sdl.GL_RED_SIZE, 8)
	sdl.GLSetAttribute(sdl.GL_GREEN_SIZE, 8)
	sdl.GLSetAttribute(sdl.GL_BLUE_SIZE, 8)
	sdl.GLSetAttribute(sdl.GL_ALPHA_SIZE, 8)

	// 4x MSAA.
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 1)
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 4)

	s.mesh = mesh.NewMeshLoaderObj()

	err = s.mesh.Load(meshPath)
	if err != nil {
		s.Destroy()
		return nil, err
	}

	s.chunkyMesh = mesh.NewChunkyTriMesh()

	verts, _ := s.mesh.GetVertexs()
	tris, ntris := s.mesh.GetTriangles()
	s.chunkyMesh.CreateChunkyTriMesh(verts, tris, ntris, 256)

	bmin, bmax := s.mesh.GetAABB()
	s.camr = math.Sqrt(sqr(bmax[0]-bmin[0])+sqr(bmax[1]-bmin[1])+sqr(bmax[2]-bmin[2])) / 2
	s.cameraPos[0] = float64((bmax[0]+bmin[0])/2) + s.camr
	s.cameraPos[1] = float64((bmax[1]+bmin[1])/2) + s.camr
	s.cameraPos[2] = float64((bmax[2]+bmin[2])/2) + s.camr
	s.camr *= 3

	s.cameraEulers[0] = 45
	s.cameraEulers[1] = -45

	gl.Fogf(gl.FOG_START, float32(s.camr*0.2))
	gl.Fogf(gl.FOG_END, float32(s.camr*1.25))

	// Fog.
	var fogColor = [4]float32{0.32, 0.31, 0.30, 1.0}
	gl.Enable(gl.FOG)
	gl.Fogi(gl.FOG_MODE, gl.LINEAR)
	gl.Fogf(gl.FOG_START, float32(s.camr*0.1))
	gl.Fogf(gl.FOG_END, float32(s.camr*1.25))
	gl.Fogfv(gl.FOG_COLOR, (*float32)(unsafe.Pointer(&fogColor)))

	gl.Enable(gl.CULL_FACE)
	gl.DepthFunc(gl.LEQUAL)

	return s, nil
}

func (s *Scene) SetExtraInterface(i ISceneExtra) {
	s.iExtra = i
}

func (s *Scene) Run(ctx context.Context) {
	s.prevFrameTime = sdl.GetTicks()
	var event sdl.Event
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch evt := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.KeyboardEvent:
				s.OnKeyboardDown(evt)
			case *sdl.MouseWheelEvent:
				s.OnMouseWheel(evt)
			case *sdl.MouseButtonEvent:
				s.OnMouseButton(evt)
			case *sdl.MouseMotionEvent:
				s.OnMouseMotion(evt)
			}
			if s.iExtra != nil {
				s.iExtra.OnEvent(event)
			}
		}

		time := sdl.GetTicks()
		s.dt = float32(time-s.prevFrameTime) / 1000
		s.prevFrameTime = time

		s.BeforeDraw()

		s.SetCamera()
		s.DrawMesh()
		s.Draw()
		s.w.GLSwap()
		sdl.Delay(25)

		s.mask = 0
	}
}

func (s *Scene) Destroy() {
	if s.glCtx != nil {
		sdl.GLDeleteContext(s.glCtx)
		s.glCtx = nil
	}
	if s.w != nil {
		s.w.Destroy()
		s.w = nil
	}
}

func (s *Scene) OnKeyboardDown(evt *sdl.KeyboardEvent) {

}

func (s *Scene) OnMouseWheel(evt *sdl.MouseWheelEvent) {
	if evt.Y < 0 {
		s.scrollZoom += 2.0
	} else {
		s.scrollZoom -= 2.0
	}
}

func (s *Scene) OnMouseButton(evt *sdl.MouseButtonEvent) {
	if evt.Type == sdl.MOUSEBUTTONDOWN {
		if evt.Button == sdl.BUTTON_RIGHT {
			s.rotate = true
			s.movedDuringRotate = false
			s.origMousePos[0] = s.mousePos[0]
			s.origMousePos[1] = s.mousePos[1]
			s.origCameraEulers[0] = s.cameraEulers[0]
			s.origCameraEulers[1] = s.cameraEulers[1]
		}
	} else if evt.Type == sdl.MOUSEBUTTONUP {
		if evt.Button == sdl.BUTTON_RIGHT {
			s.rotate = false
		} else if evt.Button == sdl.BUTTON_LEFT {
			s.mask |= LButtonUp
		}
	}
}

func (s *Scene) OnMouseMotion(evt *sdl.MouseMotionEvent) {
	_, h := s.w.GetSize()
	s.mousePos[0] = evt.X
	s.mousePos[1] = h - 1 - evt.Y

	if s.rotate {
		dx := float32(s.mousePos[0] - s.origMousePos[0])
		dy := float32(s.mousePos[1] - s.origMousePos[1])
		s.cameraEulers[0] = s.origCameraEulers[0] - dy*0.25
		s.cameraEulers[1] = s.origCameraEulers[1] + dx*0.25
		if dx*dx+dy*dy > 3*3 {
			s.movedDuringRotate = true
		}
	}
}

func (s *Scene) BeforeDraw() {
	var x, y, z float64
	x, y, z = glu.UnProject(float64(s.mousePos[0]), float64(s.mousePos[1]), 0, &s.modelviewMatrix, &s.projectionMatrix, &s.viewport)
	s.rayStart[0], s.rayStart[1], s.rayStart[2] = x, y, z

	x, y, z = glu.UnProject(float64(s.mousePos[0]), float64(s.mousePos[1]), 1, &s.modelviewMatrix, &s.projectionMatrix, &s.viewport)
	s.rayEnd[0], s.rayEnd[1], s.rayEnd[2] = x, y, z

	if s.mask&LButtonUp > 0 {
		if hit, hitTime := s.raycastMesh(&s.rayStart, &s.rayEnd); hit {
			var pos [3]float64
			pos[0] = s.rayStart[0] + (s.rayEnd[0]-s.rayStart[0])*float64(hitTime)
			pos[1] = s.rayStart[1] + (s.rayEnd[1]-s.rayStart[1])*float64(hitTime)
			pos[2] = s.rayStart[2] + (s.rayEnd[2]-s.rayStart[2])*float64(hitTime)
			if s.iExtra != nil {
				s.iExtra.OnClick(&pos)
			}
		}
	}

	if s.iExtra != nil {
		s.iExtra.BeforeDraw()
	}
}

func (s *Scene) DrawMesh() {
	if s.mesh == nil {
		return
	}
	texScale := float32(0.333)
	gl.Enable(gl.FOG)

	verts, _ := s.mesh.GetVertexs()
	tris, triCount := s.mesh.GetTriangles()
	normals := s.mesh.GetNormals()

	walkableThr := float32(math.Cos(45.0 / 180 * math.Pi))

	var uva, uvb, uvc [2]float32

	Texture(true)

	unwalkable := DuRGBA(192, 128, 0, 255)
	Begin(gl.TRIANGLES, 1.0)
	for i := 0; i < triCount*3; i += 3 {
		norm := normals[i : i+3]
		var color uint32
		a := uint8(220 * (2 + norm[0] + norm[1]) / 4)
		if norm[1] < walkableThr {
			color = DuLerpCol(DuRGBA(uint32(a), uint32(a), uint32(a), 255), unwalkable, 64)
		} else {
			color = DuRGBA(uint32(a), uint32(a), uint32(a), 255)
		}
		va := verts[tris[i+0]*3 : tris[i+0]*3+3]
		vb := verts[tris[i+1]*3 : tris[i+1]*3+3]
		vc := verts[tris[i+2]*3 : tris[i+2]*3+3]
		ax := 0
		ay := 0
		if math.Abs(float64(norm[1])) > math.Abs(float64(norm[ax])) {
			ax = 1
		}
		if math.Abs(float64(norm[2])) > math.Abs(float64(norm[ax])) {
			ax = 2
		}
		ax = (1 << uint32(ax)) & 3
		ay = (1 << uint32(ax)) & 3

		uva[0] = va[ax] * texScale
		uva[1] = va[ay] * texScale
		uvb[0] = vb[ax] * texScale
		uvb[1] = vb[ay] * texScale
		uvc[0] = vc[ax] * texScale
		uvc[1] = vc[ay] * texScale

		Vertex(va, color, uva)
		Vertex(vb, color, uvb)
		Vertex(vc, color, uvc)
	}

	End()
	Texture(false)

	gl.Disable(gl.FOG)
	gl.Enable(gl.DEPTH_TEST)
}

func (s *Scene) Draw() {
	if s.iExtra != nil {
		s.iExtra.Draw()
	}
}

func (s *Scene) SetCamera() {
	width, height := s.w.GetSize()
	// Set the viewport.
	gl.Viewport(0, 0, width, height)
	// var viewport [4]int32
	gl.GetIntegerv(gl.VIEWPORT, (*int32)(unsafe.Pointer(&s.viewport)))

	// Clear the screen
	gl.ClearColor(0.3, 0.3, 0.32, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.TEXTURE_2D)
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.POINT_SMOOTH)

	// Compute the projection matrix.
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	glu.Perspective(50.0, float64(width)/float64(height), 1.0, s.camr)
	// var projectionMatrix [16]float64
	gl.GetDoublev(gl.PROJECTION_MATRIX, (*float64)(unsafe.Pointer(&s.projectionMatrix)))

	// Compute the modelview matrix.
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	gl.Rotatef(s.cameraEulers[0], 1, 0, 0)
	gl.Rotatef(s.cameraEulers[1], 0, 1, 0)
	gl.Translatef(float32(-s.cameraPos[0]), float32(-s.cameraPos[1]), float32(-s.cameraPos[2]))
	// var modelviewMatrix [16]float64
	gl.GetDoublev(gl.MODELVIEW_MATRIX, (*float64)(unsafe.Pointer(&s.modelviewMatrix)))

	keystate := sdl.GetKeyboardState()
	s.moveFront = s.moveUtil(s.moveFront, keystate, sdl.SCANCODE_W)
	s.moveLeft = s.moveUtil(s.moveLeft, keystate, sdl.SCANCODE_A)
	s.moveBack = s.moveUtil(s.moveBack, keystate, sdl.SCANCODE_S)
	s.moveRight = s.moveUtil(s.moveRight, keystate, sdl.SCANCODE_D)
	s.moveUp = s.moveUtil(s.moveUp, keystate, sdl.SCANCODE_Q)
	s.moveDown = s.moveUtil(s.moveDown, keystate, sdl.SCANCODE_E)

	var keybSpeed float32 = 22.0
	if sdl.GetModState()&sdl.KMOD_SHIFT > 0 {
		keybSpeed *= 4
	}
	movex := (s.moveRight - s.moveLeft) * keybSpeed * s.dt
	movey := (s.moveBack-s.moveFront)*keybSpeed*s.dt + s.scrollZoom*2.0
	s.scrollZoom = 0

	s.cameraPos[0] += float64(movex) * s.modelviewMatrix[0]
	s.cameraPos[1] += float64(movex) * s.modelviewMatrix[4]
	s.cameraPos[2] += float64(movex) * s.modelviewMatrix[8]
	s.cameraPos[0] += float64(movey) * s.modelviewMatrix[2]
	s.cameraPos[1] += float64(movey) * s.modelviewMatrix[6]
	s.cameraPos[2] += float64(movey) * s.modelviewMatrix[10]

	s.cameraPos[1] += float64((s.moveUp-s.moveDown)*keybSpeed) * float64(s.dt)
}

func sqr(v float32) float64 {
	return float64(v * v)
}

func (s *Scene) moveUtil(v float32, keystate []uint8, key int) float32 {
	var delta float32
	if keystate[key] > 0 {
		delta = s.dt * 4 * 1
	} else {
		delta = s.dt * 4 * -1
	}
	v = v + delta
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return v
}

func (s *Scene) raycastMesh(src, dst *[3]float64) (bool, float32) {
	bmin, bmax := s.mesh.GetAABB()
	// var dir = [3]float64{dst[0] - src[0], dst[1] - src[1], dst[2] - src[2]}
	ok, btmin, btmax := isectSegAABB(src, dst, bmin, bmax)
	if !ok {
		return false, 0
	}
	var p = [2]float32{float32(src[0] + (dst[0]-src[0])*btmin), float32(src[2] + (dst[2]-src[2])*btmin)}
	var q = [2]float32{float32(src[0] + (dst[0]-src[0])*btmax), float32(src[2] + (dst[2]-src[2])*btmax)}

	var cid [512]int
	ncid := s.chunkyMesh.GetChunksOverlappingSegment(&p, &q, cid[0:], 512)
	if ncid == 0 {
		return false, 0
	}

	fmt.Println("cid:", cid[:ncid])

	//TODO:目前到这里都没问题，明天继续检查下面，看是什么问题导致的

	tmin := float32(1.0)
	hit := false
	verts, _ := s.mesh.GetVertexs()

	for i := 0; i < ncid; i++ {
		node := s.chunkyMesh.GetNode(cid[i])
		tris := s.chunkyMesh.GetTris(node.I() * 3)
		ntris := node.N()

		for j := 0; j < ntris*3; j += 3 {
			if ok, t := intersectSegmentTriangle(src, dst, verts[tris[j]*3:], verts[tris[j+1]*3:], verts[tris[j+2]*3:]); ok {
				if t < tmin {
					tmin = t
				}
				hit = true
			}
		}
	}

	return hit, tmin
}
