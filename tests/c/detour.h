#pragma once

#include "Detour/Include/DetourNavMesh.h"
#include "Detour/Include/DetourNavMeshQuery.h"

dtNavMesh* LoadStaticMesh(const char*path, int& errCode);
dtNavMeshQuery* CreateQuery(dtNavMesh* mesh, int maxNode);