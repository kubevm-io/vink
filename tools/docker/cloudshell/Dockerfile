FROM ghcr.io/cloudtty/cloudshell:v0.8.7

ARG ETCDCTL_VERSION=v3.4.20
ARG KREW_VERSION=v0.4.3

# RUN echo "https://mirrors.aliyun.com/alpine/v3.20/main/" >> /etc/apk/repositories \
#     && apk update \
#     && apk add lrzsz

RUN ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')" && echo "Building ARCH of $ARCH.." \
    && curl -fsSLO https://github.com/etcd-io/etcd/releases/download/$ETCDCTL_VERSION/etcd-$ETCDCTL_VERSION-linux-${ARCH}.tar.gz \
    && tar -zxf etcd-$ETCDCTL_VERSION-linux-${ARCH}.tar.gz \
    && chmod +x etcd-$ETCDCTL_VERSION-linux-${ARCH}/etcdctl \
    && rm -rf etcd-$ETCDCTL_VERSION-linux-* \
    && curl -fsSLO https://github.com/kubernetes-sigs/krew/releases/download/$KREW_VERSION/krew-linux_${ARCH}.tar.gz \
    && tar -zxf krew-linux_${ARCH}.tar.gz \
    && chmod +x krew-linux_${ARCH} \
    && mv krew-linux_${ARCH} /usr/local/bin/kubectl-krew \
    && kubectl krew version \
    && echo 'PATH="${PATH}:${HOME}/.krew/bin"' >>~/.bashrc \
    && rm -rf krew-linux_* \
    && kubectl krew install virt \
    && kubectl krew index add kvaps https://github.com/kvaps/krew-index \
    && kubectl krew install kvaps/node-shell
