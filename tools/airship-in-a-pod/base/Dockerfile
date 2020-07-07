ARG BASE_IMAGE
FROM ${BASE_IMAGE}

SHELL ["bash", "-exc"]
ENV DEBIAN_FRONTEND noninteractive

ARG CACHE_DIR=/opt/aiap-cache
ENV CACHE_DIR=$CACHE_DIR

ARG ARTIFACTS_DIR=/opt/aiap-artifacts
ENV ARTIFACTS_DIR=$ARTIFACTS_DIR

# Update distro and install common reqs
RUN apt-get update ;\
    apt-get dist-upgrade -y ;\
    apt-get install -y \
        python3-minimal \
        python3-pip \
        python3-setuptools \
        python3-libvirt \
        curl \
        make \
        sudo \
        iproute2 \
        bridge-utils \
        iputils-ping \
        net-tools \
        less \
        jq \
        vim \
        software-properties-common \
        openssh-client ;\
    pip3 install --upgrade wheel ;\
    pip3 install --upgrade ansible ;\
    pip3 install --upgrade yq ;\
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - ;\
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" ;\
    apt-get install -y --no-install-recommends docker-ce-cli ;\
    rm -rf /var/lib/apt/lists/* ; \
    mkdir -p "$ARTIFACTS_DIR" ; \
    mkdir -p "$CACHE_DIR"

COPY wait_for .
COPY signal_complete .
