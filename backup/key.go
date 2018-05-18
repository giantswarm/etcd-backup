package backup

import (
	"fmt"
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
)

func BackupPrefix(clusterID string) string {
	return "k8s-" + clusterID
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
