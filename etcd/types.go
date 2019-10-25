package etcd

type BackupInterface interface {
	Create() error
	Encrypt() error
	Upload() (int64, error)
	SendMetrics(creationTime int64, encryptionTime int64, uploadTime int64, backupSize int64) error
	Version() string
}

type PrometheusConfig struct {
	Job string
	Url string
}
