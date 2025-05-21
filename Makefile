.PHONY: build
build:
	go build -o bin/firewall ./main.go

release-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/fwtui ./main.go
