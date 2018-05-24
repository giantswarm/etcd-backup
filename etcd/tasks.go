package etcd

import (
	"github.com/giantswarm/microerror"
)

func FullBackup(b BackupInterface) error {
	var err error

	version := b.Version()

	err = b.Create()
	if err != nil {
		return microerror.Maskf(err, "Etcd %s creation failed: %s", version, err)
	}

	err = b.Encrypt()
	if err != nil {
		microerror.Maskf(err, "Etcd %s encryption failed: %s", version, err)
	}

	err = b.Upload()
	if err != nil {
		microerror.Maskf(err, "Etcd %s upload failed: %s", version, err)
	}
	return nil
}
