package backup

type BackupConfig interface {
	Create() error
	Encrypt() error
	Upload() error
	Version() string
}
