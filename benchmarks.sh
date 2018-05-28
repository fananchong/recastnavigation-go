#!/bin/bash

export CURDIR=$PWD
export GOPATH=$CURDIR/../../../../

cd ./tests/c/bin/
./cbenchmark

cd $CURDIR/benchmarks

go test -v -tags debug -test.bench=".*" -count=1


