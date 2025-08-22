package migrate

import "os"

// Backup placeholder: create /backups/<timestamp>/ before mutation.
func Backup(dir string) error {
	return os.MkdirAll(dir, 0o755)
}
