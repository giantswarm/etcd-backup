package main

import (
	"log"
	"path/filepath"

	"github.com/giantswarm/microerror"
	"github.com/mholt/archiver"
)

type etcdBackupV3 struct {
	aws       paramsAWS
	prefix    string
	fname     string
	cert      string
	cacert    string
	key       string
	endpoints string
	encPass   string
}

// Create backup in temporary directory.
func (b *etcdBackupV3) create() error {
	// Filename
	b.fname = b.prefix + "-etcd-backup-v3-" + getTimeStamp() + dbExt

	// Full path to file.
	fpath := filepath.Join(tmpDir, b.fname)

	etcdctlEnvs := []string{"ETCDCTL_API=3"}
	etcdctlArgs := []string{
		"snapshot",
		"save",
		fpath,
	}

	if b.endpoints != "" {
		etcdctlArgs = append(etcdctlArgs, "--endpoints", b.endpoints)
	}
	if b.cacert != "" {
		etcdctlArgs = append(etcdctlArgs, "--cacert", b.cacert)
	}
	if b.cert != "" {
		etcdctlArgs = append(etcdctlArgs, "--cert", b.cert)
	}
	if b.key != "" {
		etcdctlArgs = append(etcdctlArgs, "--key", b.key)
	}

	// Create a backup.
	_, err := execCmd(etcdctlCmd, etcdctlArgs, etcdctlEnvs)
	if err != nil {
		return microerror.Mask(err)
	}

	// Create tar.gz.
	err = archiver.TarGz.Make(fpath+tgzExt, []string{fpath})
	if err != nil {
		return microerror.Mask(err)
	}

	// Update fname in backup object.
	b.fname = b.fname + tgzExt

	log.Print("Etcd v3 backup created successfully")
	return nil
}

func (b *etcdBackupV3) encrypt() error {
	if b.encPass == "" {
		log.Print("No passphrase provided. Skipping etcd v3 backup encryption")
		return nil
	}

	// Full path to file.
	fpath := filepath.Join(tmpDir, b.fname)

	// Encrypt backup.
	err := encryptFile(fpath, fpath+encExt, b.encPass)
	if err != nil {
		return microerror.Mask(err)
	}

	// Update fname in backup object.
	b.fname = b.fname + encExt

	log.Print("Etcd v3 backup encrypted successfully")
	return nil
}

// Upload resulted backup to S3.
func (b *etcdBackupV3) upload() error {
	fpath := filepath.Join(tmpDir, b.fname)

	// Upload.
	err := uploadToS3(fpath, b.aws)
	if err != nil {
		return microerror.Mask(err)
	}

	log.Print("Etcd v3 backup uploaded successfully")
	return nil
}
