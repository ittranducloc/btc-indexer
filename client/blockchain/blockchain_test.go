package blockchain

import (
	"context"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/darkknightbk52/btc-indexer/common/log"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"os"
	"sync"
	"testing"
)

var client Client

func TestMain(m *testing.M) {
	log.Init(false)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	cfg := Config{
		BlockchainClientHost: "10.2.3.205:18332",
		BlockchainClientUser: "bitcoin",
		BlockchainClientPass: "local321",
	}
	c, err := NewBlockchainClient(ctx, &wg, cfg, chaincfg.TestNet3Params)
	if err != nil {
		log.S().Fatal(err)
	}
	client = c

	out := m.Run()
	cancel()
	wg.Wait()
	os.Exit(out)
}

func TestNewBlockchainClient_Failed(t *testing.T) {
	RegisterTestingT(t)

	_, err := NewBlockchainClient(context.Background(), &sync.WaitGroup{}, Config{
		BlockchainClientHost: "localhost:18332",
		BlockchainClientUser: "user",
		BlockchainClientPass: "pass",
	}, chaincfg.MainNetParams)
	Expect(err).ShouldNot(Succeed())
	log.S().Info(err)

	_, err = NewBlockchainClient(context.Background(), &sync.WaitGroup{}, Config{
		BlockchainClientHost: "10.2.3.205:18332",
		BlockchainClientUser: "user",
		BlockchainClientPass: "pass",
	}, chaincfg.MainNetParams)
	Expect(err).ShouldNot(Succeed())
	log.S().Info(err)

	_, err = NewBlockchainClient(context.Background(), &sync.WaitGroup{}, Config{
		BlockchainClientHost: "10.2.3.205:18332",
		BlockchainClientUser: "bitcoin",
		BlockchainClientPass: "local321",
	}, chaincfg.MainNetParams)
	Expect(err).ShouldNot(Succeed())
	log.S().Info(err)
}

func TestBlockchainClient_GetBlockHeaderVerboseByHeight(t *testing.T) {
	RegisterTestingT(t)
	header, err := client.GetBlockHeaderVerboseByHeight(13)
	Expect(err).Should(Succeed())
	Expect(header.Hash).Should(Equal("0000000092c69507e1628a6a91e4e69ea28fe378a1a6a636b9c3157e84c71b78"))
	Expect(header.PreviousHash).Should(Equal("000000004705938332863b772ff732d2d5ac8fe60ee824e37813569bda3a1f00"))
	log.L().Info("GetBlockHeaderVerboseByHeight", zap.Any("header", header))
}

func TestBlockchainClient_GetBlockHeaderVerboseByHash(t *testing.T) {
	RegisterTestingT(t)
	header, err := client.GetBlockHeaderVerboseByHash("0000000092c69507e1628a6a91e4e69ea28fe378a1a6a636b9c3157e84c71b78")
	Expect(err).Should(Succeed())
	Expect(header.Height).Should(Equal(int32(13)))
	Expect(header.PreviousHash).Should(Equal("000000004705938332863b772ff732d2d5ac8fe60ee824e37813569bda3a1f00"))
	log.L().Info("GetBlockHeaderVerboseByHash", zap.Any("header", header))
}

func TestBlockchainClient_GetRawBlock(t *testing.T) {
	RegisterTestingT(t)
	block, err := client.GetRawBlock("0000000092c69507e1628a6a91e4e69ea28fe378a1a6a636b9c3157e84c71b78")
	Expect(err).Should(Succeed())
	Expect(block.BlockHash().String()).Should(Equal("0000000092c69507e1628a6a91e4e69ea28fe378a1a6a636b9c3157e84c71b78"))
	log.L().Info("GetRawBlock", zap.Any("block", block))
}
