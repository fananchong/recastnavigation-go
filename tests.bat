set CURDIR=%~dp0
set BASEDIR=%CURDIR%\..\..\..\..\
set GOPATH=%BASEDIR%
echo %GOPATH%

cd tests\c\bin
call ctest.exe rand
call ctest.exe
cd %CURDIR%
go test -tags debug ./tests/...
