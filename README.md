# etcd-backup

[![CircleCI](https://circleci.com/gh/giantswarm/etcd-backup.svg?&style=shield&circle-token=2335d256956ba9d0614cec9e0b496a2f6a3b15ec)](https://circleci.com/gh/giantswarm/etcd-backup) [![Docker Repository on Quay](https://quay.io/repository/giantswarm/etcd-backup/status "Docker Repository on Quay")](https://quay.io/repository/giantswarm/etcd-backup)

Tool for creating etcd backups and uploading them to AWS S3.

## Getting project

### Quickstart with docker

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

### How to build

```
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' .
```

## Usage

### Dependencies

- `etcdctl` installed.
- AWS S3 bucket.
- AWS access and secret key with write permissions for this bucket.

### Usage help

```
etcd-backup -help
```

### Create V3 backup

By default tool creates only V3 backup and uploads file to AWS S3.

```
export ETCDBACKUP_AWS_ACCESS_KEY=XXX
export ETCDBACKUP_AWS_SECRET_KEY=YYY
export ETCDBACKUP_PASSPHRASE=ZZZ

etcd-backup -aws-s3-bucket bucket -prefix cluster1
```

### Create V2 and V3 backup

To create both V2 and V3 make sure etcd data directory accessible locally.

```
etcd-backup -aws-s3-bucket $BUCKET_NAME -prefix $CLUSTER_NAME -etcd-v2-datadir /var/lib/etcd
```

### Restore backup

To restore backup use following [guide](Documentation/01-restore-etcd-from-backups.md) as example.

## Future Development
- Implement additional storage backends.

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/etcd-backup/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on submitting patches, the contribution workflow as well as reporting bugs.

## License

PROJECT is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
