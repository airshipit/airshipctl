ARG GO_IMAGE=gcr.io/gcp-runtimes/go1-builder:1.15
ARG PLUGINS_RELEASE_IMAGE=quay.io/airshipit/alpine:3.13.5
FROM ${GO_IMAGE} as builder
ARG GOPROXY=""

# Inject custom root certificate authorities if needed
# Docker does not have a good conditional copy statement and requires that a source file exists
# to complete the copy function without error.  Therefore the README.md file will be copied to
# the image every time even if there are no .crt files.
COPY ./certs/* /usr/local/share/ca-certificates/
RUN update-ca-certificates

ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY image/go.mod .
RUN /usr/local/go/bin/go mod download
COPY main.go .
RUN /usr/local/go/bin/go build -v -o /usr/local/bin/config-function ./

FROM ${PLUGINS_RELEASE_IMAGE} as release
COPY --from=builder /usr/local/bin/config-function /usr/local/bin/config-function
CMD ["/usr/local/bin/config-function"]
