package handler

import (
	"fmt"
	btc_indexer "github.com/darkknightbk52/btc-indexer"
	"github.com/darkknightbk52/btc-indexer/common"
	"github.com/darkknightbk52/btc-indexer/model"
	proto "github.com/darkknightbk52/btc-indexer/proto"
	"github.com/darkknightbk52/btc-indexer/store"
)

const blockBatchSize int64 = 1000

type batchHandler struct {
	stream         proto.BtcIndexer_SyncStream
	manager        store.Manager
	addressWatcher btc_indexer.AddressWatcher
	fromHeight     int64
	toHeight       int64
}

func NewBatchHandler(stream proto.BtcIndexer_SyncStream, manager store.Manager, addressBook btc_indexer.AddressWatcher, fromHeight int64, toHeight int64) *batchHandler {
	return &batchHandler{stream: stream, manager: manager, addressWatcher: addressBook, fromHeight: fromHeight, toHeight: toHeight}
}

func (h *batchHandler) Handle() error {
	err := h.stream.Send(&proto.SyncResponse{
		Response: &proto.SyncResponse_BeginStream_{},
	})
	if err != nil {
		return fmt.Errorf("failed to Send 'Begin Stream' msg: %v", err)
	}

	for h.fromHeight <= h.toHeight {
		targetHeight := h.toHeight
		if h.fromHeight+blockBatchSize < h.fromHeight {
			targetHeight = h.fromHeight + blockBatchSize
		}
		blocks, txIns, txOuts, err := h.manager.GetBlocksData(h.fromHeight, targetHeight, h.addressWatcher.GetAddresses())
		if err != nil {
			return fmt.Errorf("failed to Get Blocks Data, fromHeight '%d', toHeight '%d', No Of WatchingAddresses '%d': %v", h.fromHeight, targetHeight, len(h.addressWatcher.GetAddresses()), err)
		}

		err = h.sendBlocksData(h.fromHeight, targetHeight, blocks, txIns, txOuts)
		if err != nil {
			return fmt.Errorf("failed to Send Blocks Data, fromHeight '%d', toHeight '%d', No Of Blocks '%d', No Of TxIns '%d', No Of TxOuts '%d': %v", h.fromHeight, targetHeight, len(blocks), len(txIns), len(txOuts), err)
		}

		h.fromHeight = h.fromHeight + blockBatchSize + 1
	}

	err = h.stream.Send(&proto.SyncResponse{
		Response: &proto.SyncResponse_EndStream_{},
	})
	if err != nil {
		return fmt.Errorf("failed to Send 'End Stream' msg: %v", err)
	}
	return nil
}

func (h *batchHandler) sendBlocksData(fromHeight, toHeight int64, blocks map[int64]*model.Block, txIns map[int64][]*model.TxIn, txOuts map[int64][]*model.TxOut) error {
	for height := fromHeight; height <= toHeight; height++ {
		block := blocks[height]
		if blocks == nil {
			return fmt.Errorf("block missed, height '%d'", height)
		}
		syncBlock := common.BuildProtoMsg(height, block, txIns[height], txOuts[height])
		err := h.stream.Send(&proto.SyncResponse{
			Response: &proto.SyncResponse_SyncBlock_{
				SyncBlock: syncBlock,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to Send 'Sync Block' msg from Streamer, block '%v': %v", block, err)
		}
	}
	return nil
}
