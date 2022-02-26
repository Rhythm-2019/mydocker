VERSION := $(shell cat VERSION)
BUILD := $(git rev-parse HEAD)

LDFLAGS=-ldflags "-w -s -X main.Version=${VERSION} -X main.Build=${BUILD}"

install:
	go env -w GOPROXY=https://goproxy.cn,direct && go mod vendor

build:
	GOOS=linux; go build -mod=vendor -trimpath -o ./bin/mydocker ${LDFLAGS} ./cmd/mydocker/main.go

clean:
	@rm -rf ./bin