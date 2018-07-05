#!/bin/bash

set -xe

BUILD=/tmp/working/fk-build

pushd ~/conservify/flasher
make GOARCH=arm BUILD=$BUILD binaries-all
popd

pushd ~/fieldkit/app-protocol
make GOARCH=arm BUILD=$BUILD binaries-all
popd

pushd ~/fieldkit/testing
make GOARCH=arm BUILD=$BUILD binaries-all
popd

ssh atlas-pi 'mkdir -p ~/tools/bin'

mv $BUILD/linux-arm/* $BUILD/
rmdir $BUILD/linux-arm

pushd ~/fieldkit/testing
rsync -zvua --progress bin/monitor.sh atlas-pi:
rsync -zvua --progress bin/run-tmux-atlas.sh atlas-pi:
rsync -zvua --progress bin/flash-* atlas-pi:
rsync -zvua --progress $BUILD/* atlas-pi:tools/bin
popd
