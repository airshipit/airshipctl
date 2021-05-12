ARG GO_IMAGE=gcr.io/gcp-runtimes/go1-builder:1.15
ARG PLUGINS_BUILD_IMAGE=alpine:3.12.0
ARG PLUGINS_RELEASE_IMAGE=alpine:3.12.0
FROM ${PLUGINS_BUILD_IMAGE} as ctls
ARG GOPROXY=""

RUN apk update && apk add curl
# Inject custom root certificate authorities if needed
# Docker does not have a good conditional copy statement and requires that a source file exists
# to complete the copy function without error.  Therefore the README.md file will be copied to
# the image every time even if there are no .crt files.
COPY ./certs/* /usr/local/share/ca-certificates/
RUN update-ca-certificates

RUN curl -L "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl" \
      -o /kubectl
RUN curl -L  "https://github.com/projectcalico/calicoctl/releases/download/v3.18.1/calicoctl" \
      -o /calicoctl
RUN chmod +x /kubectl /calicoctl

FROM ${GO_IMAGE} as builder
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY image/go.mod .
RUN /usr/local/go/bin/go mod download
COPY main.go .
RUN /usr/local/go/bin/go build -v -o /usr/local/bin/config-function ./

FROM ${PLUGINS_RELEASE_IMAGE} as release
RUN apk update && apk add curl && rm -rf /var/cache/apk/*
COPY --from=ctls /kubectl /calicoctl /usr/local/bin/
COPY --from=builder /usr/local/bin/config-function /usr/local/bin/config-function
CMD ["/usr/local/bin/config-function"]
