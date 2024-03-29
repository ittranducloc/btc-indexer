package blockchain

import (
	"context"
	"encoding/hex"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/darkknightbk52/btc-indexer/common/log"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"os"
	"sync"
	"testing"
)

var (
	client Client
	cfg    Config
)

func TestMain(m *testing.M) {
	log.Init(false)

	testCfg := struct {
		Host string
		User string
		Pass string
	}{}
	file, err := os.Open("./test_config.yml")
	if err != nil {
		file, err = os.Open("./default_test_config.yml")
		if err != nil {
			log.S().Fatal(err)
		}
		log.S().Info("Use Default Test Config")
	}
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&testCfg)
	if err != nil {
		log.S().Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	cfg = Config{
		Host: testCfg.Host,
		User: testCfg.User,
		Pass: testCfg.Pass,
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
		Host: "notExisted:18332",
		User: "user",
		Pass: "pass",
	}, chaincfg.MainNetParams)
	Expect(err).ShouldNot(Succeed())
	log.S().Info(err)

	_, err = NewBlockchainClient(context.Background(), &sync.WaitGroup{}, Config{
		Host: cfg.Host,
		User: "userNotExisted",
		Pass: "pass",
	}, chaincfg.MainNetParams)
	Expect(err).ShouldNot(Succeed())
	log.S().Info(err)

	isFailed := false
	if _, err = NewBlockchainClient(context.Background(), &sync.WaitGroup{}, Config{
		Host: cfg.Host,
		User: cfg.User,
		Pass: cfg.Pass,
	}, chaincfg.MainNetParams); err == nil {
		_, err = NewBlockchainClient(context.Background(), &sync.WaitGroup{}, Config{
			Host: cfg.Host,
			User: cfg.User,
			Pass: cfg.Pass,
		}, chaincfg.TestNet3Params)
		if err != nil {
			isFailed = true
		}
	} else {
		isFailed = true
	}
	Expect(isFailed).Should(BeTrue())
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

func Test(t *testing.T) {
	RegisterTestingT(t)
	block, err := client.GetRawBlock("0000000000018278632a43fa935115fd032da5eb190e6a6766fcd859c6c32495")
	Expect(err).Should(Succeed())
	for _, tx := range block.Transactions {
		log.L().Info("Tx", zap.String("Hash", tx.TxHash().String()))
		for i, txOut := range tx.TxOut {
			log.L().Info("TxOut", zap.Int("Index", i), zap.String("PkScript", hex.EncodeToString(txOut.PkScript)))
		}
	}
}
