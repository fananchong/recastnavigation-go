workspace "ctest"
    configurations { "Debug", "Release" }
    platforms { "x32", "x64" }
    targetdir "../bin/"
    language "C++"
    includedirs {
        "..",
        "../Detour/Include",
        "../DetourTileCache/Include",
        "../Contrib/fastlz",
    }
    flags {
        "C++11",
        "StaticRuntime",
    }

    filter "configurations:Debug"
    defines { "_DEBUG" }
    flags { "Symbols" }
    libdirs { }
    filter "configurations:Release"
    defines { "NDEBUG" }
    libdirs { }
    optimize "On"
    filter { }  


project "ctest"
    kind "ConsoleApp"
    targetname "ctest"
    libdirs { "../bin" }
    files {
        "../main.cpp",
        "../detour.h",
        "../detour.cpp",
        "../Detour/Include/*.h",
        "../Detour/Source/*.cpp",
        "../DetourTileCache/Include/*.h",
        "../DetourTileCache/Source/*.cpp",
        "../Contrib/fastlz",
    }
    
project "cbenchmark"
    kind "ConsoleApp"
    targetname "cbenchmark"
    libdirs { "../bin" }
    files {
        "../benchmark.cpp",
        "../detour.h",
        "../detour.cpp",
        "../Detour/Include/*.h",
        "../Detour/Source/*.cpp",
        "../DetourTileCache/Include/*.h",
        "../DetourTileCache/Source/*.cpp",
        "../Contrib/fastlz",
    }
