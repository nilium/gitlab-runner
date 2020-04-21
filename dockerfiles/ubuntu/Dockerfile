FROM ubuntu:18.04

ARG TARGETPLATFORM

ENV DEBIAN_FRONTEND=noninteractive
# Package versions pinned for reproducible builds (versions retrieved with `apt-cache show`)
RUN apt-get update -y && \
    apt-get install -y --no-install-recommends \
        ca-certificates=20180409 wget=1.19.4-1ubuntu2.2 apt-transport-https=1.6.12 \
        tzdata=2019c-0ubuntu0.18.04 \
        file=1:5.32-2ubuntu0.3 git=1:2.17.0-1ubuntu1 git-man=1:2.17.0-1ubuntu1 \
        curl=7.58.0-2ubuntu3.8 libmagic1:amd64=1:5.32-2ubuntu0.3 libx11-6:amd64=2:1.6.4-3ubuntu0.2 \
        multiarch-support=2.27-3ubuntu1 netbase=5.4 openssh-client=1:7.6p1-4ubuntu0.3 \
        patch=2.7.6-2ubuntu1.1 xauth=1:1.0.10-1 xz-utils=5.2.2-1.3 && \
    rm -rf /var/lib/apt/lists/*

ARG DOCKER_MACHINE_VERSION
ARG DUMB_INIT_VERSION
ARG GIT_LFS_VERSION

COPY gitlab-runner_*.deb checksums-* install-deps install-gitlab-runner /tmp/
RUN /tmp/install-deps "${TARGETPLATFORM}" "${DOCKER_MACHINE_VERSION}" "${DUMB_INIT_VERSION}" "${GIT_LFS_VERSION}"

COPY entrypoint /
RUN chmod +x /entrypoint

STOPSIGNAL SIGQUIT
VOLUME ["/etc/gitlab-runner", "/home/gitlab-runner"]
ENTRYPOINT ["/usr/bin/dumb-init", "/entrypoint"]
CMD ["run", "--user=gitlab-runner", "--working-directory=/home/gitlab-runner"]
