# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace

# Copy the go source
COPY main.go main.go

COPY apis/ apis/
COPY cmd/ cmd/
COPY controllers/ controllers/
COPY pkg/ pkg/
COPY vendor/ vendor/

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Build
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]

