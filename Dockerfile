FROM golang:1.10.3-alpine3.7

COPY .  /go/src/github.com/kadende-plugins/kadende-provider-file
WORKDIR /go/src/github.com/kadende-plugins/kadende-provider-file

RUN apk --no-cache add curl git make gcc musl-dev\
    && DEP_RELEASE_TAG=v0.4.1 curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# one of the plugin caveats
# https://github.com/alperkose/golangplugins#caveats
RUN go get github.com/kadende/kadende-interfaces/spi

RUN dep ensure -v \
    && make build

