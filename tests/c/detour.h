#pragma once

#include "Detour/Include/DetourNavMesh.h"
#include "Detour/Include/DetourNavMeshQuery.h"
#include "DetourTileCache/Include/DetourTileCache.h"


enum PolyAreas
{
    POLYAREA_GROUND = 0,
    POLYAREA_WATER = 1,
    POLYAREA_ROAD = 2,
    POLYAREA_DOOR = 3,
    POLYAREA_GRASS = 4,
    POLYAREA_JUMP = 5,
};

enum PolyFlags
{
    POLYFLAGS_WALK = 0x01,      // Ability to walk (ground, grass, road)
    POLYFLAGS_SWIM = 0x02,      // Ability to swim (water).
    POLYFLAGS_DOOR = 0x04,      // Ability to move through doors.
    POLYFLAGS_JUMP = 0x08,      // Ability to jump.
    POLYFLAGS_DISABLED = 0x10,  // Disabled polygon
    POLYFLAGS_ALL = 0xffff      // All abilities.
};

dtNavMesh* LoadStaticMesh(const char*path, int& errCode);
dtNavMeshQuery* CreateQuery(dtNavMesh* mesh, int maxNode);
dtNavMesh* LoadDynamicMesh(const char*path, int& errCode);

dtStatus FindRandomPoint(const dtNavMeshQuery *query,
    const dtQueryFilter* filter, float(*frand)(),
    dtPolyRef* randomRef, float* randomPt);