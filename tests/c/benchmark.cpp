#include <chrono>
#include <cassert>
#include "detour.h"
#include <stdio.h>

long long get_tick_count(void)
{
    typedef std::chrono::time_point<std::chrono::system_clock, std::chrono::nanoseconds> nanoClock_type;
    nanoClock_type tp = std::chrono::time_point_cast<std::chrono::nanoseconds>(std::chrono::system_clock::now());
    return tp.time_since_epoch().count();
}


const int RAND_MAX_COUNT = 200000;
float randPosValue[RAND_MAX_COUNT * 4];
int randPosIndex = 0;
inline void getPos(dtPolyRef& ref, float pos[3])
{
    ref = dtPolyRef(randPosValue[randPosIndex * 4 + 0]);
    pos[0] = randPosValue[randPosIndex * 4 + 1];
    pos[1] = randPosValue[randPosIndex * 4 + 2];
    pos[2] = randPosValue[randPosIndex * 4 + 3];
    randPosIndex++;
}

const int PATH_MAX_NODE = 2048;
const char* MESH_FILE = "../../nav_test.obj.tile.bin";

int main() {
    FILE* f1 = fopen("../../randpos.bin", "rb");
    fread(randPosValue, RAND_MAX_COUNT * 4 * sizeof(float), 1, f1);
    fclose(f1);


    int errCode;
    auto mesh = LoadStaticMesh(MESH_FILE, errCode);
    assert(errCode == 0);
    auto query = CreateQuery(mesh, 2048);
    assert(query != nullptr);
    auto filter = dtQueryFilter();

    int count = RAND_MAX_COUNT / 2;
    printf("total count: %d\n", count);

    auto t1 = get_tick_count();
    for (int i = 0; i < count; i++)
    {
        dtStatus stat;
        float halfExtents[3] = { 2, 4, 2 };
        float startPos[3] = { 0,0,0 };
        float endPos[3] = { 0,0,0 };
        dtPolyRef startRef = 0;
        dtPolyRef endRef = 0;
        getPos(startRef, startPos);
        getPos(endRef, endPos);
        dtPolyRef path[PATH_MAX_NODE];
        int pathCount = 0;
        stat = query->findPath(startRef, endRef, startPos, endPos, &filter, path, &pathCount, PATH_MAX_NODE);
        assert(dtStatusSucceed(stat));
    }
    auto t2 = get_tick_count();
    printf("findPath cost:        %20lldns %20.3fns/op %20.3fms/op\n",
        t2 - t1, float(t2 - t1) / count, float(t2 - t1) / count / 1000000);

    randPosIndex = 0;
    t1 = get_tick_count();
    for (int i = 0; i < count; i++)
    {
        dtStatus stat;
        float halfExtents[3] = { 2, 4, 2 };
        float startPos[3] = { 0,0,0 };
        float endPos[3] = { 0,0,0 };
        dtPolyRef startRef = 0;
        dtPolyRef endRef = 0;
        getPos(startRef, startPos);
        getPos(endRef, endPos);
        float resultPos[3] = { 0,0,0 };
        dtPolyRef visited[PATH_MAX_NODE];
        int visitedCount = 0;
        bool bHit = false;
        stat = query->moveAlongSurface(startRef, startPos, endPos, &filter, resultPos, visited, &visitedCount, PATH_MAX_NODE, bHit);
        assert(dtStatusSucceed(stat));
    }
    t2 = get_tick_count();
    printf("moveAlongSurface cost:%20lldns %20.3fns/op %20.3fms/op\n",
        t2 - t1, float(t2 - t1) / count, float(t2 - t1) / count / 1000000);
}
