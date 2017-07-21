package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// TODO:
// - check etcdctl exists and right version

const (
	envAwsAccessKey  = "ETCDBACKUP_AWS_ACCESS_KEY"
	envAwsSecretKey  = "ETCDBACKUP_AWS_SECRET_KEY"
	envEncryptPassph = "ETCDBACKUP_PASSPHRASE"
	etcdctlCmd       = "etcdctl"
	awsCmd           = "aws"
	tgzExt           = ".tar.gz"
	gpgExt           = ".gpg"
	dbExt            = ".db"
)

var (
	tmpDir string
	skipV2 bool = false
)

type backup interface {
	create()  error
	encrypt() error
	upload()  error
}

type paramsAWS struct {
	accessKey string
	secretKey string
	bucket    string
}

type etcdBackupV2 struct {
	aws     paramsAWS
	prefix  string
	fname   string
	datadir string
}

type etcdBackupV3 struct {
	aws       paramsAWS
	prefix    string
	fname     string
	cert      string
	cacert    string
	key       string
	endpoints string
}

func main() {
	// Initialize parameters
	flags := struct {
		prefix          string
		etcdV2DataDir   string
		etcdV3Cert      string
		etcdV3CACert    string
		etcdV3Key       string
		etcdV3Endpoints string
		awsAccessKey    string
		awsSecretKey    string
		awsS3Bucket     string
		encryptPassph   string
	}{}

	flag.StringVar(&flags.prefix, "prefix", "", "[mandatory] Prefix to use in backup filenames")
	flag.StringVar(&flags.etcdV2DataDir, "etcd-v2-datadir", "", "Etcd datadir. If not set V2 backup will be skipped")
	flag.StringVar(&flags.etcdV3Cert, "etcd-v3-cert", "", "Client certificate for etcd connection")
	flag.StringVar(&flags.etcdV3CACert, "etcd-v3-cacert", "", "Client CA certificate for etcd connection")
	flag.StringVar(&flags.etcdV3Key, "etcd-v3-key", "", "Client private key for etcd connection")
	flag.StringVar(&flags.etcdV3Endpoints, "etcd-v3-endpoints", "http://127.0.0.1:2379", "Endpoints for etcd connection")
	flag.StringVar(&flags.awsS3Bucket, "aws-s3-bucket", "etcdbackups", "AWS S3 bucket for backups")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  variable %s - [mandatory] AWS access key for S3\n", envAwsAccessKey)
		fmt.Fprintf(os.Stderr, "  variable %s - [mandatory] AWS secret access key for S3\n", envAwsSecretKey)
		fmt.Fprintf(os.Stderr, "\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	flags.awsAccessKey  = os.Getenv(envAwsAccessKey)
	flags.awsSecretKey  = os.Getenv(envAwsSecretKey)
	flags.encryptPassph = os.Getenv(envEncryptPassph)

	// Validate parameters
	// Prefix is required
	if flags.prefix == "" {
		log.Fatalf("-prefix required")
	}

	// AWS is requrement
	if flags.awsAccessKey == "" || flags.awsSecretKey == "" {
		log.Fatalf("No environment variables %s and %s provided", envAwsAccessKey, envAwsSecretKey)
	}

	// Skip V2 backup if not datadir provided
	if flags.etcdV2DataDir == "" {
		skipV2 = true
		log.Print("Skipping etcd V2 backup as -etcd-v2-datadir is not set")
	}

	// Create tempDir where all file related magic happens
	var err error
	tmpDir, err = ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Created temporary directory: ", tmpDir)

	defer os.RemoveAll(tmpDir) // clean up after finished

	// V2 backup
	if !skipV2 {
		v2 := etcdBackupV2{
			aws:    paramsAWS{
				accessKey: flags.awsAccessKey,
				secretKey: flags.awsSecretKey,
				bucket:    flags.awsS3Bucket,
			},
			prefix:  flags.prefix,
			datadir: flags.etcdV2DataDir,
		}

		err = v2.create()
		if err != nil {
			log.Fatal("Etcd v2 backup creation failed: ", err)
		}

		err = v2.encrypt()
		if err != nil {
			log.Fatal("Etcd v2 backup encryption failed: ", err)
		}

		err = v2.upload()
		if err != nil {
			log.Fatal("Etcd v2 backup upload failed: ", err)
		}
	}

	// V3 backup
	v3 := etcdBackupV3{
		aws:    paramsAWS{
			accessKey: flags.awsAccessKey,
			secretKey: flags.awsSecretKey,
			bucket:    flags.awsS3Bucket,
		},
		prefix:    flags.prefix,
		cert:      flags.etcdV3Cert,
		cacert:    flags.etcdV3CACert,
		key:       flags.etcdV3Key,
		endpoints: flags.etcdV3Endpoints,
	}

	err = v3.create()
	if err != nil {
		log.Fatal("Etcd v3 backup creation failed: ", err)
	}

	err = v3.encrypt()
	if err != nil {
		log.Fatal("Etcd v3 backup encryption failed: ", err)
	}

	err = v3.upload()
	if err != nil {
		log.Fatal("Etcd v3 backup upload failed: ", err)
	}

	log.Print("Success")
}
