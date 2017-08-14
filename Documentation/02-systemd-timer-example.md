# Run etcd-backup with systemd timers

Create v2 and v3 backups from etcd running locally and upload to S3 `bucket1`.

Create timer `/etc/systemd/system/etcd-backup.service`
```
[Unit]
Description=etcd
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
Environment="IMAGE=quay.io/giantswarm/etcd-backup:latest"
Environment="NAME=etcd-backup"
Environment="ETCDBACKUP_AWS_ACCESS_KEY=XXX"
Environment="ETCDBACKUP_AWS_SECRET_KEY=YYY"

ExecStartPre=/usr/bin/docker pull $IMAGE
ExecStart=/usr/bin/docker run --rm \
        -e ETCDBACKUP_AWS_ACCESS_KEY=${ETCDBACKUP_AWS_ACCESS_KEY} \
        -e ETCDBACKUP_AWS_SECRET_KEY=${ETCDBACKUP_AWS_SECRET_KEY} \
        -v /var/lib/etcd:/var/lib/etcd \
        -v /etc/etcd/:/etc/etcd/ \
        --net=host \
        --name $NAME $IMAGE \
        -prefix=centaur-vault \
        -etcd-v2-datadir=/var/lib/etcd \
        -etcd-v3-cacert=/etc/etcd/etcd-ca.pem \
        -etcd-v3-cert=/etc/etcd/etcd.pem \
        -etcd-v3-key=/etc/etcd/etcd-key.pem \
        -etcd-v3-endpoints=https://127.0.0.1:2379 \
        -aws-s3-bucket=bucket1 \
        -aws-s3-region=eu-west-1

[Install]
WantedBy=multi-user.target
```

Create timer `/etc/systemd/system/etcd-backup.timer`
```
[Unit]
Description=Exexute etcd-backup every day at 3AM UTC

[Timer]
OnCalendar=*-*-* 03:00:00 UTC

[Install]
WantedBy=multi-user.target
```

Enable timer.
```
systemctl daemon-reload
systemctl enable etcd-backup.timer
systemctl start etcd-backup.timer
```

Check timer schedule.
```
systemctl list-timers
```

Check `etcd-backup` service manually.
```
systemctl start etcd-backup
journalctl -u etcd-backup
```
