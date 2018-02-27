#!/bin/bash

set -xe

BUILD=/tmp/working/fk-build

pushd ~/conservify/flasher
make GOARCH=arm BUILD=$BUILD
popd

pushd ~/fieldkit/app-protocol
make GOARCH=arm BUILD=$BUILD
popd

pushd ~/fieldkit/testing
make GOARCH=arm BUILD=$BUILD
popd

pushd ~/fieldkit/testing
ssh weather-pi 'mkdir -p ~/tools/bin'
rsync -zvua --progress bin/monitor.sh atlas-pi:
rsync -zvua --progress bin/run-tmux-weather.sh atlas-pi:
rsync -zvua --progress bin/flash-* atlas-pi:
rsync -zvua --progress $BUILD/* atlas-pi:tools/bin
popd
