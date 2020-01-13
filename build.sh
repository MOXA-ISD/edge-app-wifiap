#!/bin/sh -ex

ARCH=${ARCH:-armhf}
DAEMON_VERSION=0.0.1
VERSION=${DAEMON_VERSION}-${DRONE_BUILD_NUMBER:-unknown}

docker build -t moxaics/wifiap:${VERSION}-${ARCH} .
./tdk pack -e VERSION=${VERSION} -e ARCH=${ARCH}
