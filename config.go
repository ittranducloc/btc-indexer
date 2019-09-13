package btc_indexer

import (
	"errors"
	"github.com/darkknightbk52/btc-indexer/client/blockchain"
	"github.com/darkknightbk52/btc-indexer/service/indexer"
	"github.com/darkknightbk52/btc-indexer/store"
	"github.com/darkknightbk52/btc-indexer/subscriber"
	"strings"
)

type Config struct {
	Indexer              indexer.Config
	BlockchainClient     blockchain.Config
	BlockchainSubscriber subscriber.Config
	DB                   store.Config
}

func (c Config) Validate() error {
	var errContents []string
	err := c.Indexer.Validate()
	if err != nil {
		errContents = append(errContents, err.Error())
	}

	err = c.BlockchainClient.Validate()
	if err != nil {
		errContents = append(errContents, err.Error())
	}

	err = c.BlockchainSubscriber.Validate()
	if err != nil {
		errContents = append(errContents, err.Error())
	}

	err = c.DB.Validate()
	if err != nil {
		errContents = append(errContents, err.Error())
	}

	if len(errContents) > 0 {
		return errors.New(strings.Join(errContents, ", "))
	}
	return nil
}
