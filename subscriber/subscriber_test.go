package subscriber

import (
	"bytes"
	"context"
	"github.com/btcsuite/btcd/wire"
	"github.com/darkknightbk52/btc-indexer/common/log"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"sync"
	"testing"
	"time"
)

func TestSubscribe(t *testing.T) {
	RegisterTestingT(t)
	log.Init(false)

	sub := NewSubscriber(BtcUrl("tcp://10.2.3.205:28332"),
		TimeoutDuration(time.Second*10),
		RetryDuration(time.Second*5))

	ctx, _ := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)
	ch := make(chan interface{})

	go func() {
		for {
			select {
			case <-ctx.Done():
			case noti := <-ch:
				msg := noti.([][]byte)
				switch string(msg[0]) {
				case "rawblock":
					block := new(wire.MsgBlock)
					err := block.Deserialize(bytes.NewBuffer(msg[1]))
					Expect(err).Should(Succeed())
					log.L().Info("block", zap.Any("Header", block.Header), zap.Int("TxNo", len(block.Transactions)))
				case "rawtx":
					tx := new(wire.MsgTx)
					err := tx.Deserialize(bytes.NewBuffer(msg[1]))
					Expect(err).Should(Succeed())
					log.L().Info("tx", zap.Int("TxInNo", len(tx.TxIn)), zap.Int("TxOutNo", len(tx.TxOut)))
				default:
					log.S().Info("Unknown msg:", msg)
				}
			}
		}
	}()

	sub.SubscribeNotification(ctx, wg, ch)

	select {}
}
