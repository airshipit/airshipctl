ARG BASE_IMAGE=quay.io/airshipit/aiap-base:latest
FROM ${BASE_IMAGE}

SHELL ["bash", "-exc"]
ENV DEBIAN_FRONTEND noninteractive

# Update distro and install ansible
RUN apt-get update ;\
    apt-get dist-upgrade -y ;\
    apt-get install -y \
        acl \
        nfs4-acl-tools \
        python3-apt \
        python3-jmespath \
        python3-lxml \
        virt-manager \
        virtinst \
    ;\
    rm -rf /var/lib/apt/lists/*

COPY assets /opt/assets/
RUN cp -ravf /opt/assets/* / ;\
    rm -rf /opt/assets

ENTRYPOINT /entrypoint.sh
