ARG GO_IMAGE=amd64/golang:1.16.6-buster
ARG BUILD_IMAGE=quay.io/airshipit/alpine:3.13.5
ARG RELEASE_IMAGE=quay.io/airshipit/alpine:3.13.5
FROM ${BUILD_IMAGE} as ctls
ARG GOPROXY=""
ARG K8S_VERSION=v1.18.6

RUN apk update && apk add curl

# Inject custom root certificate authorities if needed
# Docker does not have a good conditional copy statement and requires that a source file exists
# to complete the copy function without error.  Therefore the README.md file will be copied to
# the image every time even if there are no .crt files.
COPY ./certs/* /usr/local/share/ca-certificates/
RUN update-ca-certificates

RUN curl -L "https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kubectl" \
      -o /kubectl
RUN curl -L  "https://github.com/projectcalico/calicoctl/releases/download/v3.18.1/calicoctl" \
      -o /calicoctl
RUN chmod +x /kubectl /calicoctl

FROM ${GO_IMAGE} as builder
ARG GOPROXY=""

ENV PATH "/usr/local/go/bin:$PATH"
RUN apt-get update -yq && apt-get upgrade -yq && apt-get install -y gcc make

SHELL [ "/bin/bash", "-cex" ]
WORKDIR /usr/src/airshipctl

# Take advantage of caching for dependency acquisition
COPY go.mod go.sum /usr/src/airshipctl/
RUN go mod download

COPY . /usr/src/airshipctl/
ARG MAKE_TARGET=bin/toolbox
RUN make ${MAKE_TARGET}

FROM ${RELEASE_IMAGE} as release
ARG MAKE_TARGET=bin/toolbox
RUN apk update && apk add ca-certificates curl && rm -rf /var/cache/apk/*
COPY ./certs/* /usr/local/share/ca-certificates/
RUN update-ca-certificates

COPY --from=ctls /kubectl /calicoctl /usr/local/bin/
COPY --from=builder /usr/src/airshipctl/${MAKE_TARGET} /usr/local/bin/config-function
CMD ["/usr/local/bin/config-function"]
