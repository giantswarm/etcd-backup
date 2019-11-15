package metrics

type BackupMetrics struct {
	Successful                bool
	BackupSizeMeasurement     int64
	CreationTimeMeasurement   int64
	EncryptionTimeMeasurement int64
	UploadTimeMeasurement     int64
}

type ClusterInfo struct {
	Name string
}

func NewSuccessfulBackupMetrics(backupSize int64, creationTime int64, encryptionTime int64, uploadTime int64) *BackupMetrics {
	return &BackupMetrics{
		Successful:                true,
		BackupSizeMeasurement:     backupSize,
		CreationTimeMeasurement:   creationTime,
		EncryptionTimeMeasurement: encryptionTime,
		UploadTimeMeasurement:     uploadTime,
	}
}

func NewFailureMetrics() *BackupMetrics {
	return &BackupMetrics{
		Successful:                false,
		BackupSizeMeasurement:     -1,
		CreationTimeMeasurement:   -1,
		EncryptionTimeMeasurement: -1,
		UploadTimeMeasurement:     -1,
	}
}
