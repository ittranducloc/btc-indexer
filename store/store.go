package store

import "github.com/darkknightbk52/btc-indexer/model"

type Manager interface {
	GetLatestBlock() (*model.Block, error)
	GetBlock(height int64) (*model.Block, error)
	GetBlocks(heights []int64) (map[int64]*model.Block, error)
	Reorg(event *model.Reorg) error
	AddBlocksData([]*model.Block, []*model.Tx, []*model.TxIn, []*model.TxOut) error
	GetBlocksData(fromHeight, toHeight int64, interestedAddresses []string) (map[int64]*model.Block, map[int64][]*model.TxIn, map[int64][]*model.TxOut, error)
}
