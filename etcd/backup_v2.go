package etcd

import (
	"fmt"
	"path/filepath"

	"github.com/giantswarm/etcd-backup/config"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/mholt/archiver"
)

type EtcdBackupV2 struct {
	Logger micrologger.Logger

	Aws              config.AWSConfig
	Prefix           string
	Filename         string
	Datadir          string
	EncPass          string
	TmpDir           string
	PrometheusConfig PrometheusConfig
}

// Create etcd in temporary directory, tar and compress.
func (b *EtcdBackupV2) Create() error {
	// Filename
	b.Filename = b.Prefix + "-etcd-etcd-v2-" + getTimeStamp()

	// Full path to file.
	fpath := filepath.Join(b.TmpDir, b.Filename)

	// Create a etcd.
	etcdctlEnvs := []string{}
	etcdctlArgs := []string{
		"backup",
		"--data-dir", b.Datadir,
		"--backup-dir", filepath.Join(b.TmpDir, b.Filename),
	}

	_, err := execCmd(etcdctlCmd, etcdctlArgs, etcdctlEnvs, b.Logger)
	if err != nil {
		return microerror.Mask(err)
	}

	// Create tar.gz.
	err = archiver.TarGz.Make(fpath+tgzExt, []string{fpath})
	if err != nil {
		return microerror.Mask(err)
	}

	// Update Filename in etcd object.
	b.Filename = b.Filename + tgzExt

	b.Logger.Log("level", "info", "msg", "Etcd v2 etcd created successfully")
	return nil
}

func (b *EtcdBackupV2) Encrypt() error {
	if b.EncPass == "" {
		b.Logger.Log("level", "warning", "msg", "No passphrase provided. Skipping etcd v2 backup encryption")
		return nil
	}

	// Full path to file.
	fpath := filepath.Join(b.TmpDir, b.Filename)

	// Encrypt etcd.
	err := encryptFile(fpath, fpath+encExt, b.EncPass)
	if err != nil {
		return microerror.Mask(err)
	}

	// Update Filename in etcd object.
	b.Filename = b.Filename + encExt

	b.Logger.Log("level", "info", "msg", "Etcd v2 backup encrypted successfully")
	return nil
}

// Upload resulted etcd to S3.
func (b *EtcdBackupV2) Upload() (int64, error) {
	fpath := filepath.Join(b.TmpDir, b.Filename)

	// Upload
	size, err := uploadToS3(fpath, b.Aws, b.Logger)
	if err != nil {
		return -1, microerror.Mask(err)
	}

	b.Logger.Log("level", "info", "msg", "Etcd v2 backup uploaded successfully")
	return size, nil
}

func (b *EtcdBackupV2) SendMetrics(creationTime int64, encryptionTime int64, uploadTime int64, backupSize int64) error {
	err := sendMetrics(b.PrometheusConfig, creationTime, encryptionTime, uploadTime, backupSize)

	if err != nil {
		b.Logger.Log("level", "error", "msg", fmt.Sprintf("Failed to push metrics to pushgateway: %s", err))
	}

	return err
}

func (b *EtcdBackupV2) Version() string {
	return "v2"
}
