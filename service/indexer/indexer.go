package indexer

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	bcClient "github.com/darkknightbk52/btc-indexer/client/blockchain"
	"github.com/darkknightbk52/btc-indexer/common"
	"github.com/darkknightbk52/btc-indexer/common/log"
	"github.com/darkknightbk52/btc-indexer/model"
	"github.com/darkknightbk52/btc-indexer/store"
	"github.com/darkknightbk52/btc-indexer/subscriber"
	"go.uber.org/zap"
	"sync"
)

type Indexer struct {
	config       Config
	netParams    chaincfg.Params
	currentBlock *model.Block
	subscriber   subscriber.Subscriber
	manager      store.Manager
	client       bcClient.Client
}

const (
	blockBatchSize = 50
)

func NewIndexer(config Config, subscriber subscriber.Subscriber, manager store.Manager, client bcClient.Client) *Indexer {
	return &Indexer{
		config:     config,
		subscriber: subscriber,
		manager:    manager,
		client:     client,
	}
}

func (idx *Indexer) Listen(ctx context.Context, fromBlockHeight int64) error {
	err := idx.initState(fromBlockHeight)
	if err != nil {
		return fmt.Errorf("failed to Init State: %v", err)
	}

	var wg sync.WaitGroup
	listenCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	notiCh := make(chan interface{}, 100)
	err = idx.subscriber.SubscribeNotification(listenCtx, &wg, notiCh)
	if err != nil {
		return fmt.Errorf("failed to Subscribe Notification: %v", err)
	}

	for {
		select {
		case <-listenCtx.Done():
			wg.Wait()
			return listenCtx.Err()
		case noti := <-notiCh:
			msg, ok := noti.([][]byte)
			if !ok {
				log.L().Warn("Unexpected Notification from Subscriber", zap.Any("Noti", noti))
				continue
			}

			err := idx.sync(msg)
			if err != nil {
				log.L().Warn("Failed to Sync", zap.Error(err))
			}
		}
	}
}

func (idx *Indexer) sync(msg [][]byte) error {
	msgType := string(msg[0])
	switch msgType {
	case "rawblock":
		rawBlock := new(wire.MsgBlock)
		err := rawBlock.Deserialize(bytes.NewBuffer(msg[1]))
		if err != nil {
			return fmt.Errorf("failed to Deserialize Raw Block: %v", err)
		}
		sequence := binary.LittleEndian.Uint32(msg[2])
		return idx.syncBlock(rawBlock, sequence)
	case "rawtx":
		rawTx := new(wire.MsgTx)
		err := rawTx.Deserialize(bytes.NewBuffer(msg[1]))
		if err != nil {
			return fmt.Errorf("failed to Deserialize Raw Tx: %v", err)
		}
		sequence := binary.LittleEndian.Uint32(msg[2])
		return idx.syncTx(rawTx, sequence)
	default:
		// It's possible that the message wasn't fully read if
		// Full Node shuts down, which will produce an unreadable
		// event type. To prevent from logging it, we'll make
		// sure it conforms to the ASCII standard.
		if len(msgType) == 0 || common.IsASCII(msgType) {
			return nil
		}

		return fmt.Errorf("unexpected Message Type from Subscription: %s", msgType)
	}
}

func (idx *Indexer) syncBlock(rawBlock *wire.MsgBlock, sequence uint32) error {
	targetBlockHeader, err := idx.client.GetBlockHeaderVerboseByHash(rawBlock.BlockHash().String())
	if err != nil {
		return fmt.Errorf("failed to Get Block Header Verbose By Hash '%s': %v", rawBlock.BlockHash().String(), err)
	}

	targetBlockHeight := int64(targetBlockHeader.Height)
	for idx.currentBlock.Height < targetBlockHeight {
		nextBlockHeader := targetBlockHeader
		if targetBlockHeight-idx.currentBlock.Height > blockBatchSize {
			nextBlockHeight := idx.currentBlock.Height + blockBatchSize
			header, err := idx.client.GetBlockHeaderVerboseByHeight(nextBlockHeight)
			if err != nil {
				return fmt.Errorf("failed to Get Block Header Verbose By Height '%d': %v", targetBlockHeight, err)
			}
			nextBlockHeader = header
		}
		header, err := idx.syncBlockMaybeReorg(nextBlockHeader)
		if err != nil {
			return fmt.Errorf("failed to Sync Block Maybe Reorg, to Height '%d': %v", nextBlockHeader.Height, err)
		}
		if header != nil {
			idx.currentBlock = common.ToBlock(header)
		}
	}
	return nil
}

func (idx *Indexer) syncBlockMaybeReorg(header *btcjson.GetBlockHeaderVerboseResult) (*btcjson.GetBlockHeaderVerboseResult, error) {
	if idx.currentBlock.Height >= int64(header.Height) {
		// Ignore old block
		return nil, nil
	}

	if idx.currentBlock.Height == int64(header.Height)-1 && idx.currentBlock.Hash == header.PreviousHash {
		highestBlockHeader, err := idx.addBlocks([]*btcjson.GetBlockHeaderVerboseResult{header})
		if err != nil {
			return nil, fmt.Errorf("failed to Add A New Block: %v", err)
		}
		return highestBlockHeader, nil
	}

	reorg := &model.Reorg{
		FromHeight: idx.currentBlock.Height,
		FromHash:   idx.currentBlock.Hash,
		ToHeight:   idx.currentBlock.Height,
		ToHash:     idx.currentBlock.Hash,
	}
	headers := []*btcjson.GetBlockHeaderVerboseResult{header}
	var err error
	for {
		if idx.currentBlock.Height == int64(header.Height)-1 && idx.currentBlock.Hash == header.PreviousHash {
			reorg = nil
			break
		}

		if idx.currentBlock.Height > int64(header.Height)-1 {
			block, err := idx.manager.GetBlock(int64(header.Height) - 1)
			if err != nil && err == common.ErrNotFound {
				log.L().Warn("reorg examining - not found block in DB", zap.Int32("blockHeight", header.Height-1))
				break
			}
			if err != nil {
				return nil, fmt.Errorf("reorg examining - failed to Get Block, height '%d': %v", header.Height-1, err)
			}
			if block.Hash == header.PreviousHash {
				break
			}
			reorg.FromHeight = block.Height
			reorg.FromHash = block.Hash
		}
		previousHash := header.PreviousHash
		header, err = idx.client.GetBlockHeaderVerboseByHash(previousHash)
		if err != nil {
			return nil, fmt.Errorf("reorg examining - failed to Get Block eader Verbose By Hash '%s': %v", previousHash, err)
		}
		headers = append(headers, header)
	}

	if reorg == nil {
		highestBlockHeader, err := idx.addBlocks(headers)
		if err != nil {
			return nil, fmt.Errorf("failed to Add New Blocks, BlockNo '%d': %v", len(headers), err)
		}
		return highestBlockHeader, nil
	}

	log.L().Info("Reorg happened", zap.Any("event", reorg))

	err = idx.manager.Reorg(reorg)
	if err != nil {
		return nil, fmt.Errorf("failed to Reorg: %v", err)
	}
	reorgBlock := header
	bh, err := idx.client.GetBlockHeaderVerboseByHash(reorgBlock.PreviousHash)
	if err != nil {
		return nil, fmt.Errorf("failed to Get Block Header Verbose By Hash '%s': %v", reorgBlock.PreviousHash, err)
	}
	return bh, nil
}

func (idx *Indexer) addBlocks(headers []*btcjson.GetBlockHeaderVerboseResult) (*btcjson.GetBlockHeaderVerboseResult, error) {
	blocks, txs, txIns, txOuts, err := idx.buildBlocksData(headers)
	if err != nil {
		return nil, fmt.Errorf("failed to Build Blocks Data: %v", err)
	}
	err = idx.manager.AddBlocksData(blocks, txs, txIns, txOuts)
	if err != nil {
		return nil, fmt.Errorf("failed to Add Blocks Data: %v", err)
	}
	return headers[0], nil
}

func (idx *Indexer) buildBlocksData(headers []*btcjson.GetBlockHeaderVerboseResult) ([]*model.Block, []*model.Tx, []*model.TxIn, []*model.TxOut, error) {
	rawBlocks := make([]*wire.MsgBlock, 0, len(headers))
	blocks := make([]*model.Block, 0, len(headers))
	blockHashWithHeight := make(map[string]int64, len(headers))
	for _, h := range headers {
		b, err := idx.client.GetRawBlock(h.Hash)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to Get Raw Block, Hash '%s': %v", h.Hash, err)
		}
		rawBlocks = append(rawBlocks, b)
		blocks = append(blocks, &model.Block{
			Height:       int64(h.Height),
			Hash:         h.Hash,
			PreviousHash: h.PreviousHash,
		})
		blockHashWithHeight[h.Hash] = int64(h.Height)
	}

	var txNo, txInNo, txOutNo int
	for _, b := range rawBlocks {
		txNo += len(b.Transactions)
		for _, tx := range b.Transactions {
			txInNo += len(tx.TxIn)
			txOutNo += len(tx.TxOut)
		}
	}

	txs := make([]*model.Tx, 0, txNo)
	txIns := make([]*model.TxIn, 0, txInNo)
	txOuts := make([]*model.TxOut, 0, txOutNo)
	for _, b := range rawBlocks {
		for _, tx := range b.Transactions {
			isCoinBase := blockchain.IsCoinBaseTx(tx)
			height := blockHashWithHeight[b.BlockHash().String()]
			txs = append(txs, &model.Tx{
				Height:   height,
				Hash:     tx.TxHash().String(),
				CoinBase: &isCoinBase,
			})
			ins, outs := idx.buildTxData(height, tx, isCoinBase)
			txIns = append(txIns, ins...)
			txOuts = append(txOuts, outs...)
		}
	}

	return blocks, txs, txIns, txOuts, nil
}

func (idx *Indexer) buildTxData(height int64, tx *wire.MsgTx, isCoinBase bool) ([]*model.TxIn, []*model.TxOut) {
	txIns := make([]*model.TxIn, 0, len(tx.TxIn))
	txOuts := make([]*model.TxOut, 0, len(tx.TxOut))
	chainParams := idx.config.ChainParams()

	for i, in := range tx.TxIn {
		addr, err := common.GetAddrFromTxIn(in, &chainParams)
		if err != nil {
			log.L().Warn("failed to Get Address From Tx In", zap.String("TxHash", tx.TxHash().String()), zap.Int("TxInIndex", i), zap.Error(err))
		}

		txIns = append(txIns, &model.TxIn{
			TxHash:          tx.TxHash().String(),
			TxIndex:         int32(i),
			Height:          height,
			Address:         addr,
			PreviousTxHash:  in.PreviousOutPoint.Hash.String(),
			PreviousTxIndex: int32(in.PreviousOutPoint.Index),
		})
	}

	for i, out := range tx.TxOut {
		addr, err := common.GetAddrFromTxOut(out, &chainParams)
		if err != nil {
			log.L().Warn("failed to Get Address From Tx Out", zap.String("TxHash", tx.TxHash().String()), zap.Int("TxOutIndex", i), zap.Error(err))
			if !idx.config.IncludeNonStandard {
				log.L().Warn("Ignore Non Standard Tx Out")
				continue
			}
		}
		txOuts = append(txOuts, &model.TxOut{
			Height:       height,
			TxHash:       tx.TxHash().String(),
			TxIndex:      int32(i),
			Value:        out.Value,
			Address:      addr,
			ScriptPubKey: out.PkScript,
			CoinBase:     &isCoinBase,
		})
	}

	return txIns, txOuts
}

func (idx *Indexer) syncTx(rawTx *wire.MsgTx, sequence uint32) error {
	return nil
}

func (idx *Indexer) initState(fromBlockHeight int64) error {
	block, err := idx.manager.GetLatestBlock()
	if err != nil && err != common.ErrNotFound {
		return fmt.Errorf("failed to Get Latest Block: %v", err)
	}

	if block == nil {
		header, err := idx.client.GetBlockHeaderVerboseByHeight(fromBlockHeight)
		if err != nil {
			return fmt.Errorf("failed to Get Block Header Verbose By Height '%d': %v", fromBlockHeight, err)
		}
		result, err := idx.addBlocks([]*btcjson.GetBlockHeaderVerboseResult{header})
		if err != nil {
			return fmt.Errorf("failed to Add Initial Block: %v", err)
		}
		idx.currentBlock = common.ToBlock(result)
		return nil
	}

	if block.Height < fromBlockHeight {
		return fmt.Errorf("invalid starting Block Height: Latest Block '%d', From Block '%d'", block.Height, fromBlockHeight)
	}

	idx.currentBlock = block
	return nil
}
