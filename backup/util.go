package backup

import (
	"bytes"
	"io/ioutil"
	"log"
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
func execCmd(cmd string, args []string, envs []string) ([]byte, error) {
	log.Printf("Executing: %s %v\n", cmd, args)

	// Create cmd and add environment.
	c := exec.Command(cmd, args...)
	c.Env = append(os.Environ(), envs...)

	// Execute and get output.
	stdOutErr, err := c.CombinedOutput()
	if err != nil {
		log.Printf("%s", stdOutErr)
		log.Print(err)
		return stdOutErr, microerror.Mask(err)
	}
	return stdOutErr, nil
}

// Uploads file to S3 bucket.
// Arguments:
// - fpath - full path to target file
// - p     - paramsAWS struct with AWS keys and bucket name
func uploadToS3(fpath string, p config.AWSConfig) error {
	// Login to AWS S3
	creds := credentials.NewStaticCredentials(p.AccessKey, p.SecretKey, "")
	_, err := creds.Get()
	if err != nil {
		return microerror.Mask(err)
	}
	cfg := aws.NewConfig().WithRegion(p.Region).WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	// Upload.
	file, err := os.Open(fpath)
	if err != nil {
		return microerror.Mask(err)
	}
	defer file.Close()

	// Get file size.
	fileInfo, err := file.Stat()
	if err != nil {
		return microerror.Mask(err)
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
		return microerror.Mask(err)
	}

	log.Printf("AWS S3: object %s successfully uploaded to bucket %s", path, p.Bucket)
	return nil
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
