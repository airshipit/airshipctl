FROM ianhowell/base:latest

SHELL ["bash", "-exc"]
ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update ;\
    apt-get dist-upgrade -y ;\
    apt-get install -y \
            libvirt-daemon \
            qemu-kvm \
            libvirt-daemon-system \
            bridge-utils \
            libvirt-clients \
            systemd \
            socat ;\
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
    echo 'group = "root"' >> /etc/libvirt/qemu.conf

COPY assets /opt/assets/
RUN cp -ravf /opt/assets/* / ;\
    rm -rf /opt/assets

ENTRYPOINT /bin/systemd
