ARCH ?= amd64

.PHONY: build
build:
	go build -o bin/firewall ./main.go

release-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -ldflags="-s -w" -o bin/fwtui-$(ARCH) ./main.go

lint:
	golangci-lint run --fix