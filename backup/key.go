package backup

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
