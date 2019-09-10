package sync

import (
	"errors"
	"fmt"
)

type Config struct {
	SyncClientSafeDistance          int64
	SyncClientGetBlockIntervalInSec int
}

func (c Config) Validate() error {
	var errContent string
	if c.SyncClientSafeDistance < 100 {
		errContent = fmt.Sprintf("the Safe Distance should be greater than 100 blocks, configured value '%d'", c.SyncClientSafeDistance)
	}

	if c.SyncClientGetBlockIntervalInSec < 3 || c.SyncClientGetBlockIntervalInSec > 10 {
		errContent = fmt.Sprintf("%s, the Get Block Interval In Second should be in [3,10], configured value '%d'", errContent, c.SyncClientGetBlockIntervalInSec)
	}

	if len(errContent) > 0 {
		return errors.New(errContent)
	}

	return nil
}
