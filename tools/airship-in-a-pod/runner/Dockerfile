FROM ianhowell/base:latest

SHELL ["bash", "-exc"]
ENV DEBIAN_FRONTEND noninteractive

ARG k8s_version=v1.18.3
ARG kubectl_url=https://storage.googleapis.com/kubernetes-release/release/"${k8s_version}"/bin/linux/amd64/kubectl

# Update distro and install ansible
RUN apt-get update ;\
    apt-get dist-upgrade -y ;\
    apt-get install -y \
        git \
        git-review \
        apt-transport-https \
        ca-certificates \
        gnupg-agent \
        libvirt-clients \
        gettext-base ;\
    curl -sSLo /usr/local/bin/kubectl "${kubectl_url}" ;\
    chmod +x /usr/local/bin/kubectl ;\
    rm -rf /var/lib/apt/lists/*

COPY assets /opt/assets/
RUN cp -ravf /opt/assets/* / ;\
    rm -rf /opt/assets

ENTRYPOINT /entrypoint.sh
