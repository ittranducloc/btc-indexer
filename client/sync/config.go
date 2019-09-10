package sync

import (
	"errors"
	"fmt"
)

type Config struct {
	SafeDistance          int64
	GetBlockIntervalInSec int
}

func (c Config) Validate() error {
	var errContent string
	if c.SafeDistance < 100 {
		errContent = fmt.Sprintf("the Safe Distance should be greater than 100 blocks, configured value '%d'", c.SafeDistance)
	}

	if c.GetBlockIntervalInSec < 3 || c.GetBlockIntervalInSec > 10 {
		errContent = fmt.Sprintf("%s, the Get Block Interval In Second should be in [3,10], configured value '%d'", errContent, c.GetBlockIntervalInSec)
	}

	if len(errContent) > 0 {
		return errors.New(errContent)
	}

	return nil
}
