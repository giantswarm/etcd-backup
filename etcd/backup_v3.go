package etcd

import (
	"log"
	"path/filepath"

	"github.com/giantswarm/etcd-backup/config"
	"github.com/giantswarm/microerror"
	"github.com/mholt/archiver"
)

type EtcdBackupV3 struct {
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
	b.Filename = b.Prefix + "-etcd-etcd-v3-" + getTimeStamp() + dbExt

	// Full path to file.
	fpath := filepath.Join(b.TmpDir, b.Filename)

	etcdctlEnvs := []string{"ETCDCTL_API=3"}
	etcdctlArgs := []string{
		"snapshot",
		"save",
		fpath,
	}

	if b.Endpoints != "" {
		etcdctlArgs = append(etcdctlArgs, "--Endpoints", b.Endpoints)
	}
	if b.CACert != "" {
		etcdctlArgs = append(etcdctlArgs, "--CACert", b.CACert)
	}
	if b.Cert != "" {
		etcdctlArgs = append(etcdctlArgs, "--Cert", b.Cert)
	}
	if b.Key != "" {
		etcdctlArgs = append(etcdctlArgs, "--Key", b.Key)
	}

	// Create a etcd.
	_, err := execCmd(etcdctlCmd, etcdctlArgs, etcdctlEnvs)
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

	log.Print("Etcd v3 etcd created successfully")
	return nil
}

func (b *EtcdBackupV3) Encrypt() error {
	if b.EncPass == "" {
		log.Print("No passphrase provided. Skipping etcd v3 etcd encryption")
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

	log.Print("Etcd v3 etcd encrypted successfully")
	return nil
}

// Upload resulted etcd to S3.
func (b *EtcdBackupV3) Upload() error {
	fpath := filepath.Join(b.TmpDir, b.Filename)

	// Upload.
	err := uploadToS3(fpath, b.Aws)
	if err != nil {
		return microerror.Mask(err)
	}

	log.Print("Etcd v3 etcd uploaded successfully")
	return nil
}

func (b *EtcdBackupV3) Version() string {
	return "v3"
}
