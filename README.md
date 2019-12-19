APP Sample
==============

In this article, you will learn
- Build a sample APP
- Access device configuration
- Setup target board
- Install and verify sample APP

Build
--------------

1. Prepare Ubuntu 18.04 amd64 for development

2. Follow the link [Docker](https://docs.docker.com/install/linux/docker-ce/ubuntu/) to install Docker

3. Install armhf emulator

```consle
$ apt install -y qemu-user-static
```

4. Build

```consle
$ ./build.sh
+ docker build -t username/app:0.0.1-armhf .
Sending build context to Docker daemon  14.56MB
Step 1/14 : FROM debian:stretch-slim as qemu
 ---> c2f145c34384
Step 2/14 : RUN apt update && apt install -y qemu-user-static
 ---> Using cache
 ---> dc9be3f151c9
Step 3/14 : FROM arm32v7/debian:stretch-slim as rootfs
 ---> e1943b4072ed
Step 4/14 : COPY --from=qemu /usr/bin/qemu-arm-static /usr/bin/
 ---> Using cache
 ---> 49d5faa1ce63
Step 5/14 : RUN apt update     && apt install -y --no-install-recommends curl     && rm -rf /var/lib/apt/lists/*
 ---> Using cache
 ---> b9e7f860d51f
Step 6/14 : RUN rm /usr/bin/qemu-arm-static
 ---> Using cache
 ---> 249a1919b2b8
Step 7/14 : FROM golang:1.13-alpine as server
 ---> 3024b4e742b0
Step 8/14 : COPY server.go ./
 ---> Using cache
 ---> 39d17d0d17c8
Step 9/14 : RUN GOOS=linux GOARCH=arm GOARM=7 go build -o /usr/local/bin/server server.go
 ---> Running in 0ecd8b690c0a
Removing intermediate container 0ecd8b690c0a
 ---> 1cbc6cd8bd6c
Step 10/14 : FROM scratch
 --->
Step 11/14 : COPY --from=rootfs / /
 ---> Using cache
 ---> 6056d3b60b52
Step 12/14 : COPY --from=server /usr/local/bin/server /usr/local/bin/
 ---> a64843d39b21
Step 13/14 : COPY rootfs /
 ---> 2e833eeec36d
Step 14/14 : ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
 ---> Running in 2e265c218b74
Removing intermediate container 2e265c218b74
 ---> c4dbae0097ec
Successfully built c4dbae0097ec
Successfully tagged username/app:0.0.1-armhf
+ ./tdk pack
INFO[0000] [Save files]
INFO[0000] Copy docker-compose.yml
INFO[0000] Copy metadata.yml
INFO[0000] Copy nginx.conf
INFO[0000] [Save images]
INFO[0000] username/app:0.0.1-armhf, username/app:0.0.1-armhf
INFO[0007] [pack]
INFO[0007] Success!
INFO[0007] sample_0.0.1_armhf.mpkg 18.53 MB
```

Set up target
--------------

1. Connect to debugging port (console) with serial port setting 115200 8N1

2. Login. Default username/password is moxa/moxa, then sudo to root user

3. Set IP address either by

DHCP

```consle
root@Moxa:~# appman device set ethernets 1 enableDhcp=true enable=true wan=true
```

or static IP address

```consle
root@Moxa:~# appman device set ethernets 1 ip=10.144.48.108 netmask=255.255.252.0 gateway=10.144.51.254 'dns=["8.8.8.8"]' enableDhcp=false enable=true wan=true
```

4. Enable SSH service

```consle
root@Moxa:~# appman service set sshserver enable=true
```

Deploy
--------------

1. Copy APP(MPKG) to target

```consle
$ scp sample_0.0.1_armhf.mpkg moxa@10.144.48.108:/tmp
```

2. Install APP on target

```consle
root@Moxa:~# appman app install /tmp/sample_0.0.1_armhf.mpkg
{
  "data": {
    "arch": "armhf",
    "attributes": null,
    "availableVersions": [],
    "category": "",
    "description": "",
    "desiredState": "ready",
    "displayName": "Sample",
    "hardwares": [],
    "health": "wait",
    "icon": "",
    "imageSize": 59387904,
    "name": "sample",
    "state": "init",
    "version": "0.0.1"
  }
}

```

3. You can check status and container by appman and docker

```consle
root@Moxa:~# appman app ls
+------------------+-----------------------+-----------------------+--------+
|       NAME       |        VERSION        | STATE (DESIRED STATE) | HEALTH |
+------------------+-----------------------+-----------------------+--------+
| aie              | 0.0.1-29              | ready (ready)         | good   |
| device           | 1.0.0-153-uc-8112a-me | ready (ready)         | good   |
| edge-web         | 0.19.0-118            | ready (ready)         | good   |
| modbusmaster-tcp | 3.13.0-245            | ready (ready)         | good   |
| sample           | 0.0.1                 | ready (ready)         | good   |
| tagservice       | 1.4.0-144             | ready (ready)         | good   |
+------------------+-----------------------+-----------------------+--------+
root@Moxa:~# docker ps -a | grep username
CONTAINER ID        IMAGE                                        COMMAND                  CREATED             STATUS              PORTS               NAMES
da8deef610c4        username/app:0.0.1-armhf                     "/usr/local/bin/serv…"   52 seconds ago      Up 34 seconds                           sample_server_1
0448b4bebfa8        username/app:0.0.1-armhf                     "/usr/local/bin/entr…"   52 seconds ago      Up 35 seconds                           sample_app_1

```

4. Follow APP's log by `journalctl`. Find source code at `rootfs/usr/local/bin/entrypoint.sh`.

```
root@Moxa:~# journalctl -f sample_app_1 --tail=2
Mon Oct 21 14:30:00 CST 2019
{"data":{"id":1,"wan":true,"enable":true,"enableDhcp":false,"ip":"10.144.48.108","netmask":"255.255.252.0","gateway":"10.144.51.254","dns":["8.8.8.8"],"status":"connected","mac":"00:90:e8:00:00:41","subnet":"10.144.48.0","broadcast":"10.144.51.255","type":"ethernets","name":"eth0","displayName":"LAN1"}}Mon Oct
21 14:30:00 CST 2019

```

5. Say hello to custom HTTP API. The source code is located at `server.go`. Your HTTP service is proxy by ThingsPro Edge HTTP Server. `nginx.conf` defines proxy rules and authentication that syntax complies to [nginx](https://nginx.org/en/docs/).

```console
root@Moxa:~# curl -H "mx-api-token:`cat /var/thingspro/data/mx-api-token`" http://127.0.0.1:59000/sample/hello
hello
```

FAQ
--------------

- APPs is managed by `appman`. Document can be found by appending `-h`.

```consle
root@Moxa:~# appman app -h
app management

Usage:
  appman app [command]

Available Commands:
  export      Export apps to a dest path
  import      Import apps data from a tar file
  install     install an app
  ls          show apps
  restart     start an app
  start       start an app
  stop        stop an app
  uninstall   uninstall an app

Flags:
  -h, --help       help for app
      --no-color   color output

Global Flags:
      --debug        debug output
      --production   production mode

Use "appman app [command] --help" for more information about a command.
```

- [API document](https://thingspro-edge.moxa.online/latest/)
