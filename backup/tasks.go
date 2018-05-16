package backup

import "log"

func FullBackup(b BackupConfig) error {
	var err error

	version := b.Version()

	err = b.Create()
	if err != nil {
		log.Printf("Etcd "+version+" backup creation failed: ", err)
	}

	err = b.Encrypt()
	if err != nil {
		log.Printf("Etcd "+version+" backup encryption failed: ", err)
	}

	err = b.Upload()
	if err != nil {
		log.Printf("Etcd "+version+" backup upload failed: ", err)
	}

	return err

}
