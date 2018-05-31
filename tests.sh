#!/bin/bash

export CURDIR=$PWD
export GOPATH=$CURDIR/../../../../

cd $CURDIR/tests/c/bin/
./ctest rand
./ctest a 0

cd $CURDIR

go test -tags debug ./tests/...

cd $CURDIR/tests/c/bin/
./ctest a 1

cd $CURDIR

go test -tags debug ./tests/...



