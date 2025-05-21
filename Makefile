.PHONY: build
build:
	go build -o bin/firewall ./main.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/fwtui ./main.go
