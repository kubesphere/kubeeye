# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace

COPY apis/ apis/
COPY cmd/ cmd/
COPY controllers/ controllers/
COPY pkg/ pkg/
COPY vendor/ vendor/
COPY web/ web/

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

ENV CGO_ENABLED=0

# Build
RUN go install -v ./cmd/...

FROM alpine:3.15
WORKDIR /
COPY --from=builder /go/bin/ke .
COPY --from=builder /go/bin/ke-web .
RUN addgroup -S kubeeye && adduser -S kubeeye -G kubeeye
USER kubeeye

ENTRYPOINT ["/ke-web"]

