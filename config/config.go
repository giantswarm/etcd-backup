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
	Bucket    string
	Region    string
	SecretKey string
}

// Push gateway address
type PrometheusConfig struct {
	Job string
	Url string
}

// Initialize parameters.

type Flags struct {
	AwsAccessKey    string
	AwsSecretKey    string
	AwsS3Bucket     string
	AwsS3Region     string
	EtcdV2DataDir   string
	EtcdV3Cert      string
	EtcdV3CACert    string
	EtcdV3Key       string
	EtcdV3Endpoints string
	EncryptPass     string
	GuestBackup     bool
	Help            bool
	Prefix          string
	Provider        string
	PushGatewayURL  string
	PushGatewayJob  string
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

	// check that the Prometheus Url, if present, is a valid URL
	if f.PushGatewayURL != "" {
		_, err := url.ParseRequestURI(f.PushGatewayURL)
		if err != nil {
			log.Fatalf("--prometheus-url is invalid")
			return microerror.Mask(invalidConfigError)
		}

		if f.PushGatewayJob == "" {
			log.Fatalf("--prometheus-job is mandatory when --prometheus-url is set")
			return microerror.Mask(invalidConfigError)
		}
	} else {
		log.Print("Skipping prometheus metrics push as --prometheus-url is not set")
	}

	// Skip V2 etcd if not datadir provided.
	if f.EtcdV2DataDir == "" {
		f.SkipV2 = true
		log.Print("Skipping etcd V2 etcd as -etcd-v2-datadir is not set")
		return microerror.Mask(invalidConfigError)
	}

	return nil
}
