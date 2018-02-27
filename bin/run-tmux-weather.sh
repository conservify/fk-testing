#!/bin/bash

tmux new -d -s my-session 'echo core; ./monitor.sh ~/devices/core fk-core.bin core' \; \
            new-window -d 'echo weather; ./monitor.sh ~/devices/weather fk-weather-module.bin weather' \; \
            new-window -d 'echo core; touch core.log && tail -f core.log' \; \
            new-window -d 'echo weather; touch weather.log && tail -f weather.log' \; \
            new-window -d 'sudo ~/tools/bin/fk-wifi-tool --data-directory ~/data --device wlan1 --wpa-socket /run/wpa_supplicant_wlan1 --device-address 192.168.2.1 --upload-data' \;
