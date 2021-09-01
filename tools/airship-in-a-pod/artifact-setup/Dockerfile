ARG BASE_IMAGE=quay.io/airshipit/aiap-base:latest
FROM ${BASE_IMAGE}

SHELL ["bash", "-exc"]
ENV DEBIAN_FRONTEND noninteractive

ARG USE_CACHED_ARTIFACTS="false"
ENV USE_CACHED_ARTIFACTS=$USE_CACHED_ARTIFACTS

ARG AIRSHIPCTL_REPO_URL=https://opendev.org/airship/airshipctl
ENV AIRSHIPCTL_REPO_URL=$AIRSHIPCTL_REPO_URL

ARG AIRSHIPCTL_REPO_REF=master
ENV AIRSHIPCTL_REPO_REF=$AIRSHIPCTL_REPO_REF

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
