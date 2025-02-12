FROM golang:latest as BUILDER

MAINTAINER zengchen1024<chenzeng765@gmail.com>

# build binary
WORKDIR /go/src/github.com/opensourceways/robot-gitee-hook-delivery
COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 go build -a -o robot-gitee-hook-delivery .

# copy binary config and utils
FROM alpine:3.14
COPY  --from=BUILDER /go/src/github.com/opensourceways/robot-gitee-hook-delivery/robot-gitee-hook-delivery /opt/app/robot-gitee-hook-delivery

ENTRYPOINT ["/opt/app/robot-gitee-hook-delivery"]
