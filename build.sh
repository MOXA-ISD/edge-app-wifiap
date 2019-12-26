#!/bin/sh -ex

docker build -t moxaics/wifiap:0.0.1-armhf .
./tdk pack
