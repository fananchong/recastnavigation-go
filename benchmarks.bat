set CURDIR=%~dp0
set BASEDIR=%CURDIR%\..\..\..\..\
set GOPATH=%BASEDIR%
echo %GOPATH%

cd tests\c\bin
call cbenchmark.exe

cd %CURDIR%\benchmarks
call go test -v -tags debug -test.bench=".*" -count=1

cd %CURDIR%