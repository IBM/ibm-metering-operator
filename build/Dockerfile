FROM registry.access.redhat.com/ubi8/ubi-minimal:8.1-328

ENV OPERATOR=/usr/local/bin/metering-operator \
    USER_UID=1001 \
    USER_NAME=metering-operator

# install operator binary
COPY build/_output/bin/metering-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
