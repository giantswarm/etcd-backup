# Restoring etcd cluster from backups

## Prerequisites

- etcd cluster 3+ nodes (Healthy or broken).
- backups files for v2 and v3 stores.

## Restoring a host cluster (multiple etcd members)

This guide was written for cluster with following nodes:
- etcd node 1 - Name/UUID: 00000001, IP: 172.16.238.101
- etcd node 2 - Name/UUID: 00000003, IP: 172.16.238.102
- etcd node 3 - Name/UUID: 00000002, IP: 172.16.238.103

### Stop etcd on all nodes.

```bash
sudo systemctl stop etcd3
```

### Decrypt backups

If backups have `.enc` extension, they should be decrypted with `gpg`.

```
gpg --output backup.db --decrypt backup.db.enc
```

### Copy backup files to etcd node 1.
```bash
scp locallab-etcd-backup-v3-2017-07-27T20-50-07.db.tar.gz $ETCD_NODE_1:/tmp
scp locallab-etcd-backup-v2-2017-07-31T09-00-30.tar.gz $ETCD_NODE_1:/tmp
```

**NOTE: All commands below should be executed from root user. Use `sudo -i` to become root.**

### Prepare etcd datadir

```bash
(etcd node 1)
# Backup existing just in case
mkdir /var/lib/etcd3.backup/
mv /var/lib/etcd3/* /var/lib/etcd3.backup/

tar xf /tmp/locallab-etcd-backup-v2-2017-07-31T09-00-30.tar.gz -C /tmp
cp -r /tmp/locallab-etcd-backup-v2-2017-07-31T09-00-30/* /var/lib/etcd3/
tar xf /tmp/locallab-etcd-backup-v3-2017-07-27T20-50-07.db.tar.gz -C /tmp
cp /tmp/locallab-etcd-backup-v3-2017-07-27T20-50-07.db /var/lib/etcd3/member/snap/db
```

### Start single node etcd cluster
```bash
(etcd node 1)
# We are initializing cluster manually, so remove discovery URL from config
sed -i '/--discovery/d' /etc/systemd/system/etcd3.service

# As per official etcd guide, first start cluster with --force-new-cluster flag, to initialize new cluster for existing data.
sed -i '/--trusted-ca-file/a\ \ --force-new-cluster\ \\' /etc/systemd/system/etcd3.service
systemctl daemon-reload
systemctl start etcd3

# Check that one node cluster successfully started and listening
etcdctl --endpoints https://127.0.0.1:2379 member list

# Remove --force-new-cluster and restart
sed -i '/--force-new-cluster/d' /etc/systemd/system/etcd3.service
systemctl daemon-reload
systemctl restart etcd3

# Update peerURL, because --froce-new-cluster sets localhost
etcdctl --endpoints https://127.0.0.1:2379 member list
etcdctl --endpoints https://127.0.0.1:2379 member update $MEMBER https://172.16.238.101:2380
```

### Add second node
```bash
(etcd node 2)
rm -rf /var/lib/etcd3/*

sed -i '/--discovery/d' /etc/systemd/system/etcd3.service
sed -i '/--trusted-ca-file/a\ \ --initial-cluster-state\ existing\ \\' /etc/systemd/system/etcd3.service
sed -i '/--trusted-ca-file/a\ \ --initial-cluster\ 00000001=https://172.16.238.101:2380,00000003=https://172.16.238.102:2380\ \\' /etc/systemd/system/etcd3.service
systemctl daemon-reload
systemctl start etcd3

(etcd node 1)
etcdctl --endpoints https://127.0.0.1:2379 member add 00000003 https://172.16.238.102:2380

# Check that node successfully added to cluster (wait at least 30 seconds)
etcdctl --endpoints https://127.0.0.1:2379 member list
```

### Add third node
```bash
(etcd node 3)
rm -rf /var/lib/etcd3/*

sed -i '/--discovery/d' /etc/systemd/system/etcd3.service
sed -i '/--trusted-ca-file/a\ \ --initial-cluster-state\ existing\ \\' /etc/systemd/system/etcd3.service
sed -i '/--trusted-ca-file/a\ \ --initial-cluster\ 00000001=https:\/\/172.16.238.101:2380,00000003=https:\/\/172.16.238.102:2380,\00000002=https:\/\/172.16.238.103:2380 \\' /etc/systemd/system/etcd3.service
systemctl daemon-reload
systemctl start etcd3

(etcd node 1)
etcdctl --endpoints https://127.0.0.1:2379 member add 00000002 https://172.16.238.103:2380

# Check that node successfully added to cluster (wait at least 30 seconds)
etcdctl --endpoints https://127.0.0.1:2379 member list
```

## Restoring a guest cluster (single etcd member)

### Copy db backup from s3
Find the [etcd backup to restore](https://s3.console.aws.amazon.com/s3/buckets/etcd-backups.giantswarm.io/?region=eu-central-1&tab=overview) and make it public. Copy the link and download in the master (wget). After download it, go to the permission tab and removes `Read` right for everyone to leave as before (not public).
```
cd /tmp
wget https://s3-eu-west-1.amazonaws.com/etcd-backups.giantswarm.io/<backup_name>.db.tar.gz
tar -xvzf <BACKUP_FILE>.db.tar.gz
```

### Restore backup in tmp folder
```
ETCDCTL_API=3 etcdctl snapshot restore <backup_name>.db \
  --cacert /etc/kubernetes/ssl/etcd/client-ca.pem \
  --cert /etc/kubernetes/ssl/etcd/client-crt.pem \
  --key /etc/kubernetes/ssl/etcd/client-key.pem
```

**NOTE: All commands below should be executed from root user. Use `sudo -i` to become root.**

### Stop etcd unit
```
systemctl stop etcd3
```

### Copy etcd datadir
```
$ rm -rf /var/lib/etcd/member/
$ cp -R default.etcd/member/* /var/lib/etcd/member/
```

### Start etcd forcing new cluster data
```
# As per official etcd guide, first start cluster with --force-new-cluster flag, to initialize new cluster for existing data.
sed -i '/--trusted-ca-file/a\ \   --force-new-cluster\ \\' /etc/systemd/system/etcd3.service

systemctl daemon-reload

systemctl start etcd3
```

### Cleanup etcd unit

```
sed -i '/--force-new-cluster\ \\/d' /etc/systemd/system/etcd3.service
```

## Links

- [Official restore guide for V3](https://github.com/coreos/etcd/blob/master/Documentation/op-guide/recovery.md)
- [Official restore guide for V2](https://github.com/coreos/etcd/blob/master/Documentation/v2/admin_guide.md#restoring-a-backup)
