set CURDIR=%~dp0
set BASEDIR=%CURDIR:\src\github.com\fananchong\recastnavigation-go\=\%
set GOPATH=%BASEDIR%
echo %GOPATH%
go test -tags debug ./tests/...
REM go test ./tests/...
pause
