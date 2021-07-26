package mongodb

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Config contains mongodb basic configure.
type Config struct {
	Server     string
	Port       int
	Database   string
	Collection string
	Username   string
	Password   string
	SRV        bool
}

// URI returns mongodb uri connection string
func (c *Config) URI() string {
	var prefix, auth, server string
	if c.SRV {
		prefix = "mongodb+srv"
	} else {
		prefix = "mongodb"
	}

	if c.Username != "" && c.Password != "" {
		auth = fmt.Sprintf("%s:%s@", c.Username, c.Password)
	}

	if c.SRV || c.Port == 27017 || c.Port == 0 {
		server = c.Server
	} else {
		server = fmt.Sprintf("%s:%d", c.Server, c.Port)
	}

	return fmt.Sprintf("%s://%s%s/%s", prefix, auth, server, c.Database)
}

// Open opens a mongodb database.
func (c *Config) Open() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(c.URI()))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Backup backups mongodb database to file.
func (c *Config) Backup(file string) error {
	args := []string{}
	args = append(args, fmt.Sprintf("--uri=%q", c.URI()))
	if c.Collection != "" {
		args = append(args, "-c"+c.Collection)
	}
	args = append(args, "--gzip")
	args = append(args, fmt.Sprintf("--archive=%q", file))

	command := exec.Command("mongodump", args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("failed to backup: %s\n%v", stderr.String(), err)
	}
	return nil
}

// Restore restores mongodb database collection from file.
func (c *Config) Restore(file string) error {
	args := []string{}
	args = append(args, fmt.Sprintf("--uri=%q", c.URI()))
	args = append(args, "--gzip")
	args = append(args, "--drop")
	args = append(args, fmt.Sprintf("--archive=%q", file))

	command := exec.Command("mongorestore", args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("failed to restore: %s\n%v", stderr.String(), err)
	}
	return nil
}
