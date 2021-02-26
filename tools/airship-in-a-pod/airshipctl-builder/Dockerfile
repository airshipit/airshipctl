FROM ianhowell/base:latest

SHELL ["bash", "-exc"]
ENV DEBIAN_FRONTEND noninteractive

ARG USE_CACHED_AIRSHIPCTL="false"
ENV USE_CACHED_AIRSHIPCTL="false"

ARG AIRSHIPCTL_REPO=https://review.opendev.org/airship/airshipctl
ENV AIRSHIPCTL_REF=$AIRSHIPCTL_REF

ARG AIRSHIPCTL_REF=master
ENV AIRSHIPCTL_REPO=$AIRSHIPCTL_REPO

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
