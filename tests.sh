#!/bin/bash

export CURDIR=$PWD
export GOPATH=$CURDIR/../../../../

cd ./tests/c/bin/
./ctest rand
./ctest

cd $CURDIR

go test -tags debug ./tests/...



