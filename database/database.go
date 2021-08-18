package database

// Backupper is the interface that wraps the basic Backup method.
type Backupper interface {
	Backup(string) error
}

// Restorer is the interface that wraps the basic Restore method.
type Restorer interface {
	Restore(string) error
}

// BackupRestorer is the interface that groups the basic Backup and Restore methods.
type BackupRestorer interface {
	Backupper
	Restorer
}
