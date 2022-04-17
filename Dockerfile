# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace

COPY apis/ apis/
COPY cmd/ cmd/
COPY controllers/ controllers/
COPY pkg/ pkg/
COPY vendor/ vendor/
COPY plugins/ plugins/

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

ENV CGO_ENABLED=0

# Build
RUN go install -v ./cmd/...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o plugin-manage plugins/plugin-manage/main.go
RUN GOOS=linux GOARCH=amd64 go build -a -o ke-kubebench plugins/kubebench/main.go
RUN GOOS=linux GOARCH=amd64 go build -a -o ke-kubehunter plugins/kubehunter/main.go
RUN GOOS=linux GOARCH=amd64 go build -a -o ke-kubescape plugins/kubescape/main.go

# IMAGE TARGETS

FROM alpine:3.15 as ke-manager
WORKDIR /
COPY --from=builder /go/bin/ke .
COPY --from=builder /go/bin/ke-manager .
RUN addgroup -S kubeeye -g 1000 && adduser -S kubeeye -G kubeeye -u 1000
USER 1000:1000

ENTRYPOINT ["/ke-manager"]

FROM gcr.io/distroless/static:nonroot as pluginmanage
WORKDIR /
COPY --from=builder /workspace/plugin-manage .
USER 65532:65532

ENTRYPOINT ["/plugin-manage"]

FROM aquasec/kube-bench:v0.6.6 as kubebench
WORKDIR /
COPY --from=builder /workspace/ke-kubebench .
USER 65532:65532

ENTRYPOINT ["/ke-kubebench"]

FROM aquasec/kube-hunter:latest as kubehunter
COPY --from=builder /workspace/ke-kubehunter /usr/local/bin/
RUN addgroup -S kubeeye -g 1000 && adduser -S kubeeye -G kubeeye -u 1000
USER 1000:1000

ENTRYPOINT ["ke-kubehunter"]

FROM alpine:3.15 as kubescape
WORKDIR /
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk --no-cache add curl
RUN curl -s https://raw.githubusercontent.com/armosec/kubescape/master/install.sh | /bin/sh
COPY --from=builder /workspace/ke-kubescape /usr/local/bin/
RUN addgroup -S kubeeye -g 1000 && adduser -S kubeeye -G kubeeye -u 1000
USER 1000:1000

ENTRYPOINT ["ke-kubescape"]