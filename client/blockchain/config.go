package blockchain

import (
	"errors"
	"fmt"
)

type Config struct {
	BlockchainClientHost string
	BlockchainClientUser string
	BlockchainClientPass string
}

func (c Config) Validate() error {
	var errContent string
	if len(c.BlockchainClientHost) == 0 {
		errContent = "Host config for Blockchain Client is required"
	}

	if len(c.BlockchainClientUser) == 0 {
		errContent = fmt.Sprintf("%s, %s", errContent, "User config for Blockchain Client is required")
	}

	if len(c.BlockchainClientPass) == 0 {
		errContent = fmt.Sprintf("%s, %s", errContent, "Pass config for Blockchain Client is required")
	}

	if len(errContent) > 0 {
		return errors.New(errContent)
	}

	return nil
}
