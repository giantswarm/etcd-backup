package backup

import (
	"github.com/giantswarm/etcd-backup/config"
	"github.com/giantswarm/etcd-backup/etcd"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"log"
)

type Service struct {
	Logger micrologger.Logger

	Prefix          string
	EtcdV2DataDir   string
	EtcdV3Cert      string
	EtcdV3CACert    string
	EtcdV3Key       string
	EtcdV3Endpoints string
	AwsAccessKey    string
	AwsSecretKey    string
	AwsS3Bucket     string
	AwsS3Region     string
	EncryptPass     string
	Help            bool
	Provider        string
	SkipV2          bool
}

func CreateService(f config.Flags, logger micrologger.Logger) *Service {
	s := &Service{
		Logger: logger,

		Prefix:          f.Prefix,
		EtcdV2DataDir:   f.EtcdV2DataDir,
		EtcdV3CACert:    f.EtcdV3CACert,
		EtcdV3Cert:      f.EtcdV3Cert,
		EtcdV3Key:       f.EtcdV3Key,
		EtcdV3Endpoints: f.EtcdV3Endpoints,
		AwsAccessKey:    f.AwsAccessKey,
		AwsSecretKey:    f.AwsSecretKey,
		AwsS3Bucket:     f.AwsS3Bucket,
		AwsS3Region:     f.AwsS3Region,
		EncryptPass:     f.EncryptPass,
		Provider:        f.Provider,
	}
	return s
}

// backup host cluster etcd
func (s *Service) BackupHostCluster() error {
	var err error
	// temporary directory for files
	tmpDir, err := CreateTMPDir()
	if err != nil {
		return microerror.Maskf(err, "Failed to create temporary directory: %s", err)
	}
	defer ClearTMPDir(tmpDir)

	// V2 etcd.
	if !s.SkipV2 {
		v2 := etcd.EtcdBackupV2{
			Aws: config.AWSConfig{
				AccessKey: s.AwsAccessKey,
				SecretKey: s.AwsSecretKey,
				Bucket:    s.AwsS3Bucket,
				Region:    s.AwsS3Region,
			},
			Datadir: s.EtcdV2DataDir,
			EncPass: s.EncryptPass,
			Prefix:  s.Prefix,
			TmpDir:  tmpDir,
		}
		// run backup task
		err = etcd.FullBackup(&v2)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// V3 etcd.
	v3 := etcd.EtcdBackupV3{
		Aws: config.AWSConfig{
			AccessKey: s.AwsAccessKey,
			SecretKey: s.AwsSecretKey,
			Bucket:    s.AwsS3Bucket,
			Region:    s.AwsS3Region,
		},
		CACert:    s.EtcdV3CACert,
		Cert:      s.EtcdV3Cert,
		Prefix:    s.Prefix,
		EncPass:   s.EncryptPass,
		Endpoints: s.EtcdV3Endpoints,
		Key:       s.EtcdV3Key,
		TmpDir:    tmpDir,
	}
	// run backup task
	err = etcd.FullBackup(&v3)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// backup all guest clusters etcd
func (s *Service) BackupGuestClusters() error {
	tmpDir, err := CreateTMPDir()
	if err != nil {
		return microerror.Mask(err)
	}
	defer ClearTMPDir(tmpDir)
	// create host cluster k8s client
	k8sClient, err := CreateK8sClient(s.Logger)
	if err != nil {
		return microerror.Mask(err)
	}
	// create k8s crd client
	crdClient, err := CreateCRDClient(s.Logger)
	if err != nil {
		return microerror.Mask(err)
	}
	// fetch all guest cluster ids
	clusterList, err := GetAllGuestClusters(s.Provider, crdClient)
	if err != nil {
		return microerror.Mask(err)
	}

	// backup failed flag, we want to know if any of the backup failed,
	// but  one failed guest cluster should not cancel backup of the rest
	failed := false

	// iterate over all clusters
	for _, clusterID := range clusterList {
		// fetch etcd certs
		certs, err := FetchCerts(clusterID, k8sClient)
		if err != nil {
			failed = true
			s.Logger.Log("level", "error", "msg", "Failed to fetch etcd certs for cluster "+clusterID, "reason", err)
			continue
		}
		// write etcd certs to tmpdir
		err = CreateCertFiles(clusterID, certs, tmpDir)
		if err != nil {
			failed = true
			s.Logger.Log("level", "error", "msg", "Failed to write etcd certs to tmpdir for cluster "+clusterID, "reason", err)
			continue
		}

		// fetch etcd endpoint
		etcdEndpoint, err := GetEtcdEndpoint(clusterID, k8sClient)
		if err != nil {
			failed = true
			s.Logger.Log("level", "error", "msg", "Failed to fetch etcd endpoint for cluster "+clusterID, "reason", err)
			continue
		}
		// backup config, we only care about etcd3 in guest cluster
		backupConfig := etcd.EtcdBackupV3{
			Aws: config.AWSConfig{
				AccessKey: s.AwsAccessKey,
				SecretKey: s.AwsSecretKey,
				Bucket:    s.AwsS3Bucket,
				Region:    s.AwsS3Region,
			},
			CACert: certs.CAFile,
			Cert:   certs.CrtFile,
			Key:    certs.KeyFile,

			Prefix:    BackupPrefix(clusterID),
			EncPass:   s.EncryptPass,
			Endpoints: etcdEndpoint,

			TmpDir: tmpDir,
		}
		// run backup task
		err = etcd.FullBackup(&backupConfig)
		if err != nil {
			failed = true
			s.Logger.Log("level", "error", "msg", "Failed to backup etcd cluster "+clusterID, "reason", err)
			continue
		}

	}
	// check if any backup failed
	if failed {
		return failedBackupError
	} else {
		log.Printf("Finished guest cluster backup. Total guest clusters: %d", len(clusterList))
	}

	return nil
}
