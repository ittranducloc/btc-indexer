package indexer

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/darkknightbk52/btc-indexer/common"
)

type Config struct {
	IndexerNetwork            string
	IndexerIncludeNonStandard bool
	IndexerFullNodeUrl        string
	IndexerTimeoutInSecond    int
	IndexerRetryTimeInSecond  int
}

func (c Config) Validate() error {
	var errContent string

	if len(c.IndexerNetwork) > 0 {
		_, err := common.GetChainParams(c.IndexerNetwork)
		if err != nil {
			errContent = fmt.Sprintf("%s; %s", errContent, err.Error())
		}
	} else {
		errContent = fmt.Sprintf("%s; %s", errContent, "blockchain network required")
	}

	if len(c.IndexerFullNodeUrl) == 0 {
		errContent = fmt.Sprintf("%s; %s", errContent, "Full Node URL required")
	}

	if len(errContent) > 0 {
		return errors.New(errContent)
	}

	return nil
}

func (c Config) ChainParams() chaincfg.Params {
	params, _ := common.GetChainParams(c.IndexerNetwork)
	return *params
}
