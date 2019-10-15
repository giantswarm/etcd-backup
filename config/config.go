package config

import (
	"log"
	"net/url"
	"os"

	"github.com/giantswarm/microerror"
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
	GuestBackup     bool
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
	PrometheusUrl   string
	PrometheusJob   string
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

	// check that the Prometheus Url is a valid URL
	if f.PrometheusUrl != "" {
		_, err := url.ParseRequestURI(f.PrometheusUrl)
		if err != nil {
			log.Fatalf("--prometheus-url is invalid")
			return microerror.Mask(invalidConfigError)
		}

		if f.PrometheusJob == "" {
			log.Fatalf("--prometheus-job is mandatory when --prometheus-url is set")
			return microerror.Mask(invalidConfigError)
		}
	} else {
		log.Print("Skipping prometheus metrics push as --prometheus-url is not set")
	}

	return nil
}
