#include "detour.h"
#include <cassert>
#include <stdio.h>
#include <stdlib.h>


inline float frand()
{
    auto v = (float)rand() / (float)RAND_MAX;
    printf("frand, value: %.3f\n", v);
    return v;
}

int main() {

    srand(1);

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

    stat = query->findRandomPoint(&filter, frand, &startRef, startPos);
    assert(dtStatusSucceed(stat));
    stat = query->findRandomPoint(&filter, frand, &endRef, endPos);
    assert(dtStatusSucceed(stat));

    printf("startPos: %.2f %.2f %.2f\n", startPos[0], startPos[1], startPos[2]);
    printf("endPos: %.2f %.2f %.2f\n", endPos[0], endPos[1], endPos[2]);
    printf("startRef: %d\n", startRef);
    printf("endRef: %d\n", endRef);



    return 0;
}