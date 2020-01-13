#!/bin/bash

#apt-get update
#apt-get install wpasupplicant
pkill wpa_cli
pkill wpa_supplicant
wpa_supplicant -B -i wlan0 -D nl80211 -c /etc/wpa_supplicant/wpa_supplicant.conf
dhclient wlan0
ret1=$?
if [ $ret1 -eq 0 ]
then
	echo "connect to wifiap success"
else
	echo "connect to wifiap failed"
fi
ip route replace default via 10.0.0.1 dev wlan0
ping -I wlan0 -c 1 8.8.8.8
ret2=$?
if [ $ret2 -eq 0 ]
then
	echo "wifiap NAT pass"
else
	echo "wifiap NAT failed"
fi
