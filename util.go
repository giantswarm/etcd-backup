package main

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	microerror "github.com/giantswarm/microkit/error"
)

// Outputs timestamp
func getTimeStamp() (string) {
	return time.Now().Format("2006-01-02T15-04-05")
}

// Executes command and outputs stdout+stderr and error if any
// Arguments:
// - cmd  - command to execute
// - args - arguments for command
// - envs - envronment variables
func execCmd(cmd string, args []string, envs []string) ([]byte, error) {
	log.Printf("Executing: %s %v\n", cmd, args)

	// Create cmd and add environment
	c := exec.Command(cmd, args...)
	c.Env = append(os.Environ(), envs...)

	// Execute and get output
	stdOutErr, err := c.CombinedOutput()
	if err != nil {
		log.Printf("%s", stdOutErr)
		log.Print(err)
		return stdOutErr, microerror.MaskAny(err)
	}
	return stdOutErr, nil
}

// Uploads file to S3 bucket
// Arguments:
// - fpath - full path to target file
// - p     - paramsAWS struct with AWS keys and bucket name
func uploadToS3(fpath string, p paramsAWS) error {
	// Login to AWS S3
	creds := credentials.NewStaticCredentials(p.accessKey, p.secretKey, "")
	_, err := creds.Get()
	if err != nil {
		return microerror.MaskAny(err)
	}
	cfg := aws.NewConfig().WithRegion(p.region).WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	// Upload
	file, err := os.Open(fpath)
	if err != nil {
		return microerror.MaskAny(err)
	}
	defer file.Close()

	// Get file size
	fileInfo, _ := file.Stat()
	if err != nil {
		return microerror.MaskAny(err)
	}
	size := fileInfo.Size()

	// Read file content to buffer
	buffer := make([]byte, size)
	file.Read(buffer)
	if err != nil {
		return microerror.MaskAny(err)
	}

	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)
	// Get filename without path
	path := filepath.Base(file.Name())

	params := &s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key: aws.String(path),
		Body: fileBytes,
		ContentLength: aws.Int64(size),
		ContentType: aws.String(fileType),
	}

	// Put object to S3
	_, err = svc.PutObject(params)
	if err != nil {
		return microerror.MaskAny(err)
	}

	log.Printf("AWS S3: object %s successfully uploaded to bucket %s", path, p.bucket)
	return nil
}
