package main

import (
	"fmt"
	"time"
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

func main() {
	fmt.Println("test mesh")

	p := C.CreateNavMeshSDK()
	path := C.CString("navmesh.data")
	n := C.Load(path, p)
	C.free(unsafe.Pointer(path))
	if n != 0 {
		fmt.Println("load failed")
		return
	}

	var start, end [3]float32
	start[0] = 310.504303
	start[1] = 0.0813598633
	start[2] = 235.351929

	end[0] = 137.467896
	end[1] = 0.0813598633
	end[2] = 295.069305

	maxPolys := 256
	var ptlst [256 * 3]float32
	var ptCount int

	startTime := time.Now()
	for i := 0; i < 10000; i++ {
		ptCount = 0
		C.FindPath(p,
			(*C.float)(unsafe.Pointer(&start)),
			(*C.float)(unsafe.Pointer(&end)),
			(*C.float)(unsafe.Pointer(&ptlst)),
			(*C.int)(unsafe.Pointer(&ptCount)),
			C.int(maxPolys))
	}

	fmt.Println("use time:", time.Since(startTime))

	fmt.Println("ave time:", float32(time.Since(startTime))/10000)

	// fmt.Println("point count:", ptCount)
	// for i := 0; i < ptCount; i++ {
	// 	fmt.Printf("%v, %v, %v\n", ptlst[i*3], ptlst[i*3+1], ptlst[i*3+2])
	// }
	// fmt.Println(n)
}
