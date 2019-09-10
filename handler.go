package btc_indexer

import (
	"context"
	proto "github.com/darkknightbk52/btc-indexer/proto"
)

type handler struct {
}

func (h *handler) Sync(ctx context.Context, stream proto.BtcIndexer_SyncStream) error {
	panic("implement me")
}
