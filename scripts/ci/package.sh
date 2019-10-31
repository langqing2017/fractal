#!/bin/bash
set -ex

go install ./cmd/gtool
go install ./cmd/gftl

mkdir -p release/fractal-bin
cp $GOPATH/bin/gftl release/fractal-bin/
cp $GOPATH/bin/gtool release/fractal-bin/
if [[ "$ENV_OS" == "ubuntu" ]]; then
    cp transaction/txexec/libwasmlib.so.ubuntu release/fractal-bin/libwasmlib.so
elif [[ "$ENV_OS" == "osx" ]]; then
    cp transaction/txexec/libwasmlib.dylib release/fractal-bin/libwasmlib.dylib
fi

cd release
tar zcvf $PKG_FILE fractal-bin
cd ..
