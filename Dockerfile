FROM alpine

RUN apk add --no-cache curl

# Get etcdctl
ENV ETCD_VER=v3.2.4
RUN \
  cd /tmp && \
  curl -L https://storage.googleapis.com/etcd/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz | \
  tar xz -C /usr/local/bin --strip-components=1

COPY ./etcd-backup /
ENTRYPOINT ["/etcd-backup"]
CMD ["-h"]
