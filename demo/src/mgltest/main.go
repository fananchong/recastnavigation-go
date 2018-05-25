package main

import (
	"mgl"
)

func main() {
	mgl.InitSDL()
	defer mgl.QuitSDL()

	s := NewScene()
	s.Run(gCtx)
}
