.PHONY: ke

BINARY="ke"
GOBIN=$(shell go env GOPATH)/bin
fmt:
	gofmt -w ./pkg ./cmd

test: fmt
	GO111MODULE=on go test -v ./pkg/...

install-packr2:
	GO111MODULE=on GOPROXY=https://proxy.golang.org CGO_ENABLED=0 go get -u github.com/gobuffalo/packr/v2/packr2
	$(GOBIN)/packr2

ke: install-packr2
	GO111MODULE=on CGO_ENABLED=0 go build -o ${BINARY}

# install kubeye
install: ke
	mv ${BINARY} /usr/local/bin/

clean:
	rm ${BINARY}

# uninstall kubeye from local computer
uninstall:
	rm /usr/local/bin/${BINARY} 2> /dev/null
