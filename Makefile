.PHONY: build
build:
	go build -o bin/firewall ./main.go ./profiles.go
