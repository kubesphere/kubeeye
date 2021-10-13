.PHONY: ke

BINARY="kubeeye"
GOBIN=$(shell go env GOPATH)/bin
fmt:
	gofmt -w ./pkg ./cmd

test: fmt
	GO111MODULE=on go test -v ./pkg/...

ke:
	GO111MODULE=on GOPROXY=https://goproxy.io CGO_ENABLED=0 GO15VENDOREXPERIMENT=1 go build -o ${BINARY}

# install KubeEye
install: ke
	mv ${BINARY} /usr/local/bin/

clean:
	rm ${BINARY}

# uninstall KubeEye from local computer
uninstall:
	rm /usr/local/bin/${BINARY} 2> /dev/null
