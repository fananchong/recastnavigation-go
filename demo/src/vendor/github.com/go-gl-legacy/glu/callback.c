// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "callback.h"

void setGluTessCallback(GLUtesselator *tess, GLenum which)
{
        _GLUfuncptr func;
        switch (which) {
                case GLU_TESS_BEGIN_DATA:
                        func = (_GLUfuncptr)goTessBeginData;
                        break;
                case GLU_TESS_VERTEX_DATA:
                        func = (_GLUfuncptr)goTessVertexData;
                        break;
                case GLU_TESS_END_DATA:
                        func = (_GLUfuncptr)goTessEndData;
                        break;
                case GLU_TESS_ERROR_DATA:
                        func = (_GLUfuncptr)goTessErrorData;
                        break;
                case GLU_TESS_EDGE_FLAG_DATA:
                        func = (_GLUfuncptr)goTessEdgeFlagData;
                        break;
                case GLU_TESS_COMBINE_DATA:
                        func = (_GLUfuncptr)goTessCombineData;
                        break;
        }

        gluTessCallback(tess, which, func);
}

