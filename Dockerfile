ARG GO_IMAGE=gcr.io/gcp-runtimes/go1-builder:1.15
ARG RELEASE_IMAGE=scratch
FROM ${GO_IMAGE} as builder

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
ARG MAKE_TARGET=build
RUN for target in $MAKE_TARGET; do make $target; done

FROM ${RELEASE_IMAGE} as release

LABEL org.opencontainers.image.authors='airship-discuss@lists.airshipit.org, irc://#airshipit@freenode' \
      org.opencontainers.image.url='https://airshipit.org' \
      org.opencontainers.image.documentation='https://docs.airshipit.org/airshipctl/' \
      org.opencontainers.image.source='https://opendev.org/airship/airshipctl' \
      org.opencontainers.image.vendor='The Airship Authors' \
      org.opencontainers.image.licenses='Apache-2.0'

ARG BINARY=airshipctl
ENV BINARY=${BINARY}
COPY --from=builder /usr/src/airshipctl/bin/${BINARY} /usr/local/bin/${BINARY}
USER 65534
ENTRYPOINT /usr/local/bin/${BINARY}
