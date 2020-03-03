ARG GO_IMAGE=docker.io/golang:1.13.1-stretch
ARG RELEASE_IMAGE=scratch
FROM ${GO_IMAGE} as builder

SHELL [ "/bin/bash", "-cex" ]
WORKDIR /usr/src/airshipctl

# Take advantage of caching for dependency acquisition
COPY go.mod go.sum /usr/src/airshipctl/
RUN go mod download

COPY . /usr/src/airshipctl/
ARG MAKE_TARGET=build
RUN for target in $MAKE_TARGET; do make $target; done

FROM ${RELEASE_IMAGE} as release
COPY --from=builder /usr/src/airshipctl/bin/airshipctl /usr/local/bin/airshipctl
USER 65534
ENTRYPOINT [ "/usr/local/bin/airshipctl" ]
