.PHONY: build run test clean

build:
	go build -o bin/llmfetch ./cmd/llmfetch

run:
	go run ./cmd/llmfetch

test:
	go test ./...

clean:
	rm -rf bin dist
