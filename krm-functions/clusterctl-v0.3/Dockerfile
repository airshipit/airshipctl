ARG PLUGINS_BUILD_IMAGE=quay.io/airshipit/alpine:3.13.5
ARG PLUGINS_RELEASE_IMAGE=quay.io/airshipit/alpine:3.13.5
ARG CCTL_IMAGE=quay.io/airshipit/clusterctl:latest

FROM ${PLUGINS_BUILD_IMAGE} as ctls
# Inject custom root certificate authorities if needed
# Docker does not have a good conditional copy statement and requires that a source file exists
# to complete the copy function without error.  Therefore the README.md file will be copied to
# the image every time even if there are no .crt files.
RUN apk update && apk add curl
COPY ./certs/* /usr/local/share/ca-certificates/
RUN update-ca-certificates
ARG CCTL_VERSION=0.3.23
RUN curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/v${CCTL_VERSION}/clusterctl-linux-amd64 -o /clusterctl
RUN chmod +x /clusterctl

FROM ${CCTL_IMAGE} as cctl_image
FROM ${PLUGINS_RELEASE_IMAGE} as release
# Inject custom root certificate authorities if needed
# Docker does not have a good conditional copy statement and requires that a source file exists
# to complete the copy function without error.  Therefore the README.md file will be copied to
# the image every time even if there are no .crt files.
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY ./certs/* /usr/local/share/ca-certificates/
RUN update-ca-certificates
COPY --from=ctls /clusterctl /usr/local/bin/clusterctl
COPY --from=cctl_image /usr/local/bin/config-function /usr/local/bin/config-function
ENV HOME=/workdir
WORKDIR $HOME/.cluster-api
RUN chmod -R a+w $HOME/.cluster-api
CMD ["config-function"]
