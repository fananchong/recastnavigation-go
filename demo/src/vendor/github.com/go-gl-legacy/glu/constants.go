// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glu

import "github.com/go-gl/gl/v2.1/gl"

const (
	// TessCallback
	TESS_BEGIN_DATA     gl.GLenum = 100106
	TESS_VERTEX_DATA              = 100107
	TESS_END_DATA                 = 100108
	TESS_ERROR_DATA               = 100109
	TESS_EDGE_FLAG_DATA           = 100110
	TESS_COMBINE_DATA             = 100111

	// TessProperty
	TESS_WINDING_RULE  = 100140
	TESS_BOUNDARY_ONLY = 100141
	TESS_TOLERANCE     = 100142

	// TessWinding
	TESS_WINDING_ODD         = 100130
	TESS_WINDING_NONZERO     = 100131
	TESS_WINDING_POSITIVE    = 100132
	TESS_WINDING_NEGATIVE    = 100133
	TESS_WINDING_ABS_GEQ_TWO = 100134
)
