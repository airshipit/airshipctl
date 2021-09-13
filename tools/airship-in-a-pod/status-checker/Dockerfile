ARG BASE_IMAGE=alpine
FROM ${BASE_IMAGE}

COPY assets /opt/assets/
RUN cp -ravf /opt/assets/* / ;\
    rm -rf /opt/assets

ENTRYPOINT /entrypoint.sh
