package sync

import (
	"github.com/darkknightbk52/btc-indexer/model"
	"github.com/darkknightbk52/btc-indexer/service/indexer"
)

var (
	modelBlocks      = make(map[int64]*model.Block)
	modelTxIns       = make(map[int64][]*model.TxIn)
	modelTxOuts      = make(map[int64][]*model.TxOut)
	reorgModelBlocks = make(map[int64]*model.Block)
	reorgModelTxIns  = make(map[int64][]*model.TxIn)
	reorgModelTxOuts = make(map[int64][]*model.TxOut)
	watchedAddresses []string
)

func init() {
	indexer.InitTestData()
	modelBlocks = indexer.ModelBlocks()
	modelTxIns = indexer.ModelTxIns()
	modelTxOuts = indexer.ModelTxOuts()
	reorgModelBlocks = indexer.ReorgModelBlocks()
	reorgModelTxIns = indexer.ReorgModelTxIns()
	reorgModelTxOuts = indexer.ReorgModelTxOuts()
	watchedAddresses = indexer.Addresses()
}
