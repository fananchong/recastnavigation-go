#ifndef _SDK_H_
#define _SDK_H_

#ifndef NavMeshAPI
#ifdef _WIN32
#	define NavMeshAPI __declspec(dllimport)
#else
#   define NavMeshAPI 
#endif
#endif


#ifdef __cplusplus
extern "C" {
#endif

	NavMeshAPI void* CreateNavMeshSDK();
	NavMeshAPI int Load(const char* path, void* sdkPtr);
	//path的长度是maxPolys的三倍
	NavMeshAPI void FindPath(void* sdkPtr, float *startPos, float *endPos, float *path, int* pathLen, const int maxPolys);

#ifdef __cplusplus
}
#endif

#endif