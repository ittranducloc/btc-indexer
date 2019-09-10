package indexer

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/darkknightbk52/btc-indexer/common"
)

type Config struct {
	Network            string
	IncludeNonStandard bool
	Url                string
	TimeoutInSecond    int
	RetryTimeInSecond  int
}

func (c Config) Validate() error {
	var errContent string

	if len(c.Network) > 0 {
		_, err := common.GetChainParams(c.Network)
		if err != nil {
			errContent = fmt.Sprintf("%s; %s", errContent, err.Error())
		}
	} else {
		errContent = fmt.Sprintf("%s; %s", errContent, "blockchain network required")
	}

	if len(c.Url) == 0 {
		errContent = fmt.Sprintf("%s; %s", errContent, "Full Node URL required")
	}

	if len(errContent) > 0 {
		return errors.New(errContent)
	}

	return nil
}

func (c Config) ChainParams() chaincfg.Params {
	params, _ := common.GetChainParams(c.Network)
	return *params
}
