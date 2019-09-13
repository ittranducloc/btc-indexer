package store

import (
	"errors"
	"fmt"
)

type Config struct {
	User     string // Username
	Password string // Password
	Host     string // Network address
	Port     int
	DBName   string // Database name
}

func (c *Config) Validate() error {
	if c.User == "" {
		return errors.New("database user is required")
	}
	if c.Host == "" {
		return errors.New("database address is required")
	}
	if c.DBName == "" {
		return errors.New("database name is required")
	}
	return nil
}

// DSN returns a connection string compatible with POSTGRES server
// A DSN in its fullest form: host=localhost port=5432 user=postgres dbname=postgres password=123@123a sslmode=disable
// See: https://github.com/go-sql-driver/mysql#dsn-data-source-name
func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", c.Host, c.Port, c.User, c.DBName, c.Password)
}
