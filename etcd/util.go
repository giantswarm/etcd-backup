package etcd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/etcd-backup/config"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"golang.org/x/crypto/openpgp"
)

const (
	etcdctlCmd = "etcdctl"
	awsCmd     = "Aws"
	tgzExt     = ".tar.gz"
	encExt     = ".enc"
	dbExt      = ".db"
)

// Outputs timestamp.
func getTimeStamp() string {
	return time.Now().Format("2006-01-02T15-04-05")
}

// Executes command and outputs stdout+stderr and error if any.
// Arguments:
// - cmd  - command to execute
// - args - arguments for command
// - envs - envronment variables
func execCmd(cmd string, args []string, envs []string, logger micrologger.Logger) ([]byte, error) {
	logger.Log("level", "info", "msg", fmt.Sprintf("Executing: %s %v", cmd, args))

	// Create cmd and add environment.
	c := exec.Command(cmd, args...)
	c.Env = append(os.Environ(), envs...)

	// Execute and get output.
	stdOutErr, err := c.CombinedOutput()
	if err != nil {
		logger.Log("level", "error", "msg", "execCmd failed", "reason", fmt.Sprintf("%s", stdOutErr), "err", err)
		return stdOutErr, microerror.Mask(err)
	}
	return stdOutErr, nil
}

// Uploads file to S3 bucket.
// Arguments:
// - fpath - full path to target file
// - p     - paramsAWS struct with AWS keys and bucket name
func uploadToS3(fpath string, p config.AWSConfig, logger micrologger.Logger) (int64, error) {
	// Login to AWS S3
	creds := credentials.NewStaticCredentials(p.AccessKey, p.SecretKey, "")
	_, err := creds.Get()
	if err != nil {
		return -1, microerror.Mask(err)
	}
	cfg := aws.NewConfig().WithRegion(p.Region).WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	// Upload.
	file, err := os.Open(fpath)
	if err != nil {
		return -1, microerror.Mask(err)
	}
	defer file.Close()

	// Get file size.
	fileInfo, err := file.Stat()
	if err != nil {
		return -1, microerror.Mask(err)
	}
	size := fileInfo.Size()

	// Get filename without path.
	path := filepath.Base(fileInfo.Name())

	params := &s3.PutObjectInput{
		Bucket:        aws.String(p.Bucket),
		Key:           aws.String(path),
		Body:          file,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String("application/octet-stream"),
	}

	// Put object to S3.
	_, err = svc.PutObject(params)
	if err != nil {
		return -1, microerror.Mask(err)
	}

	logger.Log("level", "info", "msg", fmt.Sprintf("AWS S3: object %s successfully uploaded to bucket %s", path, p.Bucket))

	pms := &s3.GetObjectInput{
		Bucket: aws.String(p.Bucket),
		Key:    aws.String(path),
	}

	obj, err := svc.GetObject(pms)
	if err != nil {
		return -1, microerror.Mask(err)
	}

	return *obj.ContentLength, nil
}

// Encrypt data with passphrase.
func encryptData(value []byte, pass string) (ciphertext []byte, err error) {
	buf := bytes.NewBuffer(nil)

	encrypter, err := openpgp.SymmetricallyEncrypt(buf, []byte(pass), nil, nil)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	_, err = encrypter.Write(value)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	encrypter.Close()

	return buf.Bytes(), nil
}

// Encrypts file from srcPath and writes encrypted data to dstPart.
func encryptFile(srcPath string, dstPart string, passphrase string) error {
	data, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return microerror.Mask(err)
	}

	encData, err := encryptData(data, passphrase)
	if err != nil {
		return microerror.Mask(err)
	}

	err = ioutil.WriteFile(dstPart, encData, os.FileMode(0600))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

var (
	creationTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "etcd_backup_creation_time_ms",
		Help: "Time in ms that che backup creation process took.",
	})
	encryptionTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "etcd_backup_encryption_time_ms",
		Help: "Time in ms that che backup encryption process took.",
	})
	uploadTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "etcd_backup_upload_time_ms",
		Help: "Time in ms that che backup upload process took.",
	})
	backupSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "etcd_backup_size",
		Help: "The size in bytes of the backup file.",
	})
)

func sendMetrics(prometheusConfig PrometheusConfig, creationTimeMeasurement int64, encryptionTimeMeasurement int64, uploadTimeMeasurement int64, backupSizeMeasurement int64) error {
	// prometheus URL might be empty, in that case we can't push any metric
	if prometheusConfig.Url != "" {
		registry := prometheus.NewRegistry()
		registry.MustRegister(creationTime, encryptionTime, uploadTime, backupSize)

		pusher := push.New(prometheusConfig.Url, prometheusConfig.Job).Gatherer(registry)

		creationTime.Set(float64(creationTimeMeasurement))
		encryptionTime.Set(float64(encryptionTimeMeasurement))
		uploadTime.Set(float64(uploadTimeMeasurement))
		backupSize.Set(float64(backupSizeMeasurement))

		// Add is used here rather than Push to not delete a previously pushed
		// success timestamp in case of a failure of this backup.
		if err := pusher.Add(); err != nil {
			fmt.Println("Could not push to Pushgateway:", err)
		}
	}

	return nil
}
