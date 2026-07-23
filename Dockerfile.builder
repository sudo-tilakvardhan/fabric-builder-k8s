ARG GO_VER=1.25.0

FROM golang:${GO_VER} AS build

WORKDIR /src
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/build   ./cmd/build   && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/detect  ./cmd/detect  && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/release ./cmd/release && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/run     ./cmd/run

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

COPY --from=build /out/ /opt/hyperledger/k8s_builder/bin/

ENTRYPOINT ["/bin/sh"]
CMD ["-c", "tail -f /dev/null"]
