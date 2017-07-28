package main

import (
	"log"
	"path/filepath"

	microerror "github.com/giantswarm/microkit/error"
)

type etcdBackupV3 struct {
	aws       paramsAWS
	prefix    string
	fname     string
	cert      string
	cacert    string
	key       string
	endpoints string
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
		return microerror.MaskAny(err)
	}

	log.Print("Etcd v3 backup created successfully")
	return nil
}

func (b *etcdBackupV3) encrypt() error {
	log.Print("Etcd v3 encryption is not implemented. Skipping")
	return nil
}

// Upload resulted backup to S3.
func (b *etcdBackupV3) upload() error {
	fpath := filepath.Join(tmpDir, b.fname)

	// Upload.
	err := uploadToS3(fpath, b.aws)
	if err != nil {
		return microerror.MaskAny(err)
	}

	log.Print("Etcd v3 backup uploaded successfully")
	return nil
}
