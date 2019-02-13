FROM golang:1.10-alpine

RUN apk --no-cache update && apk add git

RUN mkdir -p /go/src/github.com/lfdominguez/docker_log_driver_loki
WORKDIR /go/src/github.com/lfdominguez/docker_log_driver_loki
COPY . /go/src/github.com/lfdominguez/docker_log_driver_loki

RUN go get -d -v ./...
RUN go build --ldflags '-extldflags "-static"' -o /usr/bin/docker-loki-log-driver

FROM scratch
COPY --from=0 /usr/bin/docker-loki-log-driver /docker-loki-log-driver