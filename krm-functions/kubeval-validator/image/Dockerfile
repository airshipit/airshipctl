ARG GO_IMAGE=gcr.io/gcp-runtimes/go1-builder:1.15
ARG PLUGINS_RELEASE_IMAGE=alpine:3.12.0
FROM ${GO_IMAGE} as function
ENV PATH "/usr/local/go/bin:$PATH"
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY krm-functions/kubeval-validator/image/go.mod .
COPY krm-functions/kubeval-validator/image/go.sum .
RUN go mod download
COPY krm-functions/kubeval-validator/image/main.go .
RUN go build -v -o /usr/local/bin/config-function ./

FROM ${PLUGINS_RELEASE_IMAGE} as release
ENV PYTHONUNBUFFERED=1
RUN echo "**** install Python ****" && \
    apk add --no-cache python3 && \
    if [ ! -e /usr/bin/python ]; then ln -sf python3 /usr/bin/python ; fi && \
    \
    echo "**** install pip ****" && \
    python3 -m ensurepip && \
    rm -r /usr/lib/python*/ensurepip && \
    pip3 install --no-cache --upgrade pip setuptools wheel && \
    if [ ! -e /usr/bin/pip ]; then ln -s pip3 /usr/bin/pip ; fi

RUN pip3 install 'ruamel.yaml==0.16.13' 'openapi2jsonschema==0.9.0' openapi-spec-validator
COPY --from=function /usr/local/bin/config-function /usr/local/bin/config-function
COPY krm-functions/kubeval-validator/image/extract-openapi.py /usr/local/bin/
CMD ["config-function"]
