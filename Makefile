.PHONY: build

BINARY="ke"

ke-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BINARY}

ke-darwin:
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ${BINARY}
