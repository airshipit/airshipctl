FROM quay.io/airshipit/toolbox:latest as release

RUN apk update \
    && apk add ca-certificates libvirt-client \
    && rm -rf /var/cache/apk/*
COPY ./certs/* /usr/local/share/ca-certificates/
RUN update-ca-certificates
