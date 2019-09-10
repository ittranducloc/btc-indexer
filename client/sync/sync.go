package sync

import (
	"context"
	"fmt"
	btc_indexer "github.com/darkknightbk52/btc-indexer"
	"github.com/darkknightbk52/btc-indexer/client/sync/handler"
	proto "github.com/darkknightbk52/btc-indexer/proto"
	"github.com/darkknightbk52/btc-indexer/store"
	"sort"
)

type Client interface {
	Sync() error
}

type syncClient struct {
	config         Config
	ctx            context.Context
	stream         proto.BtcIndexer_SyncStream
	manager        store.Manager
	addressWatcher btc_indexer.AddressWatcher
}

func NewSyncClient(config Config, ctx context.Context, stream proto.BtcIndexer_SyncStream, manager store.Manager, addressBook btc_indexer.AddressWatcher) *syncClient {
	return &syncClient{config: config, ctx: ctx, stream: stream, manager: manager, addressWatcher: addressBook}
}

func (c *syncClient) Sync() error {
	for {
		req, err := c.stream.Recv()
		if err != nil {
			return fmt.Errorf("failed to Receive from Streamer: %v", err)
		}
		h, err := c.makeHandler(req)
		if err != nil {
			return fmt.Errorf("failed to Make Handler, req '%v': %v", req, err)
		}
		err = h.Handle()
		if err != nil {
			return fmt.Errorf("failed to Handle req '%v': %v", req, err)
		}
	}
}

func (c *syncClient) makeHandler(req *proto.SyncRequest) (handler.Handler, error) {
	latestBlock, err := c.manager.GetLatestBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to Get Latest Block: %v", err)
	}

	// Sort Recent Blocks ascending by Height
	sort.SliceStable(req.RecentBlocks, func(i, j int) bool {
		return req.RecentBlocks[i].Height < req.RecentBlocks[j].Height
	})

	mostRecentBlockHeight := func() int64 {
		if len(req.RecentBlocks) == 0 {
			return 0
		}
		return req.RecentBlocks[len(req.RecentBlocks)-1].Height
	}()

	if mostRecentBlockHeight == 0 || mostRecentBlockHeight < latestBlock.Height-c.config.SafeDistance {
		fromHeight := func() int64 {
			if mostRecentBlockHeight == 0 {
				return 0
			}
			return mostRecentBlockHeight + 1
		}
		return handler.NewBatchHandler(c.stream, c.manager, c.addressWatcher, fromHeight(), latestBlock.Height-c.config.SafeDistance), nil
	}
	return handler.NewSequenceHandler(c.ctx, c.stream, c.manager, c.addressWatcher, req.RecentBlocks, c.config.GetBlockIntervalInSec), nil
}
