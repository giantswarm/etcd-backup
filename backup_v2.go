package main

import (
	"log"
	"path/filepath"

	"github.com/giantswarm/microerror"
	"github.com/mholt/archiver"
)

type etcdBackupV2 struct {
	aws     paramsAWS
	prefix  string
	fname   string
	datadir string
	encPass string
}

// Create backup in temporary directory, tar and compress.
func (b *etcdBackupV2) create() error {
	// Filename
	b.fname = b.prefix + "-etcd-backup-v2-" + getTimeStamp()

	// Full path to file.
	fpath := filepath.Join(tmpDir, b.fname)

	// Create a backup.
	etcdctlEnvs := []string{}
	etcdctlArgs := []string{
		"backup",
		"--data-dir", b.datadir,
		"--backup-dir", filepath.Join(tmpDir, b.fname),
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

	// Update fname in backup object.
	b.fname = b.fname + tgzExt

	log.Print("Etcd v2 backup created successfully")
	return nil
}

func (b *etcdBackupV2) encrypt() error {
	if b.encPass == "" {
		log.Print("No passphrase provided. Skipping etcd v2 backup encryption")
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

	log.Print("Etcd v2 backup encrypted successfully")
	return nil
}

// Upload resulted backup to S3.
func (b *etcdBackupV2) upload() error {
	fpath := filepath.Join(tmpDir, b.fname)

	// Upload
	err := uploadToS3(fpath, b.aws)
	if err != nil {
		return microerror.Mask(err)
	}

	log.Print("Etcd v2 backup uploaded successfully")
	return nil
}
