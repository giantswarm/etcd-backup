package etcd

type BackupInterface interface {
	Create() error
	Encrypt() error
	Upload() (int64, error)
	Version() string
}
