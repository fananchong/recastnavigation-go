package mgl

import (
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
)

const (
	TSIZE = 64
)

type GLCheckerTexture struct {
	texId uint32
}

func (ct *GLCheckerTexture) Bind() {
	if ct.texId == 0 {
		col0 := DuRGBA(215, 215, 215, 255)
		col1 := DuRGBA(255, 255, 255, 255)
		data := make([]uint32, TSIZE*TSIZE)
		gl.GenTextures(1, &ct.texId)
		gl.BindTexture(gl.TEXTURE_2D, ct.texId)

		var level int32
		size := int32(TSIZE)
		for size > 0 {
			for y := int32(0); y < size; y++ {
				for x := int32(0); x < size; x++ {
					var col uint32
					if x == 0 || y == 0 {
						col = col0
					} else {
						col = col1
					}
					data[x+y*size] = col
				}
			}
			gl.TexImage2D(gl.TEXTURE_2D, level, gl.RGBA, size, size, 0, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&data[0]))
			size /= 2
			level++
		}
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	} else {
		gl.BindTexture(gl.TEXTURE_2D, ct.texId)
	}
}
