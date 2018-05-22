package etcd

import (
	"path/filepath"

	"github.com/giantswarm/etcd-backup/config"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/mholt/archiver"
)

type EtcdBackupV3 struct {
	Logger micrologger.Logger

	Aws       config.AWSConfig
	Prefix    string
	Filename  string
	Cert      string
	CACert    string
	Key       string
	Endpoints string
	EncPass   string
	TmpDir    string
}

// Create etcd in temporary directory.
func (b *EtcdBackupV3) Create() error {
	// Filename
	b.Filename = b.Prefix + "-backup-etcd-v3-" + getTimeStamp() + dbExt

	// Full path to file.
	fpath := filepath.Join(b.TmpDir, b.Filename)

	etcdctlEnvs := []string{"ETCDCTL_API=3"}
	etcdctlArgs := []string{
		"snapshot",
		"save",
		fpath,
	}

	if b.Endpoints != "" {
		etcdctlArgs = append(etcdctlArgs, "--endpoints", b.Endpoints)
	}
	if b.CACert != "" {
		etcdctlArgs = append(etcdctlArgs, "--cacert", b.CACert)
	}
	if b.Cert != "" {
		etcdctlArgs = append(etcdctlArgs, "--cert", b.Cert)
	}
	if b.Key != "" {
		etcdctlArgs = append(etcdctlArgs, "--key", b.Key)
	}

	// Create a etcd.
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

	b.Logger.Log("level", "info", "msg", "Etcd v3 backup created successfully")
	return nil
}
// encrypt backup
func (b *EtcdBackupV3) Encrypt() error {
	if b.EncPass == "" {
		b.Logger.Log("level", "warning", "msg", "No passphrase provided. Skipping etcd v3 backup encryption")
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

	b.Logger.Log("level", "info", "msg", "Etcd v3 backup encrypted successfully")
	return nil
}

// Upload resulted etcd to S3.
func (b *EtcdBackupV3) Upload() error {
	fpath := filepath.Join(b.TmpDir, b.Filename)

	// Upload.
	err := uploadToS3(fpath, b.Aws, b.Logger)
	if err != nil {
		return microerror.Mask(err)
	}

	b.Logger.Log("level", "info", "msg", "Etcd v3 backup uploaded successfully")
	return nil
}

func (b *EtcdBackupV3) Version() string {
	return "v3"
}
