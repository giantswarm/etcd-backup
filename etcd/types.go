package etcd

type BackupInterface interface {
	Create() error
	Encrypt() error
	Upload() error
	Version() string
}
