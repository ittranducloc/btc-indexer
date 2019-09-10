package blockchain

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/wire"
)

type Client interface {
	GetBlockHeaderVerboseByHeight(height int64) (*btcjson.GetBlockHeaderVerboseResult, error)
	GetBlockHeaderVerboseByHash(hash string) (*btcjson.GetBlockHeaderVerboseResult, error)
	GetRawBlock(hash string) (*wire.MsgBlock, error)
}
