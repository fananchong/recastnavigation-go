set CURDIR=%~dp0
set BASEDIR=%CURDIR%\..\..\..\..\
set GOPATH=%BASEDIR%
echo %GOPATH%

cd %CURDIR%\tests\c\bin
call ctest.exe rand
call ctest.exe a 1
cd %CURDIR%
go test -v -tags debug ./tests/...

cd %CURDIR%\tests\c\bin
call ctest.exe a 0
cd %CURDIR%
go test -tags debug ./tests/...
