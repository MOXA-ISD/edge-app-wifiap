#!/bin/sh -ex

docker build -t username/app:0.0.1-armhf .
./tdk pack
