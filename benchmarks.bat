set CURDIR=%~dp0
set BASEDIR=%CURDIR:\src\github.com\fananchong\recastnavigation-go\=\%
set GOPATH=%BASEDIR%
echo %GOPATH%

cd benchmarks
call go test -test.bench=".*" -count=1

cd %CURDIR%