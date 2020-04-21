FROM alpine:3.10.4

ARG TARGETPLATFORM

RUN adduser -D -S -h /home/gitlab-runner gitlab-runner

# Package versions pinned for reproducible builds (versions retrieved with `apk info`)
RUN apk add --no-cache \
        bash=5.0.0-r0 \
        ca-certificates=20191127-r0 \
        git=2.22.4-r0 \
        tzdata=2019c-r0

ARG DOCKER_MACHINE_VERSION
ARG DUMB_INIT_VERSION
ARG GIT_LFS_VERSION

COPY gitlab-runner-linux-* /usr/bin/
COPY checksums-* install-deps install-gitlab-runner /tmp/

# Install GNU wget for "-nv" flag support
RUN apk add --no-cache --virtual .fetch-deps wget=1.20.3-r0 && \
    /tmp/install-deps "${TARGETPLATFORM}" "${DOCKER_MACHINE_VERSION}" "${DUMB_INIT_VERSION}" "${GIT_LFS_VERSION}" && \
    apk del .fetch-deps

COPY entrypoint /
RUN chmod +x /entrypoint

STOPSIGNAL SIGQUIT
VOLUME ["/etc/gitlab-runner", "/home/gitlab-runner"]
ENTRYPOINT ["/usr/bin/dumb-init", "/entrypoint"]
CMD ["run", "--user=gitlab-runner", "--working-directory=/home/gitlab-runner"]
