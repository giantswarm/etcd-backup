package backup

type Backup interface {
	Create() error
	Encrypt() error
	Upload() error
}
