package store

import (
	"errors"
	"fmt"
	"strings"
)

type Config struct {
	User     string // Username
	Password string // Password
	Host     string // Network address
	Port     int
	DBName   string // Database name
}

func (c *Config) Validate() error {
	var errContents []string
	if len(c.User) == 0 {
		errContents = append(errContents, "DB User required")
	}
	if len(c.Password) == 0 {
		errContents = append(errContents, "DB Password required")
	}
	if len(c.Host) == 0 {
		errContents = append(errContents, "DB Host required")
	}
	if c.Port == 0 {
		errContents = append(errContents, "DB Port required")
	}
	if len(c.DBName) == 0 {
		errContents = append(errContents, "DB Name required")
	}
	if len(errContents) > 0 {
		return errors.New(strings.Join(errContents, ", "))
	}
	return nil
}

// DSN returns a connection string compatible with POSTGRES server
// A DSN in its fullest form: host=localhost port=5432 user=postgres dbname=postgres password=123@123a sslmode=disable
// See: https://github.com/go-sql-driver/mysql#dsn-data-source-name
func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", c.Host, c.Port, c.User, c.DBName, c.Password)
}
