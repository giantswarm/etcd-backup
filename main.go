package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/etcd-backup/backup"
	"github.com/giantswarm/etcd-backup/config"
)

// TODO:
// - check etcdctl exists and right version
const backupFailedCode = 1

// Common variables.
var (
	description string = "Application to etcd etcd."
	gitCommit   string = "n/a"
	name        string = "etcd-etcd"
	source      string = "https://github.com/giantswarm/etcd-etcd"
)

var (
	f config.Flags
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
	flag.StringVar(&f.EtcdV2DataDir, "etcd-v2-datadir", "", "Etcd datadir. If not set V2 etcd will be skipped")
	flag.StringVar(&f.EtcdV3Cert, "etcd-v3-cert", "", "Client certificate for etcd connection")
	flag.StringVar(&f.EtcdV3CACert, "etcd-v3-cacert", "", "Client CA certificate for etcd connection")
	flag.StringVar(&f.EtcdV3Key, "etcd-v3-key", "", "Client private key for etcd connection")
	flag.StringVar(&f.EtcdV3Endpoints, "etcd-v3-endpoints", "http://127.0.0.1:2379", "Endpoints for etcd connection")
	flag.StringVar(&f.Prefix, "prefix", "", "[mandatory] Prefix to use in etcd filenames")

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
	config.ParseEnvs(&f)

	// Print usage.
	if f.Help {
		flag.Usage()
		return
	}

	// check flags
	config.CheckConfig(f)
	// create micrologger
	loggerConfig := micrologger.Config{}
	logger, err := micrologger.New(loggerConfig)

	// create backup service
	backupService := backup.CreateService(f, logger)

	// backup host cluster
	err = backupService.BackupHostCluster()
	if err != nil {
		logger.Log("level", "error", "msg", "failed to backup host cluster etcd", "reason", err)
		os.Exit(backupFailedCode)
	}
	// backup guest cluster
	err = backupService.BackupGuestClusters()
	if err != nil {
		logger.Log("level", "error", "msg", "failed to backup guest cluster etcd", "reason", err)
		os.Exit(backupFailedCode)
	}
	logger.Log("level", "info", "msg", "Success")
}
