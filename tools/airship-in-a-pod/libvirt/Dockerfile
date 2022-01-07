ARG BASE_IMAGE=quay.io/airshipit/aiap-base:latest
FROM ${BASE_IMAGE}

SHELL ["bash", "-exc"]
ENV DEBIAN_FRONTEND noninteractive

ARG k8s_version=v1.18.3
ARG kubectl_url=https://storage.googleapis.com/kubernetes-release/release/"${k8s_version}"/bin/linux/amd64/kubectl

RUN apt-get update ;\
    apt-get dist-upgrade -y ;\
    apt-get install -y \
            libvirt-daemon \
            qemu-kvm \
            libvirt-daemon-system \
            bridge-utils \
            libvirt-clients \
            systemd \
            socat \
            libguestfs-tools \
            linux-image-generic ;\
    find /etc/systemd/system \
         /usr/lib/systemd/system \
         -path '*.wants/*' \
         -not -name '*journald*' \
         -not -name '*systemd-tmpfiles*' \
         -not -name '*systemd-user-sessions*' \
         -exec rm \{} \; ;\
    systemctl set-default multi-user.target ;\
    sed -i 's|SocketMode=0660|SocketMode=0666|g' /lib/systemd/system/libvirtd.socket ;\
    systemctl enable libvirtd ;\
    systemctl enable virtlogd ;\
    echo 'user = "root"' >> /etc/libvirt/qemu.conf ;\
    echo 'group = "root"' >> /etc/libvirt/qemu.conf ;\
    curl -sSLo /usr/local/bin/kubectl "${kubectl_url}" ;\
    chmod +x /usr/local/bin/kubectl

COPY assets /opt/assets/
RUN cp -ravf /opt/assets/* / ;\
    rm -rf /opt/assets

ENTRYPOINT /bin/systemd
