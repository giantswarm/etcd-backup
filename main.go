package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"github.com/giantswarm/etcd-backup/backup"
	"github.com/giantswarm/etcd-backup/config"
)

// TODO:
// - check etcdctl exists and right version

// Common variables.
var (
	description string = "Application to backup etcd."
	gitCommit   string = "n/a"
	name        string = "etcd-backup"
	source      string = "https://github.com/giantswarm/etcd-backup"
)

var (
	tmpDir string
	f      config.Flags
)

func main() {
	// Print version.
	// This is only for compatibility until switching to microkit.
	if (len(os.Args) > 1) && (os.Args[1] == "version") {
		fmt.Printf("Description:    %s\n", description)
		fmt.Printf("Git Commit:     %s\n", gitCommit)
		fmt.Printf("Go Version:     %s\n", runtime.Version())
		fmt.Printf("Name:           %s\n", name)
		fmt.Printf("OS / Arch:      %s / %s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Source:         %s\n", source)
		return
	}

	// Print flags related messages to stdout instead of stderr.
	flag.CommandLine.SetOutput(os.Stdout)

	flag.StringVar(&f.AwsS3Bucket, "aws-s3-bucket", "etcdbackups", "AWS S3 bucket for backups")
	flag.StringVar(&f.AwsS3Region, "aws-s3-region", "us-east-1", "AWS S3 region for backups")
	flag.StringVar(&f.EtcdV2DataDir, "etcd-v2-datadir", "", "Etcd datadir. If not set V2 backup will be skipped")
	flag.StringVar(&f.EtcdV3Cert, "etcd-v3-cert", "", "Client certificate for etcd connection")
	flag.StringVar(&f.EtcdV3CACert, "etcd-v3-cacert", "", "Client CA certificate for etcd connection")
	flag.StringVar(&f.EtcdV3Key, "etcd-v3-key", "", "Client private key for etcd connection")
	flag.StringVar(&f.EtcdV3Endpoints, "etcd-v3-endpoints", "http://127.0.0.1:2379", "Endpoints for etcd connection")
	flag.StringVar(&f.Prefix, "prefix", "", "[mandatory] Prefix to use in backup filenames")

	flag.BoolVar(&f.Help, "help", false, "Print usage and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "  variable %s - [mandatory] AWS access key for S3\n", config.EnvAwsAccessKey)
		fmt.Fprintf(os.Stdout, "  variable %s - [mandatory] AWS secret access key for S3\n", config.EnvAwsSecretKey)
		fmt.Fprintf(os.Stdout, "  variable %s - passphrase for AES encryption\n", config.EnvEncryptPassph)
		fmt.Fprintf(os.Stdout, "\n")
		flag.PrintDefaults()
	}
	// parse flags
	flag.Parse()
	config.ParseEnvs(f)

	// check flags
	config.CheckConfig(f)

	// Print usage.
	if f.Help {
		flag.Usage()
		return
	}

	// Create tempDir where all file related magic happens.
	var err error
	tmpDir, err = ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Created temporary directory: ", tmpDir)

	defer os.RemoveAll(tmpDir) // clean up after finished.

	// V2 backup.
	if !f.SkipV2 {
		v2 := backup.EtcdBackupV2{
			Aws: config.AWSConfig{
				AccessKey: f.AwsAccessKey,
				SecretKey: f.AwsSecretKey,
				Bucket:    f.AwsS3Bucket,
				Region:    f.AwsS3Region,
			},
			Datadir: f.EtcdV2DataDir,
			EncPass: f.EncryptPass,
			Prefix:  f.Prefix,
			TmpDir:  tmpDir,
		}

		err = v2.Create()
		if err != nil {
			log.Fatal("Etcd v2 backup creation failed: ", err)
		}

		err = v2.Encrypt()
		if err != nil {
			log.Fatal("Etcd v2 backup encryption failed: ", err)
		}

		err = v2.Upload()
		if err != nil {
			log.Fatal("Etcd v2 backup upload failed: ", err)
		}
	}

	// V3 backup.
	v3 := backup.EtcdBackupV3{
		Aws: config.AWSConfig{
			AccessKey: f.AwsAccessKey,
			SecretKey: f.AwsSecretKey,
			Bucket:    f.AwsS3Bucket,
			Region:    f.AwsS3Region,
		},
		CACert:    f.EtcdV3CACert,
		Cert:      f.EtcdV3Cert,
		Prefix:    f.Prefix,
		EncPass:   f.EncryptPass,
		Endpoints: f.EtcdV3Endpoints,
		Key:       f.EtcdV3Key,
	}

	err = v3.Create()
	if err != nil {
		log.Fatal("Etcd v3 backup creation failed: ", err)
	}

	err = v3.Encrypt()
	if err != nil {
		log.Fatal("Etcd v3 backup encryption failed: ", err)
	}

	err = v3.Upload()
	if err != nil {
		log.Fatal("Etcd v3 backup upload failed: ", err)
	}

	log.Print("Success")
}
