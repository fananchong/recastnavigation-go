#include "detour.h"
#include <cassert>
#include <string>
#include <vector>
#include <time.h>

const int RAND_MAX_COUNT = 20000000;
float randValue[RAND_MAX_COUNT];
int randIndex = 0;
inline float frand()
{
    auto v = randValue[randIndex];
    randIndex++;
    return v;
}

const int PATH_MAX_NODE = 2048;

const int OUT_MAX_COUNT = 10000;
float outValue[OUT_MAX_COUNT];
int outIndex = 0;


float myrand() {
    return (float)rand() / (float)RAND_MAX;
}

const char* MESH_FILE = "../../nav_test.obj.tile.bin";
const char* MESH_FILE_CACHE = "../../nav_test.obj.tilecache.bin";
int main(int argn, char* argv[]) {
    srand((unsigned int)(time(0)));

    int errCode;
    dtNavMesh* mesh;
    if (argn >= 3 && argv[2] == std::string("1")) {
        mesh = LoadDynamicMesh(MESH_FILE_CACHE, errCode);
        outValue[outIndex++] = 1;
    }
    else {
        mesh = LoadStaticMesh(MESH_FILE, errCode);
        outValue[outIndex++] = 0;
    }
    assert(errCode == 0);
    auto query = CreateQuery(mesh, 2048);
    assert(query != nullptr);
    auto filter = dtQueryFilter();


    if (argn > 1 && argv[1] == std::string("rand")) {
        FILE* f = fopen("../../rand.bin", "wb");
        for (int i = 0; i < RAND_MAX_COUNT;i++) {
            float v = (float)rand() / (float)RAND_MAX;
            fwrite(&v, sizeof(float), 1, f);
        }
        fclose(f);
    }
    else if (argn > 1 && argv[1] == std::string("randpos")) {
        void randomPos(dtNavMesh* mesh, dtNavMeshQuery* query, dtQueryFilter* filter);
        randomPos(mesh, query, &filter);
    }
    else {
        FILE* f1 = fopen("../../rand.bin", "rb");
        fread(randValue, RAND_MAX_COUNT * sizeof(float), 1, f1);
        fclose(f1);
        int test(dtNavMesh* mesh, dtNavMeshQuery* query, dtQueryFilter* filter);
        test(mesh, query, &filter);
        FILE* f2 = fopen("../../result.bin", "wb");
        fwrite(outValue, outIndex * sizeof(float), 1, f2);
        fclose(f2);
    }
    return 0;
}

void randomPos(dtNavMesh* mesh, dtNavMeshQuery* query, dtQueryFilter* filter) {
    FILE* f = fopen("../../randpos.bin", "wb");
    for (int i = 0; i < RAND_MAX_COUNT;i++) {
        float startPos[3] = { 0,0,0 };
        dtPolyRef startRef = 0;
        dtStatus stat = query->findRandomPoint(filter, myrand, &startRef, startPos);
        float tempRef = float(startRef);
        assert(dtStatusSucceed(stat));
        fwrite(&tempRef, sizeof(float), 1, f);
        fwrite(&startPos, sizeof(float) * 3, 1, f);
    }
    fclose(f);
}

int test(dtNavMesh* mesh, dtNavMeshQuery* query, dtQueryFilter* filter) {
    dtStatus stat;
    float halfExtents[3] = { 2, 4, 2 };
    float startPos[3] = { 0,0,0 };
    float endPos[3] = { 0,0,0 };
    dtPolyRef startRef = 0;
    dtPolyRef endRef = 0;

    //printf("================================================ findRandomPoint ================================================\n");
    stat = query->findRandomPoint(filter, frand, &startRef, startPos);
    assert(dtStatusSucceed(stat));
    stat = query->findRandomPoint(filter, frand, &endRef, endPos);
    assert(dtStatusSucceed(stat));
    //printf("startPos: %.2f %.2f %.2f\n", startPos[0], startPos[1], startPos[2]);
    //printf("endPos: %.2f %.2f %.2f\n", endPos[0], endPos[1], endPos[2]);
    //printf("startRef: %d\n", startRef);
    //printf("endRef: %d\n", endRef);
    //printf("\n");
    outValue[outIndex++] = (startPos[0]);
    outValue[outIndex++] = (startPos[1]);
    outValue[outIndex++] = (startPos[2]);
    outValue[outIndex++] = (endPos[0]);
    outValue[outIndex++] = (endPos[1]);
    outValue[outIndex++] = (endPos[2]);
    outValue[outIndex++] = (float(startRef));
    outValue[outIndex++] = (float(endRef));


    //printf("================================================ findNearestPoly ================================================\n");
    float tempPos[3] = { 0,0,0 };
    dtPolyRef nearestRef = 0;
    float nearestPos[3] = { 0,0,0 };
    stat = query->findNearestPoly(tempPos, halfExtents, filter, &nearestRef, nearestPos);
    assert(dtStatusSucceed(stat));
    //printf("nearestPos: %.2f %.2f %.2f\n", nearestPos[0], nearestPos[1], nearestPos[2]);
    //printf("nearestRef: %d\n", nearestRef);
    //printf("\n");
    outValue[outIndex++] = (nearestPos[0]);
    outValue[outIndex++] = (nearestPos[1]);
    outValue[outIndex++] = (nearestPos[2]);
    outValue[outIndex++] = (float(nearestRef));


    //printf("================================================ findPath ================================================\n");
    dtPolyRef path[PATH_MAX_NODE];
    int pathCount = 0;
    stat = query->findPath(startRef, endRef, startPos, endPos, filter, path, &pathCount, PATH_MAX_NODE);
    assert(dtStatusSucceed(stat));
    //printf("pathCount: %d\n", pathCount);
    outValue[outIndex++] = (float(pathCount));
    for (int i = 0; i < pathCount; i++) {
        //printf("%d\n", path[i]);
        outValue[outIndex++] = (float(path[i]));
    }
    //printf("\n");

    {
        //printf("================================================ findStraightPath # DT_STRAIGHTPATH_AREA_CROSSINGS ================================================\n");
        float straightPath[PATH_MAX_NODE * 3];
        unsigned char straightPathFlags[PATH_MAX_NODE];
        dtPolyRef straightPathRefs[PATH_MAX_NODE];
        int straightPathCount = 0;
        stat = query->findStraightPath(startPos, endPos, path, pathCount,
            straightPath, straightPathFlags, straightPathRefs,
            &straightPathCount, PATH_MAX_NODE, DT_STRAIGHTPATH_AREA_CROSSINGS);
        assert(dtStatusSucceed(stat));
        //printf("straightPathCount: %d\n", straightPathCount);
        outValue[outIndex++] = (float(straightPathCount));
        for (int i = 0; i < straightPathCount; i++) {
            //printf("straightPath: %.3f %.3f %.3f, straightPathFlags: %d, straightPathRefs: %d\n",
            //straightPath[i * 3 + 0], straightPath[i * 3 + 1], straightPath[i * 3 + 2],
            //    straightPathFlags[i], straightPathRefs[i]);
            outValue[outIndex++] = (float(straightPath[i * 3 + 0]));
            outValue[outIndex++] = (float(straightPath[i * 3 + 1]));
            outValue[outIndex++] = (float(straightPath[i * 3 + 2]));
            outValue[outIndex++] = (float(straightPathFlags[i]));
            outValue[outIndex++] = (float(straightPathRefs[i]));
        }
        //printf("\n");
    }

    {
        //printf("================================================ findStraightPath # DT_STRAIGHTPATH_ALL_CROSSINGS ================================================\n");
        float straightPath[PATH_MAX_NODE * 3];
        unsigned char straightPathFlags[PATH_MAX_NODE];
        dtPolyRef straightPathRefs[PATH_MAX_NODE];
        int straightPathCount = 0;
        stat = query->findStraightPath(startPos, endPos, path, pathCount,
            straightPath, straightPathFlags, straightPathRefs,
            &straightPathCount, PATH_MAX_NODE, DT_STRAIGHTPATH_ALL_CROSSINGS);
        assert(dtStatusSucceed(stat));
        //printf("straightPathCount: %d\n", straightPathCount);
        outValue[outIndex++] = (float(straightPathCount));
        for (int i = 0; i < straightPathCount; i++) {
            //printf("straightPath: %.3f %.3f %.3f, straightPathFlags: %d, straightPathRefs: %d\n",
            //straightPath[i * 3 + 0], straightPath[i * 3 + 1], straightPath[i * 3 + 2],
            //    straightPathFlags[i], straightPathRefs[i]);
            outValue[outIndex++] = (float(straightPath[i * 3 + 0]));
            outValue[outIndex++] = (float(straightPath[i * 3 + 1]));
            outValue[outIndex++] = (float(straightPath[i * 3 + 2]));
            outValue[outIndex++] = (float(straightPathFlags[i]));
            outValue[outIndex++] = (float(straightPathRefs[i]));
        }
        //printf("\n");
    }

    //printf("================================================ moveAlongSurface ================================================\n");
    float resultPos[3] = { 0,0,0 };
    dtPolyRef visited[PATH_MAX_NODE];
    int visitedCount = 0;
    bool bHit = false;
    stat = query->moveAlongSurface(startRef, startPos, endPos, filter, resultPos, visited, &visitedCount, PATH_MAX_NODE, bHit);
    assert(dtStatusSucceed(stat));
    //printf("resultPos: %.2f %.2f %.2f\n", resultPos[0], resultPos[1], resultPos[2]);
    //printf("bHit: %d\n", bHit);
    //printf("visitedCount: %d\n", visitedCount);
    outValue[outIndex++] = (float(resultPos[0]));
    outValue[outIndex++] = (float(resultPos[1]));
    outValue[outIndex++] = (float(resultPos[2]));
    outValue[outIndex++] = (float(bHit));
    outValue[outIndex++] = (float(visitedCount));
    for (int i = 0; i < visitedCount; i++) {
        //printf("%d\n", visited[i]);
        outValue[outIndex++] = (float(visited[i]));
    }
    //printf("\n");

    //printf("================================================ findDistanceToWall # 0================================================\n");
    float hitDist = 0;
    float hitPos[3] = { 0,0,0 };
    float hitNormal[3] = { 0,0,0 };
    stat = query->findDistanceToWall(startRef, startPos, 30, filter, &hitDist, hitPos, hitNormal);
    assert(dtStatusSucceed(stat));
    //printf("hitPos: %.2f %.2f %.2f\n", hitPos[0], hitPos[1], hitPos[2]);
    //printf("hitDist: %f\n", hitDist);
    //printf("hitNormal: %.2f %.2f %.2f\n", hitNormal[0], hitNormal[1], hitNormal[2]);
    //printf("\n");

    outValue[outIndex++] = (float(hitPos[0]));
    outValue[outIndex++] = (float(hitPos[1]));
    outValue[outIndex++] = (float(hitPos[2]));
    outValue[outIndex++] = (float(hitDist));
    outValue[outIndex++] = (float(hitNormal[0]));
    outValue[outIndex++] = (float(hitNormal[1]));
    outValue[outIndex++] = (float(hitNormal[2]));

    {
        //printf("================================================ SlicedFindPath # 0================================================\n");
        stat = query->initSlicedFindPath(startRef, endRef, startPos, endPos, filter, 0);
        assert(dtStatusInProgress(stat) || dtStatusSucceed(stat));
        for (;true;) {

            if (dtStatusInProgress(stat)) {
                int doneIters;
                stat = query->updateSlicedFindPath(4, &doneIters);
            }
            if (dtStatusSucceed(stat)) {
                dtPolyRef path[PATH_MAX_NODE];
                int pathCount;
                stat = query->finalizeSlicedFindPath(path, &pathCount, PATH_MAX_NODE);
                //printf("pathCount: %d\n", pathCount);
                outValue[outIndex++] = (float(pathCount));
                for (int i = 0; i < pathCount; i++) {
                    //printf("%d\n", path[i]);
                    outValue[outIndex++] = (float(path[i]));
                }
                break;
            }
        }
        //printf("\n");
    }

    {
        //printf("================================================ SlicedFindPath # DT_FINDPATH_ANY_ANGLE ================================================\n");
        stat = query->initSlicedFindPath(startRef, endRef, startPos, endPos, filter, DT_FINDPATH_ANY_ANGLE);
        assert(dtStatusInProgress(stat) || dtStatusSucceed(stat));
        for (;true;) {

            if (dtStatusInProgress(stat)) {
                int doneIters;
                stat = query->updateSlicedFindPath(4, &doneIters);
            }
            if (dtStatusSucceed(stat)) {
                dtPolyRef path[PATH_MAX_NODE];
                int pathCount;
                stat = query->finalizeSlicedFindPath(path, &pathCount, PATH_MAX_NODE);
                //printf("pathCount: %d\n", pathCount);
                outValue[outIndex++] = (float(pathCount));
                for (int i = 0; i < pathCount; i++) {
                    //printf("%d\n", path[i]);
                    outValue[outIndex++] = (float(path[i]));
                }
                break;
            }
        }
        //printf("\n");
    }
    return 0;
}