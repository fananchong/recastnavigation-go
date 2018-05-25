package main

import (
	"context"
	"fmt"
	"mgl"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/veandco/go-sdl2/sdl"
)

type Scene struct {
	s *mgl.Scene

	u User
}

func NewScene() *Scene {
	s, err := mgl.NewScene("recast", 1840, 1000, "mine.obj")
	if err != nil {
		panic(err)
	}
	ss := &Scene{
		s: s,
	}

	s.SetExtraInterface(ss)

	return ss
}

func (s *Scene) OnEvent(event sdl.Event) {
	// switch evt := event.(type) {
	// case *sdl.MouseButtonEvent:
	// 	if evt.Type == sdl.MOUSEBUTTONDOWN {

	// 	}
	// }
}

func (s *Scene) BeforeDraw() {

}

func (s *Scene) Draw() {
	if s.u.pos != nil {
		// mgl.Begin(gl.POINTS, 50)
		color := mgl.DuRGBA(255, 0, 0, 255)
		// mgl.Vertex2(s.u.pos[0], s.u.pos[1], s.u.pos[2], color)
		mgl.DrawCylinder(s.u.pos[0]-0.5, s.u.pos[1], s.u.pos[2]-0.5, s.u.pos[0]+0.5, s.u.pos[1]+2, s.u.pos[2]+0.5, color)
		// mgl.End()
	}

	if s.u.ptCount > 0 {
		spathCol := mgl.DuRGBA(64, 16, 0, 220)
		s.DrawPath(s.u.ptlst[:], s.u.ptCount, spathCol, 4)
	}
	if s.u.goptCount > 0 {
		spathCol := mgl.DuRGBA(64, 255, 0, 220)
		s.DrawPath(s.u.goptlst[:], s.u.goptCount, spathCol, 2)
	}
}

func (s *Scene) DrawPath(ptlst []float32, ptCount int, spathCol uint32, blod float32) {

	mgl.Begin(gl.LINES, blod)
	for i := 0; i < ptCount-1; i++ {
		mgl.Vertex2(ptlst[i*3], ptlst[i*3+1]+0.4, ptlst[i*3+2], spathCol)
		mgl.Vertex2(ptlst[(i+1)*3], ptlst[(i+1)*3+1]+0.4, ptlst[(i+1)*3+2], spathCol)
	}
	mgl.End()

	mgl.Begin(gl.POINTS, 6)
	for i := 0; i < ptCount; i++ {
		mgl.Vertex2(ptlst[i*3], ptlst[i*3+1]+0.4, ptlst[i*3+2], spathCol)
	}
	mgl.End()
}

func (s *Scene) Run(ctx context.Context) {
	s.s.Run(ctx)
}

func (s *Scene) OnClick(pos *[3]float64) {
	fmt.Println("clickpos:", pos[0], pos[1], pos[2])
	s.u.SetPos(pos)
}
