# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
COPY vendor vendor

# Copy the go source
COPY main.go main.go
COPY pkg/ pkg/

# Build
ENV GO111MODULE="on"
ENV CGO_ENABLED=0
RUN GOOS=linux GOARCH=amd64 go build -a -v -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM alpine:3.15
WORKDIR /
RUN apk --no-cache add curl
RUN curl -s https://raw.githubusercontent.com/armosec/kubescape/master/install.sh | /bin/sh
COPY --from=builder /workspace/manager /usr/local/bin/
RUN addgroup -S kubeeye -g 1000 && adduser -S kubeeye -G kubeeye -u 1000
USER 1000:1000

ENTRYPOINT ["manager"]
