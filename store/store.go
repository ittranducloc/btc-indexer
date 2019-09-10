package store

import "github.com/darkknightbk52/btc-indexer/model"

type Manager interface {
	GetLatestBlock() (*model.Block, error)
	GetBlock(height int64) (*model.Block, error)
	Reorg(event *model.Reorg) error
	AddBlocksData([]*model.Block, []*model.Tx, []*model.TxIn, []*model.TxOut) error
}
