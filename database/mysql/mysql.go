package mysql

import (
	"bytes"
	"database/sql"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	_ "github.com/go-sql-driver/mysql" // mysql driver
)

// Config contains mysql basic configure.
type Config struct {
	Server   string
	Port     int
	Database string
	Username string
	Password string
}

// Open opens a mysql database.
func (c *Config) Open() (*sql.DB, error) {
	return sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		c.Username, c.Password, c.Server, c.Port, c.Database))
}

// Backup backups mysql database to file.
func (c *Config) Backup(file string) error {
	var cmd, arg string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		arg = "/c"
	case "linux":
		cmd = "bash"
		arg = "-c"
	default:
		return fmt.Errorf("unsupported operating system")
	}

	var args []string
	args = append(args, "mysqldump")
	args = append(args, fmt.Sprintf("-h%s", c.Server))
	args = append(args, fmt.Sprintf("-P%d", c.Port))
	args = append(args, fmt.Sprintf("-u%s", c.Username))
	args = append(args, fmt.Sprintf("-p%s", c.Password))
	args = append(args, fmt.Sprintf("-r%s", file))
	args = append(args, "--add-drop-database")
	args = append(args, "-RB")
	args = append(args, c.Database)

	command := exec.Command(cmd, arg, strings.Join(args, " "))
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("failed to backup database: %s\n%v", stderr.String(), err)
	}
	return nil
}

// Restore restores mysql database from file.
func (c *Config) Restore(file string) error {
	var cmd, arg string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		arg = "/c"
	case "linux":
		cmd = "bash"
		arg = "-c"
	default:
		return fmt.Errorf("unsupported operating system")
	}

	var args []string
	args = append(args, "mysql")
	args = append(args, c.Database)
	args = append(args, fmt.Sprintf("-h%s", c.Server))
	args = append(args, fmt.Sprintf("-P%d", c.Port))
	args = append(args, fmt.Sprintf("-u%s", c.Username))
	args = append(args, fmt.Sprintf("-p%s", c.Password))
	args = append(args, "<")
	args = append(args, file)

	command := exec.Command(cmd, arg, strings.Join(args, " "))
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("failed to restore database: %s\n%v", stderr.String(), err)
	}
	return nil
}
