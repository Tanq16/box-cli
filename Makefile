BINARY_NAME=box

.PHONY: build build-all run tidy clean

build:
	go build -ldflags "-s -w" -o $(BINARY_NAME) .

build-all:
	GOOS=linux   GOARCH=amd64 go build -ldflags "-s -w" -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags "-s -w" -o dist/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -ldflags "-s -w" -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags "-s -w" -o dist/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o dist/$(BINARY_NAME)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -ldflags "-s -w" -o dist/$(BINARY_NAME)-windows-arm64.exe .

run:
	go run . $(ARGS)

tidy:
	go mod tidy

clean:
	rm -rf dist/ $(BINARY_NAME)
