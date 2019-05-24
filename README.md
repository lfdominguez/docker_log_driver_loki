# Grafana Loki log driver for Docker
[![Build Status](https://travis-ci.com/lfdominguez/docker_log_driver_loki.svg?branch=master)](https://travis-ci.com/lfdominguez/docker_log_driver_loki)
[![](https://img.shields.io/github/release/lfdominguez/docker_log_driver_loki.svg)](https://github.com/lfdominguez/docker_log_driver_loki/releases)
![](https://img.shields.io/github/license/lfdominguez/docker_log_driver_loki.svg)
![](https://img.shields.io/github/downloads/lfdominguez/docker_log_driver_loki/total.svg)
![](https://img.shields.io/github/release-date/lfdominguez/docker_log_driver_loki.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/lfdominguez/docker_log_driver_loki)](https://goreportcard.com/report/github.com/lfdominguez/docker_log_driver_loki)
[![Maintainability](https://api.codeclimate.com/v1/badges/52ca62a9438c4e97ac59/maintainability)](https://codeclimate.com/github/lfdominguez/docker_log_driver_loki/maintainability)

This project allow to send all Docker Logs to Grafana Loki Server.

## Build

### From source

Clone from *GitHub*:

```sh
git clone https://github.com/lfdominguez/docker_log_driver_loki
```

and then use the *Makefile*:

```sh
make
```

### From releases

Download:
 * [Latest release](https://github.com/lfdominguez/docker_log_driver_loki/releases)
 * [Plugin config file `config.json`](https://raw.githubusercontent.com/lfdominguez/docker_log_driver_loki/master/config.json).

Create the *rootfs* of the plugin and copy the release and config file

```sh
mkdir -p ./plugin/rootfs
cp config.json ./plugin/
cp docker_log_driver_loki ./plugin/rootfs/
```

## Create plugin

## If you build it or downloaded the release

```sh
docker plugin disable -f lfdominguez/docker-log-driver-loki
docker plugin rm -f lfdominguez/docker-log-driver-loki
docker plugin create lfdominguez/docker-log-driver-loki ./plugin
docker plugin enable lfdominguez/docker-log-driver-loki
```

## Form DockerHub

`TODO`

## Use

Docker support on command line set the logging driver with `--log-driver`, the plugin has this options (required):

* `loki-host`: Loki host address.
* `loki-port`: Loki service port.

can be setted with `--log-opt`, example:

```sh
docker run --rm -it -e LOG_LEVEL=debug --log-driver lfdominguez/docker-log-driver-loki --log-opt loki-host=192.168.120.159 --log-opt loki-port=80 hello-world
```