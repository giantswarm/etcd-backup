package etcd

import (
	"log"
	"path/filepath"

	"github.com/giantswarm/etcd-backup/config"
	"github.com/giantswarm/microerror"
	"github.com/mholt/archiver"
)

type EtcdBackupV2 struct {
	Aws      config.AWSConfig
	Prefix   string
	Filename string
	Datadir  string
	EncPass  string
	TmpDir   string
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
		"--etcd-dir", filepath.Join(b.TmpDir, b.Filename),
	}

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

	log.Print("Etcd v2 etcd created successfully")
	return nil
}

func (b *EtcdBackupV2) Encrypt() error {
	if b.EncPass == "" {
		log.Print("No passphrase provided. Skipping etcd v2 etcd encryption")
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

	log.Print("Etcd v2 etcd encrypted successfully")
	return nil
}

// Upload resulted etcd to S3.
func (b *EtcdBackupV2) Upload() error {
	fpath := filepath.Join(b.TmpDir, b.Filename)

	// Upload
	err := uploadToS3(fpath, b.Aws)
	if err != nil {
		return microerror.Mask(err)
	}

	log.Print("Etcd v2 etcd uploaded successfully")
	return nil
}

func (b *EtcdBackupV2) Version() string {
	return "v2"
}
