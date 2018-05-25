package main

import (
	"fmt"
	"unsafe"
)

/*
#cgo LDFLAGS: -L./ -lNavMesh
#include "sdk.h"
#include <stdlib.h>
int GetInt() {
	return 0;
}
*/
import "C"

var (
	p unsafe.Pointer
)

func init() {
	p = C.CreateNavMeshSDK()
	path := C.CString("navmesh.data")
	n := C.Load(path, p)
	C.free(unsafe.Pointer(path))
	if n != 0 {
		fmt.Println("load failed")
		return
	}
}

func FindPath(start, end, ptlst []float32, ptCount *int, maxPolys int) {
	*ptCount = 0
	C.FindPath(p,
		(*C.float)(unsafe.Pointer(&start[0])),
		(*C.float)(unsafe.Pointer(&end[0])),
		(*C.float)(unsafe.Pointer(&ptlst[0])),
		(*C.int)(unsafe.Pointer(ptCount)),
		C.int(maxPolys))
}
