.PHONY: ke

BINARY="ke"

fmt:
	gofmt -w ./pkg ./cmd

ke:
	CGO_ENABLED=0 go build -o ${BINARY}

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