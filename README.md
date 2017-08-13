[![Docker Repository on Quay](https://quay.io/repository/giantswarm/etcd-backup/status?token=0728b9a8-10a6-47ac-a1e8-e6e07d2a8747 "Docker Repository on Quay")](https://quay.io/repository/giantswarm/etcd-backup)

# etcd-backup - tool for creating etcd backups and uploading them to AWS S3

## Quickstart with docker

Create etcd V2 and V3 backup and upload to S3.

```
docker run --rm \
    -e ETCDBACKUP_AWS_ACCESS_KEY=XXX \
    -e ETCDBACKUP_AWS_SECRET_KEY=YYY \
    -e ETCDBACKUP_PASSPHRASE=ZZZ \
    -v /var/lib/etcd:/var/lib/etcd \
    quay.io/giantswarm/etcd-backup \
    -aws-s3-bucket bucket \
    -prefix cluster1 \
    -etcd-v3-endpoints http://172.17.0.1:2379 \
    -etcd-v2-datadir /var/lib/etcd
```

## Build

```
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' .
```

## Usage

By default tool creates only V3 backup and uploads file to AWS S3.

```
export ETCDBACKUP_AWS_ACCESS_KEY=XXX
export ETCDBACKUP_AWS_SECRET_KEY=YYY
export ETCDBACKUP_PASSPHRASE=ZZZ

etcd-backup -aws-s3-bucket bucket -prefix cluster1
```

To create both V2 and V3 make sure etcd data directory accessible locally.

```
etcd-backup -aws-s3-bucket $BUCKET_NAME -prefix $CLUSTER_NAME -etcd-v2-datadir /var/lib/etcd
```

# TODO
- implement encryption
