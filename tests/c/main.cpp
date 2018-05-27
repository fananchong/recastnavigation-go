#include "detour.h"
#include <cassert>
#include <stdio.h>
#include <stdlib.h>



float randValue[] = {
    0.001f,
    0.564f,
    0.193f,
    0.809f,
    0.585f,
    0.480f,
    0.350f,
    0.896f,
    0.823f,
    0.747f,
    0.174f,
    0.859f,
    0.711f,
    0.514f,
    0.304f,
    0.015f,
    0.091f,
    0.364f,
    0.147f,
    0.166f,
    0.989f,
    0.446f,
    0.119f,
    0.005f,
    0.009f,
    0.378f,
    0.532f,
    0.571f,
    0.602f,
    0.607f,
    0.166f,
    0.663f,
    0.451f,
    0.352f,
    0.057f,
    0.608f,
    0.783f,
    0.803f,
    0.520f,
    0.302f,
    0.876f,
    0.727f,
    0.956f,
    0.926f,
    0.539f,
    0.142f,
    0.462f,
    0.235f,
    0.862f,
    0.210f,
    0.780f,
    0.844f,
    0.997f,
    1.000f,
    0.611f,
    0.392f,
    0.266f,
    0.297f,
    0.840f,
    0.024f,
    0.376f,
    0.093f,
    0.677f,
    0.056f,
    0.009f,
    0.919f,
    0.276f,
    0.273f,
    0.588f,
    0.691f,
    0.838f,
    0.726f,
    0.485f,
    0.205f,
    0.744f,
    0.468f,
    0.458f,
    0.949f,
    0.744f,
    0.108f,
    0.599f,
    0.385f,
    0.735f,
    0.609f,
    0.572f,
    0.361f,
    0.152f,
    0.225f,
    0.425f,
    0.803f,
    0.517f,
    0.990f,
    0.752f,
    0.346f,
    0.169f,
    0.657f,
    0.492f,
    0.064f,
    0.700f,
    0.505f,
    0.147f,
    0.950f,
    0.142f,
    0.905f,
    0.693f,
    0.303f,
    0.427f,
    0.070f,
    0.967f,
    0.683f,
    0.153f,
    0.877f,
    0.822f,
    0.582f,
    0.191f,
    0.178f,
    0.817f,
    0.475f,
    0.156f,
    0.504f,
    0.732f,
    0.406f,
    0.280f,
    0.569f,
    0.682f,
    0.756f,
    0.722f,
    0.475f,
    0.123f,
    0.368f,
    0.835f,
    0.035f,
    0.517f,
    0.663f,
    0.426f,
    0.105f,
    0.949f,
    0.921f,
    0.550f,
};

int randIndex = 0;

inline float frand()
{
    auto v = randValue[randIndex];
    randIndex++;
    return v;
}

const int PATH_MAX_NODE = 2048;

int main() {
    int errCode;
    auto mesh = LoadStaticMesh("../../nav_test.obj.tile.bin", errCode);
    assert(errCode == 0);
    auto query = CreateQuery(mesh, 2048);
    assert(query != nullptr);
    auto filter = dtQueryFilter();

    dtStatus stat;
    float halfExtents[3] = { 2, 4, 2 };
    float startPos[3], endPos[3];
    dtPolyRef startRef, endRef;

    printf("================================================ findRandomPoint ================================================\n");
    stat = query->findRandomPoint(&filter, frand, &startRef, startPos);
    assert(dtStatusSucceed(stat));
    stat = query->findRandomPoint(&filter, frand, &endRef, endPos);
    assert(dtStatusSucceed(stat));
    printf("startPos: %.2f %.2f %.2f\n", startPos[0], startPos[1], startPos[2]);
    printf("endPos: %.2f %.2f %.2f\n", endPos[0], endPos[1], endPos[2]);
    printf("startRef: %d\n", startRef);
    printf("endRef: %d\n", endRef);
    printf("\n");

    printf("================================================ findNearestPoly ================================================\n");
    float tempPos[3] = { 0,0,0 };
    dtPolyRef nearestRef;
    float nearestPos[3];
    stat = query->findNearestPoly(tempPos, halfExtents, &filter, &nearestRef, nearestPos);
    assert(dtStatusSucceed(stat));
    printf("nearestPos: %.2f %.2f %.2f\n", nearestPos[0], nearestPos[1], nearestPos[2]);
    printf("nearestRef: %d\n", nearestRef);
    printf("\n");

    printf("================================================ findPath ================================================\n");
    dtPolyRef path[PATH_MAX_NODE];
    int pathCount;
    stat = query->findPath(startRef, endRef, startPos, endPos, &filter, path, &pathCount, PATH_MAX_NODE);
    assert(dtStatusSucceed(stat));
    printf("pathCount: %d\n", pathCount);
    for (int i = 0; i < pathCount; i++) {
        printf("%d\n", path[i]);
    }
    printf("\n");

    {
        printf("================================================ findStraightPath # DT_STRAIGHTPATH_AREA_CROSSINGS ================================================\n");
        float straightPath[PATH_MAX_NODE * 3];
        unsigned char straightPathFlags[PATH_MAX_NODE];
        dtPolyRef straightPathRefs[PATH_MAX_NODE];
        int straightPathCount;
        stat = query->findStraightPath(startPos, endPos, path, pathCount,
            straightPath, straightPathFlags, straightPathRefs,
            &straightPathCount, PATH_MAX_NODE, DT_STRAIGHTPATH_AREA_CROSSINGS);
        assert(dtStatusSucceed(stat));
        printf("straightPathCount: %d\n", straightPathCount);
        for (int i = 0; i < straightPathCount; i++) {
            printf("straightPath: %.3f %.3f %.3f, straightPathFlags: %d, straightPathRefs: %d\n",
                straightPath[i * 3 + 0], straightPath[i * 3 + 1], straightPath[i * 3 + 2],
                straightPathFlags[i], straightPathRefs[i]);
        }
        printf("\n");
    }

    {
        printf("================================================ findStraightPath # DT_STRAIGHTPATH_ALL_CROSSINGS ================================================\n");
        float straightPath[PATH_MAX_NODE * 3];
        unsigned char straightPathFlags[PATH_MAX_NODE];
        dtPolyRef straightPathRefs[PATH_MAX_NODE];
        int straightPathCount;
        stat = query->findStraightPath(startPos, endPos, path, pathCount,
            straightPath, straightPathFlags, straightPathRefs,
            &straightPathCount, PATH_MAX_NODE, DT_STRAIGHTPATH_ALL_CROSSINGS);
        assert(dtStatusSucceed(stat));
        printf("straightPathCount: %d\n", straightPathCount);
        for (int i = 0; i < straightPathCount; i++) {
            printf("straightPath: %.3f %.3f %.3f, straightPathFlags: %d, straightPathRefs: %d\n",
                straightPath[i * 3 + 0], straightPath[i * 3 + 1], straightPath[i * 3 + 2],
                straightPathFlags[i], straightPathRefs[i]);
        }
        printf("\n");
    }

    printf("================================================ moveAlongSurface ================================================\n");
    float resultPos[3];
    dtPolyRef visited[PATH_MAX_NODE];
    int visitedCount;
    bool bHit;
    stat = query->moveAlongSurface(startRef, startPos, endPos, &filter, resultPos, visited, &visitedCount, PATH_MAX_NODE, bHit);
    assert(dtStatusSucceed(stat));
    printf("resultPos: %.2f %.2f %.2f\n", resultPos[0], resultPos[1], resultPos[2]);
    printf("bHit: %d\n", bHit);
    printf("visitedCount: %d\n", visitedCount);
    for (int i = 0; i < visitedCount; i++) {
        printf("%d\n", visited[i]);
    }
    printf("\n");

    printf("================================================ findDistanceToWall # 0================================================\n");
    float hitDist;
    float hitPos[3];
    float hitNormal[3];
    stat = query->findDistanceToWall(startRef, startPos, 30, &filter, &hitDist, hitPos, hitNormal);
    assert(dtStatusSucceed(stat));
    printf("hitPos: %.2f %.2f %.2f\n", hitPos[0], hitPos[1], hitPos[2]);
    printf("hitDist: %f\n", hitDist);
    printf("hitNormal: %.2f %.2f %.2f\n", hitNormal[0], hitNormal[1], hitNormal[2]);
    printf("\n");

    {
        printf("================================================ SlicedFindPath # 0================================================\n");
        stat = query->initSlicedFindPath(startRef, endRef, startPos, endPos, &filter, 0);
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
                printf("pathCount: %d\n", pathCount);
                for (int i = 0; i < pathCount; i++) {
                    printf("%d\n", path[i]);
                }
                break;
            }
        }
        printf("\n");
    }

    {
        printf("================================================ SlicedFindPath # DT_FINDPATH_ANY_ANGLE ================================================\n");
        stat = query->initSlicedFindPath(startRef, endRef, startPos, endPos, &filter, DT_FINDPATH_ANY_ANGLE);
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
                printf("pathCount: %d\n", pathCount);
                for (int i = 0; i < pathCount; i++) {
                    printf("%d\n", path[i]);
                }
                break;
            }
        }
        printf("\n");
    }
    return 0;
}