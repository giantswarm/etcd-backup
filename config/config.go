package config

import (
	"github.com/giantswarm/microerror"
	"log"
	"os"
)

const (
	EnvAwsAccessKey  = "ETCDBACKUP_AWS_ACCESS_KEY"
	EnvAwsSecretKey  = "ETCDBACKUP_AWS_SECRET_KEY"
	EnvEncryptPassph = "ETCDBACKUP_PASSPHRASE"
)

//AWS config
type AWSConfig struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
}

// Initialize parameters.

type Flags struct {
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

// parse
func ParseEnvs(f *Flags) {
	f.AwsAccessKey = os.Getenv(EnvAwsAccessKey)
	f.AwsSecretKey = os.Getenv(EnvAwsSecretKey)
	f.EncryptPass = os.Getenv(EnvEncryptPassph)
}

func CheckConfig(f Flags) error {
	// Validate parameters.
	// Prefix is required.
	if f.Prefix == "" {
		log.Fatalf("-prefix required")
		return microerror.Mask(invalidConfigError)
	}

	// AWS is requirement.
	if f.AwsAccessKey == "" || f.AwsSecretKey == "" {
		log.Fatalf("No environment variables %s and %s provided", EnvAwsAccessKey, EnvAwsSecretKey)
		return microerror.Mask(invalidConfigError)
	}

	// Skip V2 etcd if not datadir provided.
	if f.EtcdV2DataDir == "" {
		f.SkipV2 = true
		log.Print("Skipping etcd V2 etcd as -etcd-v2-datadir is not set")
		return microerror.Mask(invalidConfigError)
	}

	return nil
}
