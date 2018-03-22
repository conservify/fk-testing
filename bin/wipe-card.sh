#!/bin/bash

set -xe

sudo dd if=/dev/zero of=$1 bs=1M count=1
