ARG GO_IMAGE=docker.io/golang:1.12.6-stretch
ARG RELEASE_IMAGE=scratch
FROM ${GO_IMAGE} as builder

SHELL [ "/bin/bash", "-cex" ]
ADD . /usr/src/airshipctl
WORKDIR /usr/src/airshipctl
ENV GO111MODULE=on

RUN make get-modules

ARG MAKE_TARGET=build
RUN make ${MAKE_TARGET} && \
    if [[ "${MAKE_TARGET}" == 'lint' ]]; then \
      mkdir -p /usr/src/airshipctl/bin; \
      touch /usr/src/airshipctl/bin/airshipctl; \
    fi

FROM ${RELEASE_IMAGE} as release
COPY --from=builder /usr/src/airshipctl/bin/airshipctl /usr/local/bin/airshipctl
USER 65534
ENTRYPOINT [ "/usr/local/bin/airshipctl" ]
