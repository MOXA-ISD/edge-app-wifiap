FROM debian:stretch-slim as qemu
RUN apt update && apt install -y qemu-user-static

################################################################################
# ROOTFS
################################################################################
FROM arm32v7/debian:stretch-slim as rootfs

COPY --from=qemu /usr/bin/qemu-arm-static /usr/bin/

RUN apt update \
    && apt install -y --no-install-recommends hostapd curl dhcpd iproute2 iptables iw procps\
    && rm -rf /var/lib/apt/lists/*

RUN rm /usr/bin/qemu-arm-static

################################################################################
# SERVER
################################################################################
FROM golang:1.13-alpine as server

COPY vendor src/
COPY server.go ./
RUN GOOS=linux GOARCH=arm GOARM=7 go build -o /usr/local/bin/server server.go

################################################################################
# merge
################################################################################
FROM scratch

COPY --from=rootfs / /
COPY --from=server /usr/local/bin/server /usr/local/bin/
COPY rootfs /
