package subscriber

import (
	"errors"
)

type Config struct {
	FullNodeUrl       string
	TimeoutInSecond   int
	RetryTimeInSecond int
}

func (c Config) Validate() error {
	if len(c.FullNodeUrl) == 0 {
		return errors.New("the Full Node URL for Subscriber required")
	}
	return nil
}
