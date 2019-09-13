package blockchain

import (
	"errors"
	"strings"
)

type Config struct {
	Host string
	User string
	Pass string
}

func (c Config) Validate() error {
	var errContents []string
	if len(c.Host) == 0 {
		errContents = append(errContents, "Host config for Blockchain Client is required")
	}

	if len(c.User) == 0 {
		errContents = append(errContents, "User config for Blockchain Client is required")
	}

	if len(c.Pass) == 0 {
		errContents = append(errContents, "Pass config for Blockchain Client is required")
	}

	if len(errContents) > 0 {
		return errors.New(strings.Join(errContents, ", "))
	}

	return nil
}
