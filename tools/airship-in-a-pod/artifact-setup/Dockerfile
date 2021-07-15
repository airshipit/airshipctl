FROM quay.io/airshipit/aiap-base:latest

SHELL ["bash", "-exc"]
ENV DEBIAN_FRONTEND noninteractive

ARG USE_CACHED_ARTIFACTS="false"
ENV USE_CACHED_ARTIFACTS="false"

ARG MANIFEST_REPO_URL=https://review.opendev.org/airship/airshipctl
ENV MANIFEST_REPO_URL=$MANIFEST_REPO_URL

ARG MANIFEST_REPO_REF=master
ENV MANIFEST_REPO_REF=$MANIFEST_REPO_REF

ARG MANIFEST_REPO_NAME=airshipctl
ENV MANIFEST_REPO_NAME=$MANIFEST_REPO_NAME

ARG AIRSHIPCTL_REF=
ENV AIRSHIPCTL_REF=$AIRSHIPCTL_REF

# Update distro and install ansible
RUN apt-get update ;\
    apt-get dist-upgrade -y ;\
    apt-get install -y \
        git \
        apt-transport-https \
        ca-certificates \
        gnupg-agent \
        gettext-base ;\
    rm -rf /var/lib/apt/lists/*

COPY assets /opt/assets/
RUN cp -ravf /opt/assets/* / ;\
    rm -rf /opt/assets

ENTRYPOINT /entrypoint.sh
