#!/bin/bash

pushd ~/fieldkit/testing
make clean
make GOARCH=arm
ssh wpi 'mkdir -p ~/tools/bin'
rsync -zvua --progress build/* wpi:tools/bin
rsync -zvua --progress bin/monitor.sh wpi:
rsync -zvua --progress bin/run-tmux-weather.sh wpi:
rsync -zvua --progress bin/flash-* wpi:

popd

