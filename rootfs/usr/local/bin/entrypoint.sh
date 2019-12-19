#!/bin/sh

while :; do
    date
    curl -s -H "mx-api-token:`cat /run/mx-api-token`" ${APPMAN_HOST_IP}/api/v1/device/ethernets/1
    echo
    sleep 10
done &

wait
