package indexer

import (
	"context"
	"errors"
	"github.com/btcsuite/btcd/chaincfg"
	clientMock "github.com/darkknightbk52/btc-indexer/client/blockchain/mocks"
	commonIndexer "github.com/darkknightbk52/btc-indexer/common"
	"github.com/darkknightbk52/btc-indexer/common/log"
	"github.com/darkknightbk52/btc-indexer/model"
	managerMock "github.com/darkknightbk52/btc-indexer/store/mocks"
	subMock "github.com/darkknightbk52/btc-indexer/subscriber/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
	"time"
)

var _ = Describe("Indexer Test", func() {
	var (
		mockSubscriber *subMock.Subscriber
		mockClient     *clientMock.Client
		mockManager    *managerMock.Manager
		indexer        *Indexer
	)

	ch := make(chan<- interface{})
	subFunc := func(ctx context.Context, wg *sync.WaitGroup, subChan chan<- interface{}) error {
		ch = subChan
		return nil
	}
	BeforeEach(func() {
		log.Init(false)
		mockSubscriber = new(subMock.Subscriber)
		mockClient = new(clientMock.Client)
		mockManager = new(managerMock.Manager)
		indexer = &Indexer{
			config: Config{
				Network: "TestNet3",
			},
			netParams:  chaincfg.TestNet3Params,
			subscriber: mockSubscriber,
			manager:    mockManager,
			client:     mockClient,
		}
	})

	AfterEach(func() {
		mockSubscriber.AssertExpectations(GinkgoT())
		mockClient.AssertExpectations(GinkgoT())
		mockManager.AssertExpectations(GinkgoT())
	})

	Context("Functional", func() {
		Context("Listen - sync normally", func() {
			It("Empty DB, full scan from genesis block", func() {
				// Start syncing from block 0
				mockManager.On("GetLatestBlock").Return(nil, commonIndexer.ErrNotFound).Once()
				mockClient.On("GetBlockHeaderVerboseByHeight", int64(0)).Return(rawBlockHeaders[0], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[0].Hash).Return(rawBlocks[0], nil).Once()
				mockManager.On("AddBlocksData", []*model.Block{modelBlocks[0]}, modelTxs[0], modelTxIns[0], modelTxOuts[0]).Return(nil).Once()

				// Sync block 1
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[1].BlockHash().String()).Return(rawBlockHeaders[1], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[1].Hash).Return(rawBlocks[1], nil).Once()
				mockManager.On("AddBlocksData", []*model.Block{modelBlocks[1]}, modelTxs[1], modelTxIns[1], modelTxOuts[1]).Return(nil).Once()

				// Sync block 2
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[2].BlockHash().String()).Return(rawBlockHeaders[2], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[2].Hash).Return(rawBlocks[2], nil).Once()
				mockManager.On("AddBlocksData", []*model.Block{modelBlocks[2]}, modelTxs[2], modelTxIns[2], modelTxOuts[2]).Return(nil).Once()

				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(time.Second)
					ch <- rawNotifications[1]
					ch <- rawNotifications[2]
					time.Sleep(time.Second)
					cancel()
					log.L().Info("Shutdown Indexer")
				}()

				mockSubscriber.On("SubscribeNotification", mock.Anything, mock.Anything, mock.Anything).Return(subFunc).Once()

				err := indexer.Listen(ctx, 0)
				Expect(err).Should(Equal(context.Canceled))
			})

			It("Local latest block as 1, be notified with block 2", func() {
				// Start syncing from block 1
				mockManager.On("GetLatestBlock").Return(modelBlocks[1], nil).Once()

				// Sync block 2
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[2].BlockHash().String()).Return(rawBlockHeaders[2], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[2].Hash).Return(rawBlocks[2], nil).Once()
				mockManager.On("AddBlocksData", []*model.Block{modelBlocks[2]}, modelTxs[2], modelTxIns[2], modelTxOuts[2]).Return(nil).Once()

				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(time.Second)
					ch <- rawNotifications[2]
					time.Sleep(time.Second)
					cancel()
					log.L().Info("Shutdown Indexer")
				}()

				mockSubscriber.On("SubscribeNotification", mock.Anything, mock.Anything, mock.Anything).Return(subFunc).Once()

				err := indexer.Listen(ctx, 0)
				Expect(err).Should(Equal(context.Canceled))
			})

			It("Local latest block as 1, be notified with block 4 => rescan from block 4 to 2", func() {
				// Start syncing from block 1
				mockManager.On("GetLatestBlock").Return(modelBlocks[1], nil).Once()

				// Sync block 4 to 2
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[4].BlockHash().String()).Return(rawBlockHeaders[4], nil).Once()
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[3].BlockHash().String()).Return(rawBlockHeaders[3], nil).Once()
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[2].BlockHash().String()).Return(rawBlockHeaders[2], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[4].Hash).Return(rawBlocks[4], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[3].Hash).Return(rawBlocks[3], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[2].Hash).Return(rawBlocks[2], nil).Once()
				blocks := []*model.Block{
					modelBlocks[4],
					modelBlocks[3],
					modelBlocks[2],
				}
				txs := append(append(modelTxs[4], modelTxs[3]...), modelTxs[2]...)
				txIns := append(append(modelTxIns[4], modelTxIns[3]...), modelTxIns[2]...)
				txOuts := append(append(modelTxOuts[4], modelTxOuts[3]...), modelTxOuts[2]...)
				mockManager.On("AddBlocksData", blocks, txs, txIns, txOuts).Return(nil).Once()

				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(time.Second)
					ch <- rawNotifications[4]
					time.Sleep(time.Second)
					cancel()
					log.L().Info("Shutdown Indexer")
				}()

				mockSubscriber.On("SubscribeNotification", mock.Anything, mock.Anything, mock.Anything).Return(subFunc).Once()

				err := indexer.Listen(ctx, 0)
				Expect(err).Should(Equal(context.Canceled))
			})
		})

		Context("Listen - have reorg", func() {
			It("Sync to block 4, reorg at block 4", func() {
				initReorgBlock4(rawBlockHeaders[3].Hash)
				initReorgBlock5(reorgRawBlockHeaders[4].Hash)

				// Start syncing from block 4
				mockManager.On("GetLatestBlock").Return(modelBlocks[4], nil).Once()

				// Notified reorg block 4 => just ignore
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlocks[4].BlockHash().String()).Return(reorgRawBlockHeaders[4], nil).Once()

				// Notified block 5 chained with reorg block 4 => start to handle reorg
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlocks[5].BlockHash().String()).Return(reorgRawBlockHeaders[5], nil).Once()

				// Tracing blocks backward in local DB by height to calculate reorg depth, in the case, as 1 block (block 4)
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlocks[4].BlockHash().String()).Return(reorgRawBlockHeaders[4], nil).Once()
				mockManager.On("GetBlock", int64(3)).Return(modelBlocks[3], nil).Once()
				reorg := &model.Reorg{
					FromHeight: 4,
					FromHash:   rawBlocks[4].BlockHash().String(),
					ToHeight:   4,
					ToHash:     rawBlocks[4].BlockHash().String(),
				}

				// Process reorg event, expect delete old block 4, now the current block header points to block 3
				mockManager.On("Reorg", reorg).Return(nil).Once()
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlockHeaders[4].PreviousHash).Return(rawBlockHeaders[3], nil).Once()

				// Rescan from branch block as block 4, local DB be updated with reorg blocks 4 & 5
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlockHeaders[4].Hash).Return(reorgRawBlockHeaders[4], nil).Once()
				mockClient.On("GetRawBlock", reorgRawBlockHeaders[5].Hash).Return(reorgRawBlocks[5], nil).Once()
				mockClient.On("GetRawBlock", reorgRawBlockHeaders[4].Hash).Return(reorgRawBlocks[4], nil).Once()
				blocks := []*model.Block{
					reorgModelBlocks[5],
					reorgModelBlocks[4],
				}
				txs := append(reorgModelTxs[5], reorgModelTxs[4]...)
				txIns := append(reorgModelTxIns[5], reorgModelTxIns[4]...)
				txOuts := append(reorgModelTxOuts[5], reorgModelTxOuts[4]...)
				mockManager.On("AddBlocksData", blocks, txs, txIns, txOuts).Return(nil).Once()

				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(time.Second)
					ch <- reorgRawNotifications[4]
					ch <- reorgRawNotifications[5]
					time.Sleep(time.Second)
					cancel()
					log.L().Info("Shutdown Indexer")
				}()

				mockSubscriber.On("SubscribeNotification", mock.Anything, mock.Anything, mock.Anything).Return(subFunc).Once()

				err := indexer.Listen(ctx, 0)
				Expect(err).Should(Equal(context.Canceled))
			})

			It("Sync to block 4, reorg at block 3", func() {
				initReorgBlock3(rawBlockHeaders[2].Hash)
				initReorgBlock4(reorgRawBlockHeaders[3].Hash)
				initReorgBlock5(reorgRawBlockHeaders[4].Hash)

				// Start syncing from block 4
				mockManager.On("GetLatestBlock").Return(modelBlocks[4], nil).Once()

				// Notified reorg block 3,4 => just ignore
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlocks[3].BlockHash().String()).Return(reorgRawBlockHeaders[3], nil).Once()
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlocks[4].BlockHash().String()).Return(reorgRawBlockHeaders[4], nil).Once()

				// Notified block 5 chained with reorg block 4, 3 => start to handle reorg
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlocks[5].BlockHash().String()).Return(reorgRawBlockHeaders[5], nil).Once()

				// Tracing blocks backward in local DB by height to calculate reorg depth, in the case, as 2 blocks (block 4 & 3)
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlocks[4].BlockHash().String()).Return(reorgRawBlockHeaders[4], nil).Once()
				mockManager.On("GetBlock", int64(3)).Return(modelBlocks[3], nil).Once()
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlocks[3].BlockHash().String()).Return(reorgRawBlockHeaders[3], nil).Once()
				mockManager.On("GetBlock", int64(2)).Return(modelBlocks[2], nil).Once()
				reorg := &model.Reorg{
					FromHeight: 3,
					FromHash:   rawBlocks[3].BlockHash().String(),
					ToHeight:   4,
					ToHash:     rawBlocks[4].BlockHash().String(),
				}

				// Process reorg event, expect delete old blocks 4 & 3, now the current block header points to block 2
				mockManager.On("Reorg", reorg).Return(nil).Once()
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlockHeaders[3].PreviousHash).Return(rawBlockHeaders[2], nil).Once()

				// Rescan from branch block as block 3, local DB be updated with reorg blocks 3, 4 & 5
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlockHeaders[4].Hash).Return(reorgRawBlockHeaders[4], nil).Once()
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlockHeaders[3].Hash).Return(reorgRawBlockHeaders[3], nil).Once()
				mockClient.On("GetRawBlock", reorgRawBlockHeaders[5].Hash).Return(reorgRawBlocks[5], nil).Once()
				mockClient.On("GetRawBlock", reorgRawBlockHeaders[4].Hash).Return(reorgRawBlocks[4], nil).Once()
				mockClient.On("GetRawBlock", reorgRawBlockHeaders[3].Hash).Return(reorgRawBlocks[3], nil).Once()
				blocks := []*model.Block{
					reorgModelBlocks[5],
					reorgModelBlocks[4],
					reorgModelBlocks[3],
				}
				txs := append(append(reorgModelTxs[5], reorgModelTxs[4]...), reorgModelTxs[3]...)
				txIns := append(append(reorgModelTxIns[5], reorgModelTxIns[4]...), reorgModelTxIns[3]...)
				txOuts := append(append(reorgModelTxOuts[5], reorgModelTxOuts[4]...), reorgModelTxOuts[3]...)
				mockManager.On("AddBlocksData", blocks, txs, txIns, txOuts).Return(nil).Once()

				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(time.Second)
					ch <- reorgRawNotifications[3]
					ch <- reorgRawNotifications[4]
					ch <- reorgRawNotifications[5]
					time.Sleep(time.Second)
					cancel()
					log.L().Info("Shutdown Indexer")
				}()

				mockSubscriber.On("SubscribeNotification", mock.Anything, mock.Anything, mock.Anything).Return(subFunc).Once()

				err := indexer.Listen(ctx, 0)
				Expect(err).Should(Equal(context.Canceled))
			})

			It("Local current block as block 1, notified with block 4, rescan to block 4, reorg at block 3", func() {
				initReorgBlock3(rawBlockHeaders[2].Hash)
				initReorgBlock4(reorgRawBlockHeaders[3].Hash)
				initReorgBlock5(reorgRawBlockHeaders[4].Hash)

				// Start syncing from block 1
				mockManager.On("GetLatestBlock").Return(modelBlocks[1], nil).Once()

				// Sync block 4 to 2
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[4].BlockHash().String()).Return(rawBlockHeaders[4], nil).Once()
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[3].BlockHash().String()).Return(rawBlockHeaders[3], nil).Once()
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[2].BlockHash().String()).Return(rawBlockHeaders[2], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[4].Hash).Return(rawBlocks[4], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[3].Hash).Return(rawBlocks[3], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[2].Hash).Return(rawBlocks[2], nil).Once()
				blocks := []*model.Block{
					modelBlocks[4],
					modelBlocks[3],
					modelBlocks[2],
				}
				txs := append(append(modelTxs[4], modelTxs[3]...), modelTxs[2]...)
				txIns := append(append(modelTxIns[4], modelTxIns[3]...), modelTxIns[2]...)
				txOuts := append(append(modelTxOuts[4], modelTxOuts[3]...), modelTxOuts[2]...)
				mockManager.On("AddBlocksData", blocks, txs, txIns, txOuts).Return(nil).Once()

				// Notified reorg block 3,4 => just ignore
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlocks[3].BlockHash().String()).Return(reorgRawBlockHeaders[3], nil).Once()
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlocks[4].BlockHash().String()).Return(reorgRawBlockHeaders[4], nil).Once()

				// Notified block 5 chained with reorg block 4, 3 => start to handle reorg
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlocks[5].BlockHash().String()).Return(reorgRawBlockHeaders[5], nil).Once()

				// Tracing blocks backward in local DB by height to calculate reorg depth, in the case, as 2 blocks (block 4 & 3)
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlocks[4].BlockHash().String()).Return(reorgRawBlockHeaders[4], nil).Once()
				mockManager.On("GetBlock", int64(3)).Return(modelBlocks[3], nil).Once()
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlocks[3].BlockHash().String()).Return(reorgRawBlockHeaders[3], nil).Once()
				mockManager.On("GetBlock", int64(2)).Return(modelBlocks[2], nil).Once()
				reorg := &model.Reorg{
					FromHeight: 3,
					FromHash:   rawBlocks[3].BlockHash().String(),
					ToHeight:   4,
					ToHash:     rawBlocks[4].BlockHash().String(),
				}

				// Process reorg event, expect delete old blocks 4 & 3, now the current block header points to block 2
				mockManager.On("Reorg", reorg).Return(nil).Once()
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlockHeaders[3].PreviousHash).Return(rawBlockHeaders[2], nil).Once()

				// Rescan from branch block as block 3, local DB be updated with reorg blocks 3, 4 & 5
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlockHeaders[4].Hash).Return(reorgRawBlockHeaders[4], nil).Once()
				mockClient.On("GetBlockHeaderVerboseByHash", reorgRawBlockHeaders[3].Hash).Return(reorgRawBlockHeaders[3], nil).Once()
				mockClient.On("GetRawBlock", reorgRawBlockHeaders[5].Hash).Return(reorgRawBlocks[5], nil).Once()
				mockClient.On("GetRawBlock", reorgRawBlockHeaders[4].Hash).Return(reorgRawBlocks[4], nil).Once()
				mockClient.On("GetRawBlock", reorgRawBlockHeaders[3].Hash).Return(reorgRawBlocks[3], nil).Once()
				blocks = []*model.Block{
					reorgModelBlocks[5],
					reorgModelBlocks[4],
					reorgModelBlocks[3],
				}
				txs = append(append(reorgModelTxs[5], reorgModelTxs[4]...), reorgModelTxs[3]...)
				txIns = append(append(reorgModelTxIns[5], reorgModelTxIns[4]...), reorgModelTxIns[3]...)
				txOuts = append(append(reorgModelTxOuts[5], reorgModelTxOuts[4]...), reorgModelTxOuts[3]...)
				mockManager.On("AddBlocksData", blocks, txs, txIns, txOuts).Return(nil).Once()

				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(time.Second)
					ch <- rawNotifications[4]
					ch <- reorgRawNotifications[3]
					ch <- reorgRawNotifications[4]
					ch <- reorgRawNotifications[5]
					time.Sleep(time.Second)
					cancel()
					log.L().Info("Shutdown Indexer")
				}()

				mockSubscriber.On("SubscribeNotification", mock.Anything, mock.Anything, mock.Anything).Return(subFunc).Once()

				err := indexer.Listen(ctx, 0)
				Expect(err).Should(Equal(context.Canceled))
			})
		})
	})

	Context("Occur errors", func() {
		Context("InitState failed", func() {
			It("GetLatestBlock failed", func() {
				mockManager.On("GetLatestBlock").Return(nil, errors.New("failed")).Once()

				e := indexer.Listen(context.Background(), 13)
				Expect(e).Should(Equal(errors.New("failed to Init State: failed to Get Latest Block: failed")))
			})

			It("GetBlockHeaderVerboseByHeight failed", func() {
				mockManager.On("GetLatestBlock").Return(nil, commonIndexer.ErrNotFound).Once()
				mockClient.On("GetBlockHeaderVerboseByHeight", int64(13)).Return(nil, errors.New("failed")).Once()

				e := indexer.Listen(context.Background(), 13)
				Expect(e).Should(Equal(errors.New("failed to Init State: failed to Get Block Header Verbose By Height '13': failed")))
			})

			It("Invalid starting Block Height", func() {
				mockManager.On("GetLatestBlock").Return(modelBlocks[1], nil).Once()

				e := indexer.Listen(context.Background(), 13)
				Expect(e).Should(Equal(errors.New("failed to Init State: invalid starting Block Height: Latest Block '1', From Block '13'")))
			})
		})

		It("SubscribeNotification failed", func() {
			mockManager.On("GetLatestBlock").Return(modelBlocks[3], nil).Once()
			mockSubscriber.On("SubscribeNotification", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed")).Once()

			e := indexer.Listen(context.Background(), 2)
			Expect(e).Should(Equal(errors.New("failed to Subscribe Notification: failed")))
		})

		Context("Occur errors while syncing, log & retry", func() {
			It("Invalid notification", func() {
				// Start syncing from block 1
				mockManager.On("GetLatestBlock").Return(modelBlocks[1], nil).Once()

				// Sync block 2
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[2].BlockHash().String()).Return(rawBlockHeaders[2], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[2].Hash).Return(rawBlocks[2], nil).Once()
				mockManager.On("AddBlocksData", []*model.Block{modelBlocks[2]}, modelTxs[2], modelTxIns[2], modelTxOuts[2]).Return(nil).Once()

				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(time.Second)
					ch <- "invalid notification"
					ch <- rawNotifications[2]
					time.Sleep(time.Second)
					cancel()
					log.L().Info("Shutdown Indexer")
				}()

				mockSubscriber.On("SubscribeNotification", mock.Anything, mock.Anything, mock.Anything).Return(subFunc).Once()

				err := indexer.Listen(ctx, 0)
				Expect(err).Should(Equal(context.Canceled))
			})

			It("GetBlockHeaderVerboseByHash failed", func() {
				// Start syncing from block 1
				mockManager.On("GetLatestBlock").Return(modelBlocks[1], nil).Once()

				// Occur error
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[2].BlockHash().String()).Return(nil, errors.New("failed")).Once()

				// Sync block 2
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[2].BlockHash().String()).Return(rawBlockHeaders[2], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[2].Hash).Return(rawBlocks[2], nil).Once()
				mockManager.On("AddBlocksData", []*model.Block{modelBlocks[2]}, modelTxs[2], modelTxIns[2], modelTxOuts[2]).Return(nil).Once()

				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(time.Second)
					ch <- rawNotifications[2]
					ch <- rawNotifications[2]
					time.Sleep(time.Second)
					cancel()
					log.L().Info("Shutdown Indexer")
				}()

				mockSubscriber.On("SubscribeNotification", mock.Anything, mock.Anything, mock.Anything).Return(subFunc).Once()

				err := indexer.Listen(ctx, 0)
				Expect(err).Should(Equal(context.Canceled))
			})

			It("GetRawBlock failed", func() {
				// Start syncing from block 1
				mockManager.On("GetLatestBlock").Return(modelBlocks[1], nil).Once()

				// Occur error
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[2].BlockHash().String()).Return(rawBlockHeaders[2], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[2].Hash).Return(nil, errors.New("failed")).Once()

				// Sync block 2
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[2].BlockHash().String()).Return(rawBlockHeaders[2], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[2].Hash).Return(rawBlocks[2], nil).Once()
				mockManager.On("AddBlocksData", []*model.Block{modelBlocks[2]}, modelTxs[2], modelTxIns[2], modelTxOuts[2]).Return(nil).Once()

				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(time.Second)
					ch <- rawNotifications[2]
					ch <- rawNotifications[2]
					time.Sleep(time.Second)
					cancel()
					log.L().Info("Shutdown Indexer")
				}()

				mockSubscriber.On("SubscribeNotification", mock.Anything, mock.Anything, mock.Anything).Return(subFunc).Once()

				err := indexer.Listen(ctx, 0)
				Expect(err).Should(Equal(context.Canceled))
			})

			It("AddBlocksData failed", func() {
				// Start syncing from block 1
				mockManager.On("GetLatestBlock").Return(modelBlocks[1], nil).Once()

				// Occur error
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[2].BlockHash().String()).Return(rawBlockHeaders[2], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[2].Hash).Return(rawBlocks[2], nil).Once()
				mockManager.On("AddBlocksData", []*model.Block{modelBlocks[2]}, modelTxs[2], modelTxIns[2], modelTxOuts[2]).Return(errors.New("failed")).Once()

				// Sync block 2
				mockClient.On("GetBlockHeaderVerboseByHash", rawBlocks[2].BlockHash().String()).Return(rawBlockHeaders[2], nil).Once()
				mockClient.On("GetRawBlock", rawBlockHeaders[2].Hash).Return(rawBlocks[2], nil).Once()
				mockManager.On("AddBlocksData", []*model.Block{modelBlocks[2]}, modelTxs[2], modelTxIns[2], modelTxOuts[2]).Return(nil).Once()

				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(time.Second)
					ch <- rawNotifications[2]
					ch <- rawNotifications[2]
					time.Sleep(time.Second)
					cancel()
					log.L().Info("Shutdown Indexer")
				}()

				mockSubscriber.On("SubscribeNotification", mock.Anything, mock.Anything, mock.Anything).Return(subFunc).Once()

				err := indexer.Listen(ctx, 0)
				Expect(err).Should(Equal(context.Canceled))
			})
		})
	})
},
)

func TestIndexer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Indexer Test")
}
