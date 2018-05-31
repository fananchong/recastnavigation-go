#!/bin/bash

export CURDIR=$PWD
export GOPATH=$CURDIR/../../../../

cd ./tests/c/bin/
./cbenchmark 0
./cbenchmark 1

cd $CURDIR/benchmarks

go test -v -tags debug -test.bench=".*" -count=1


