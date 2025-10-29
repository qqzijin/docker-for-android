#!/bin/sh

# docker bin
[ -s bin/dockerd ] || {
mkdir bin
curl https://download.docker.com/linux/static/stable/aarch64/docker-28.0.1.tgz | 
	tar -xz --strip-components 1 -C bin
}

# ssl CA certs
[ -s etc/ssl/certs/ca-certificates.crt ] || {
curl https://dl-cdn.alpinelinux.org/alpine/v3.21/main/aarch64/ca-certificates-bundle-20241121-r1.apk | \
	tar -xz etc/ssl/certs/ca-certificates.crt
}

adb root
adb push ../docker /data/local/
