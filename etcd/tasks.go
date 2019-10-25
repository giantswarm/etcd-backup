package etcd

import (
	"github.com/giantswarm/microerror"
	"time"
)

func FullBackup(b BackupInterface) error {
	var err error

	version := b.Version()

	start := time.Now()

	err = b.Create()
	if err != nil {
		return microerror.Maskf(err, "Etcd %s creation failed: %s", version, err)
	}

	creationTime := time.Since(start).Milliseconds()

	start = time.Now()

	err = b.Encrypt()
	if err != nil {
		microerror.Maskf(err, "Etcd %s encryption failed: %s", version, err)
	}

	encryptionTime := time.Since(start).Milliseconds()
	start = time.Now()

	size, err := b.Upload()
	if err != nil {
		microerror.Maskf(err, "Etcd %s upload failed: %s", version, err)
	}

	uploadTime := time.Since(start).Milliseconds()

	err = b.SendMetrics(creationTime, encryptionTime, uploadTime, size)
	if err != nil {
		microerror.Maskf(err, "Etcd %s metrics push failed: %s", version, err)
	}

	return nil
}
