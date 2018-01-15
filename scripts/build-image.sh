#!/bin/bash

set -ex

ROOT=$(cd "$(dirname "$0")/.."; pwd)
REV=$(git rev-parse --short HEAD)

docker build \
    --build-arg version=$REV \
    -t bryanl/gimmemotd-server:$REV \
    $ROOT