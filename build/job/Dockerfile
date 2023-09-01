
# Build the manager binary
FROM --platform=${TARGETPLATFORM} golang:1.19 as builder

#RUN #go env -w GOPROXY=https://goproxy.cn,direct
# Run this with docker build --build_arg $(go env GOPROXY) to override the goproxy
ARG goproxy=https://goproxy.cn,direct
ENV GOPROXY=$goproxy


WORKDIR /workspace

COPY ../../apis apis/
COPY ../../clients clients/
COPY ../../cmd cmd/
COPY ../../pkg/controllers controllers/
COPY ../../pkg pkg/
COPY ../../pkg/constant constant/
COPY ../../plugins plugins/
COPY ../../swaggerDocs swaggerDocs/



COPY ../../go.mod go.mod
COPY ../../go.sum go.sum

RUN go mod tidy && go mod vendor

# Do not force rebuild of up-to-date packages (do not use -a) and use the compiler cache folder

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETPLATFORM} go build   -o /workspace/ke ./cmd/ke


FROM --platform=${TARGETPLATFORM} alpine:3.16

WORKDIR /

COPY --from=builder /workspace/ke /bin/ke

ENTRYPOINT ["/bin/ke"]
