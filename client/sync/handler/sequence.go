package handler

import (
	"context"
	"fmt"
	btc_indexer "github.com/darkknightbk52/btc-indexer"
	"github.com/darkknightbk52/btc-indexer/common"
	"github.com/darkknightbk52/btc-indexer/common/log"
	"github.com/darkknightbk52/btc-indexer/model"
	proto "github.com/darkknightbk52/btc-indexer/proto"
	"github.com/darkknightbk52/btc-indexer/store"
	"go.uber.org/zap"
	"time"
)

type sequenceHandler struct {
	ctx                           context.Context
	stream                        proto.BtcIndexer_SyncStream
	manager                       store.Manager
	addressWatcher                btc_indexer.AddressWatcher
	recentBlocksAscendingByHeight []*proto.Block
	getBlockIntervalInSec         int
}

func NewSequenceHandler(ctx context.Context, stream proto.BtcIndexer_SyncStream, manager store.Manager, addressBook btc_indexer.AddressWatcher, recentBlocks []*proto.Block, getBlockIntervalInSec int) *sequenceHandler {
	return &sequenceHandler{ctx: ctx, stream: stream, manager: manager, addressWatcher: addressBook, recentBlocksAscendingByHeight: recentBlocks, getBlockIntervalInSec: getBlockIntervalInSec}
}

func (h *sequenceHandler) Handle() error {
	branchBlock, newBlock, err := h.checkReorg()
	if err != nil {
		return fmt.Errorf("failed to Check Reorg: %v", err)
	}

	if branchBlock != nil || newBlock != nil {
		err := h.stream.Send(&proto.SyncResponse{
			Response: &proto.SyncResponse_ReorgBlock_{ReorgBlock: &proto.SyncResponse_ReorgBlock{
				Height:  branchBlock.Height,
				OldHash: branchBlock.Hash,
				NewHash: newBlock.Hash,
			}},
		})
		if err != nil {
			return fmt.Errorf("failed to Send 'Reorg' msg from Streamer: %v", err)
		}
	}

	nextHeight := h.recentBlocksAscendingByHeight[len(h.recentBlocksAscendingByHeight)-1].Height + 1
	ticker := time.NewTicker(time.Second * time.Duration(h.getBlockIntervalInSec))
	defer ticker.Stop()
	for {
		nextBlock, err := h.manager.GetBlock(nextHeight)
		if err != nil && err != common.ErrNotFound {
			return fmt.Errorf("failed to Get Next Block, height '%d': %v", nextHeight, err)
		}

		if nextBlock != nil {
			err := h.process(nextBlock)
			if err != nil {
				return fmt.Errorf("failed to Process Next Block '%v': %v", nextBlock, err)
			}
			return nil
		}

		log.L().Debug("Retry to get block", zap.Int64("height", nextHeight), zap.Time("At", time.Now()))
		select {
		case <-h.ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (h *sequenceHandler) process(nextBlock *model.Block) error {
	mostRecentBlock := h.recentBlocksAscendingByHeight[len(h.recentBlocksAscendingByHeight)-1]
	if nextBlock.PreviousHash == mostRecentBlock.Hash {
		err := h.processNewBlock(nextBlock)
		if err != nil {
			return fmt.Errorf("failed to Process New Block: %v", err)
		}
		return nil
	}

	err := h.processReorgBlock()
	if err != nil {
		return fmt.Errorf("failed to Process Reorg Block: %v", err)
	}
	return nil
}

func (h *sequenceHandler) processReorgBlock() error {
	branchBlock, newBlock, err := h.checkReorg()
	if err != nil {
		return fmt.Errorf("failed to Check Reorg: %v", err)
	}

	if branchBlock == nil || newBlock == nil {
		return fmt.Errorf("failed to Get Reorg Block info")
	}

	err = h.stream.Send(&proto.SyncResponse{
		Response: &proto.SyncResponse_ReorgBlock_{ReorgBlock: &proto.SyncResponse_ReorgBlock{
			Height:  branchBlock.Height,
			OldHash: branchBlock.Hash,
			NewHash: newBlock.Hash,
		}},
	})
	if err != nil {
		return fmt.Errorf("failed to Send 'Reorg' msg from Streamer: %v", err)
	}

	return nil
}

func (h *sequenceHandler) processNewBlock(b *model.Block) error {
	_, txIns, txOuts, err := h.manager.GetBlocksData(b.Height, b.Height, h.addressWatcher.GetAddresses())
	if err != nil {
		return fmt.Errorf("failed to Get Block Data, Height '%d', No Of WatchingAddresses '%d': %v", b.Height, len(h.addressWatcher.GetAddresses()), err)
	}

	syncBlock := common.BuildProtoMsg(b.Height, b, txIns[b.Height], txOuts[b.Height])
	err = h.stream.Send(&proto.SyncResponse{
		Response: &proto.SyncResponse_SyncBlock_{
			SyncBlock: syncBlock,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to Send 'Sync Block' msg from Streamer, block '%v': %v", b, err)
	}
	return nil
}

func (h *sequenceHandler) checkReorg() (branchBlock *proto.Block, newBlock *model.Block, err error) {
	heights := make([]int64, 0, len(h.recentBlocksAscendingByHeight))
	for _, b := range h.recentBlocksAscendingByHeight {
		heights = append(heights, b.Height)
	}
	localBlocks, err := h.manager.GetBlocks(heights)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to Get Blocks, block heights '%v': %v", heights, err)
	}

	for _, recentBlock := range h.recentBlocksAscendingByHeight {
		localBlock := localBlocks[recentBlock.Height]
		if localBlock == nil {
			return nil, nil, fmt.Errorf("block missed in local, height '%d'", recentBlock.Height)
		}

		if localBlock.Hash != recentBlock.Hash {
			return recentBlock, localBlock, nil
		}
	}
	return nil, nil, nil
}
