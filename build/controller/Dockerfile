# Build the manager binary
FROM  golang:1.19 as builder

#ARG goproxy=https://goproxy.cn,direct
#ENV GOPROXY=$goproxy


WORKDIR /workspace

COPY ../../apis apis/
COPY ../../clients clients/
COPY ../../cmd cmd/
COPY ../../pkg/controllers controllers/
COPY ../../pkg pkg/
COPY ../../pkg/constant constant/
COPY ../../plugins plugins/
COPY ../../swaggerDocs swaggerDocs/

# Copy the Go Modules manifests
COPY ../../go.mod go.mod
COPY ../../go.sum go.sum

RUN go mod tidy && go mod vendor

ENV CGO_ENABLED=0

# Build
RUN go install -v ./cmd/...

FROM alpine:3.15 as ke-manager

WORKDIR /

COPY --from=builder /go/bin/ke-manager .

RUN addgroup -S kubeeye -g 1000 && adduser -S kubeeye -G kubeeye -u 1000

USER 1000:1000

ENTRYPOINT ["/ke-manager"]
