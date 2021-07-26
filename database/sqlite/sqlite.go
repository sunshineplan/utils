package sqlite

import (
	"bytes"
	"database/sql"
	"fmt"
	"os/exec"
	"runtime"

	_ "github.com/mattn/go-sqlite3" // sqlite driver
)

const backupScript = "import sys;import sqlite3;db=sqlite3.connect(sys.argv[1]);f=open(sys.argv[2],'w');f.write('\\n'.join(db.iterdump()))"
const restoreScript = "import sys;import sqlite3;db=sqlite3.connect(sys.argv[1]);f=open(sys.argv[2]);db.executescript(f.read())"

// Config contains sqlite basic configure.
type Config struct {
	Path string
}

// Open opens a sqlite database.
func (c *Config) Open() (*sql.DB, error) {
	return sql.Open("sqlite3", c.Path)
}

// Backup backups sqlite database to file.
func (c *Config) Backup(file string) error {
	var cmd string
	switch runtime.GOOS {
	case "windows":
		cmd = "python"
	case "linux":
		cmd = "python3"
		return fmt.Errorf("unsupported operating system")
	}

	var args []string
	args = append(args, "-c")
	args = append(args, backupScript)
	args = append(args, c.Path)
	args = append(args, file)

	command := exec.Command(cmd, args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("failed to backup database: %s\n%v", stderr.String(), err)
	}
	return nil
}

// Restore restores sqlite database from file.
func (c *Config) Restore(file string) error {
	var cmd string
	switch runtime.GOOS {
	case "windows":
		cmd = "python"
	case "linux":
		cmd = "python3"
	default:
		return fmt.Errorf("unsupported operating system")
	}

	var args []string
	args = append(args, "-c")
	args = append(args, restoreScript)
	args = append(args, c.Path)
	args = append(args, file)

	command := exec.Command(cmd, args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("failed to restore database: %s\n%v", stderr.String(), err)
	}
	return nil
}
