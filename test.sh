#!/bin/bash

ps aux | grep hostapd | grep -v grep
ret1=$?
if [ $ret1 -eq 0 ]
then
echo "hostapd process test pass"
else
echo "hostapd process test failed"
fi
ps aux | grep udhcpd | grep -v grep
ret2=$?
if [ $ret2 -eq 0 ]
then
echo "udhcpd process test pass"
else
echo "udhcpd process test failed"
fi