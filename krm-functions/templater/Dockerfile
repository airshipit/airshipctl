ARG GO_IMAGE=quay.io/airshipit/golang:1.16.8-buster
ARG RELEASE_IMAGE=quay.io/airshipit/alpine:3.13.5
FROM ${GO_IMAGE} as builder
ARG GOPROXY=""

ENV PATH "/usr/local/go/bin:$PATH"

# Inject custom root certificate authorities if needed
# Docker does not have a good conditional copy statement and requires that a source file exists
# to complete the copy function without error.  Therefore the README.md file will be copied to
# the image every time even if there are no .crt files.
COPY ./certs/* /usr/local/share/ca-certificates/
RUN update-ca-certificates

RUN apt-get update -yq && apt-get upgrade -yq && apt-get install -y gcc make

SHELL [ "/bin/bash", "-cex" ]
WORKDIR /usr/src/airshipctl

# Take advantage of caching for dependency acquisition
COPY go.mod go.sum /usr/src/airshipctl/
RUN go mod download

COPY . /usr/src/airshipctl/
ARG MAKE_TARGET=bin/templater
RUN make ${MAKE_TARGET}

FROM ${RELEASE_IMAGE} as release
ARG MAKE_TARGET=bin/templater
COPY --from=builder /usr/src/airshipctl/${MAKE_TARGET} /usr/local/bin/config-function
USER 65534
CMD ["/usr/local/bin/config-function"]
