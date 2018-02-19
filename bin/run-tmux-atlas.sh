#!/bin/bash

tmux new -d -s my-session 'echo core; ./monitor.sh ~/devices/core fk-core.bin core' \; \
     new-window -d 'echo atlas; ./monitor.sh ~/devices/atlas fk-atlas-module.bin atlas' \; \
     new-window -d 'echo core; touch core.log && tail -f core.log' \; \
     new-window -d 'echo atlas; touch atlas.log && tail -f atlas.log' \; \
     new-window -d './fk-wifi-tool --data-directory ~/data --device wlan1 --wpa-socket /run/wpa_supplicant_wlan1 --device-address 192.168.2.1' \;
