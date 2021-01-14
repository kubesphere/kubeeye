.PHONY: ke

BINARY="ke"
GOBIN=$(shell go env GOPATH)/bin
fmt:
	gofmt -w ./pkg ./cmd

ke:
	GO111MODULE=on GOPROXY=https://proxy.golang.org CGO_ENABLED=0 go get -u github.com/gobuffalo/packr/v2/packr2 
    $(GOBIN)/packr2 build -a -o ${BINARY} main.go

# install kubeye
install: ke
	mv ${BINARY} /usr/local/bin/

clean:
	rm ${BINARY}

# uninstall kubeye from local computer
uninstall:
	rm /usr/local/bin/${BINARY} 2> /dev/null

install-goreleaser:
	curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh

build-multiarch: install-goreleaser
	./bin/goreleaser release --snapshot --skip-publish --rm-dist 
