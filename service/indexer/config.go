package indexer

import (
	"errors"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/darkknightbk52/btc-indexer/common"
)

type Config struct {
	Network            string
	IncludeNonStandard bool
	FromBlockHeight    int64
}

func (c Config) Validate() error {
	if len(c.Network) == 0 {
		return errors.New("the Blockchain Network for Indexer required")
	}
	_, err := common.GetChainParams(c.Network)
	return err
}

func (c Config) ChainParams() chaincfg.Params {
	params, _ := common.GetChainParams(c.Network)
	return *params
}
