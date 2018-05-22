package backup

import (
	"fmt"
	"github.com/coreos/go-semver/semver"
	"path"
)

const (
	// crd names for each provider
	aws   = "aws"
	azure = "azure"
	kvm   = "kvm"

	// namespace where secrets are located
	secretNamespace = "default"

	// namespace where crds are located
	crdNamespace = "default"

	fileMode = 0600
)

var awsSupportFrom *semver.Version = semver.Must(semver.NewVersion("4.0.0"))
var azureSupportFrom *semver.Version = semver.Must(semver.NewVersion("1.0.0"))

func BackupPrefix(clusterID string) string {
	return "-" + clusterID
}

func CertFile(clusterID string, tmpDir string) string {
	return path.Join(tmpDir, fmt.Sprintf("%s-%s.pem", clusterID, "crt"))
}

func CAFile(clusterID string, tmpDir string) string {
	return path.Join(tmpDir, fmt.Sprintf("%s-%s.pem", clusterID, "ca"))
}

func KeyFile(clusterID string, tmpDir string) string {
	return path.Join(tmpDir, fmt.Sprintf("%s-%s.pem", clusterID, "key"))
}

func AwsEtcdEndpoint(etcdDomain string) string {
	return fmt.Sprintf("https://%s:2379", etcdDomain)
}
func AzureEtcdEndpoint(etcdDomain string) string {
	return fmt.Sprintf("https://%s:2379", etcdDomain)
}
func KVMEtcdEndpoint(etcdDomain string) string {
	return fmt.Sprintf("https://%s:443", etcdDomain)
}
