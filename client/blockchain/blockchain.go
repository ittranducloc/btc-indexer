package blockchain

import (
	"context"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/darkknightbk52/btc-indexer/common/log"
	"sync"
)

type Client interface {
	GetBlockHeaderVerboseByHeight(height int64) (*btcjson.GetBlockHeaderVerboseResult, error)
	GetBlockHeaderVerboseByHash(hash string) (*btcjson.GetBlockHeaderVerboseResult, error)
	GetRawBlock(hash string) (*wire.MsgBlock, error)
}

type blockchainClient struct {
	rpcClient *rpcclient.Client
}

func NewBlockchainClient(ctx context.Context, wg *sync.WaitGroup, cfg Config, chainParams chaincfg.Params) (Client, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:                 cfg.Host,
		User:                 cfg.User,
		Pass:                 cfg.Pass,
		DisableAutoReconnect: false,
		DisableTLS:           true,
		DisableConnectOnNew:  true,
		HTTPPostMode:         true,
	}
	rpcClient, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to Create RPC client to Full Node, config %v: %v", cfg, err)
	}

	genesisBlockHash, err := rpcClient.GetBlockHash(0)
	if err != nil {
		rpcClient.Disconnect()
		return nil, fmt.Errorf("failed to Get Genesis (height=0) block: %v", err)
	}

	if *genesisBlockHash != *chainParams.GenesisHash {
		rpcClient.Disconnect()
		return nil, fmt.Errorf("not corresponding network, expect: '%v'", chainParams.Net)
	}

	wg.Add(1)
	log.L().Info("Blockchain Client connected")
	go func() {
		select {
		case <-ctx.Done():
			defer wg.Done()
			rpcClient.Shutdown()
			rpcClient.WaitForShutdown()
			log.L().Info("Blockchain Client disconnected")
		}
	}()

	return &blockchainClient{
		rpcClient: rpcClient,
	}, nil
}

func (c *blockchainClient) GetBlockHeaderVerboseByHeight(height int64) (*btcjson.GetBlockHeaderVerboseResult, error) {
	h, err := c.rpcClient.GetBlockHash(height)
	if err != nil {
		return nil, fmt.Errorf("failed to Get Block Hash, Height '%d': %v", height, err)
	}
	header, err := c.rpcClient.GetBlockHeaderVerbose(h)
	if err != nil {
		return nil, fmt.Errorf("failed to Get Block Header Verbose by Hash '%s', %v", h.String(), err)
	}
	return header, nil
}

func (c *blockchainClient) GetBlockHeaderVerboseByHash(hash string) (*btcjson.GetBlockHeaderVerboseResult, error) {
	h, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to Create Hash from String, hash '%s': %v", hash, err)
	}
	header, err := c.rpcClient.GetBlockHeaderVerbose(h)
	if err != nil {
		return nil, fmt.Errorf("failed to Get Block Header Verbose by Hash '%s', %v", h.String(), err)
	}
	return header, nil
}

func (c *blockchainClient) GetRawBlock(hash string) (*wire.MsgBlock, error) {
	h, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to Create Hash from String, hash '%s': %v", hash, err)
	}
	block, err := c.rpcClient.GetBlock(h)
	if err != nil {
		return nil, fmt.Errorf("failed to Get Block, Hash '%s': %v", h.String(), err)
	}
	return block, nil
}
