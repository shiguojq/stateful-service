# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace

RUN go env -w GO111MODULE=on

COPY go.mod ./go.mod
COPY go.sum ./go.sum
RUN GOPROXY=https://goproxy.cn,direct go mod download

# Copy the go source
COPY api ./api
COPY config ./config
COPY hack ./hack
COPY pkg ./pkg
COPY proto ./proto
COPY slog ./slog
COPY utils ./utils
COPY examples ./examples

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o service_b ./examples/example1/serviceb/service_b.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/service_b .
USER 65532:65532

ENTRYPOINT ["/service_b"]
