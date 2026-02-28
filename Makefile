APP_NAME := nvim-simple-keybind-helper

.PHONY: test build fmt tidy

test:
	go test ./...

build:
	go build ./...

fmt:
	gofmt -w main.go main_test.go

tidy:
	go mod tidy
