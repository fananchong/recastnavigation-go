package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

const (
	maxPolys = 256
)

type User struct {
	pos     []float32
	destPos []float32
	ptlst   [256 * 3]float32
	ptCount int
}

func (u *User) SetPos(pos *[3]float64) {
	if sdl.GetModState()&sdl.KMOD_CTRL > 0 {
		if u.pos == nil {
			u.pos = make([]float32, 3)
		}
		u.pos[0], u.pos[1], u.pos[2] = float32(pos[0]), float32(pos[1]), float32(pos[2])
	} else if u.pos != nil {
		if u.destPos == nil {
			u.destPos = make([]float32, 3)
		}
		u.destPos[0], u.destPos[1], u.destPos[2] = float32(pos[0]), float32(pos[1]), float32(pos[2])

		// FindPath(u.pos, u.destPos, u.ptlst[0:], &u.ptCount, maxPolys)
		GoFindPath(u.pos, u.destPos, u.ptlst[0:], &u.ptCount, maxPolys)
	}
}
