package sync

import (
	"context"
	"fmt"
	"github.com/darkknightbk52/btc-indexer/common"
	"github.com/darkknightbk52/btc-indexer/common/log"
	indexerMocks "github.com/darkknightbk52/btc-indexer/mocks"
	"github.com/darkknightbk52/btc-indexer/model"
	proto "github.com/darkknightbk52/btc-indexer/proto"
	protoMocks "github.com/darkknightbk52/btc-indexer/proto/mocks"
	storeMocks "github.com/darkknightbk52/btc-indexer/store/mocks"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

var _ = Describe("Sync Client test", func() {
	var (
		client             *syncClient
		cancel             context.CancelFunc
		mockAddressWatcher *indexerMocks.AddressWatcher
		mockStream         *protoMocks.BtcIndexer_SyncStream
		mockManager        *storeMocks.Manager
	)

	BeforeEach(func() {
		log.Init(false)
		mockAddressWatcher = new(indexerMocks.AddressWatcher)
		mockStream = new(protoMocks.BtcIndexer_SyncStream)
		mockManager = new(storeMocks.Manager)
		client = &syncClient{
			config: Config{
				SafeDistance:          1,
				GetBlockIntervalInSec: 1,
			},
			addressWatcher: mockAddressWatcher,
			stream:         mockStream,
			manager:        mockManager,
		}
		client.ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		mockAddressWatcher.AssertExpectations(GinkgoT())
		mockStream.AssertExpectations(GinkgoT())
		mockManager.AssertExpectations(GinkgoT())
	})

	Context("Functional", func() {
		It("Rescan", func() {
			// Client sends request with empty recent blocks
			mockStream.On("Recv").Return(&proto.SyncRequest{}, nil).Once()
			mockManager.On("GetLatestBlock").Return(modelBlocks[2], nil).Once()

			// Client should rescan, start to stream safe blocks data to client
			mockStream.On("Send", &proto.SyncResponse{
				Response: &proto.SyncResponse_BeginStream_{},
			}).Return(nil).Once()
			mockAddressWatcher.On("GetAddresses").Return(watchedAddresses).Once()
			blocks := make(map[int64]*model.Block)
			txIns := make(map[int64][]*model.TxIn)
			txOuts := make(map[int64][]*model.TxOut)
			blocks[0] = modelBlocks[0]
			txIns[0] = modelTxIns[0]
			txOuts[0] = modelTxOuts[0]
			blocks[1] = modelBlocks[1]
			txIns[1] = modelTxIns[1]
			txOuts[1] = modelTxOuts[1]
			mockManager.On("GetBlocksData", int64(0), int64(1), watchedAddresses).Return(blocks, txIns, txOuts, nil).Once()

			// Send each block at a time
			mockStream.On("Send", &proto.SyncResponse{
				Response: &proto.SyncResponse_SyncBlock_{
					SyncBlock: common.BuildProtoMsg(0, blocks[0], txIns[0], txOuts[0]),
				},
			}).Return(nil).Once()
			mockStream.On("Send", &proto.SyncResponse{
				Response: &proto.SyncResponse_SyncBlock_{
					SyncBlock: common.BuildProtoMsg(1, blocks[1], txIns[1], txOuts[1]),
				},
			}).Return(nil).Once()

			// After sent all safe blocks data, end streaming
			mockStream.On("Send", &proto.SyncResponse{
				Response: &proto.SyncResponse_EndStream_{},
			}).Return(nil).Once()

			// Client sends request with recent blocks as 0 & 1
			req := &proto.SyncRequest{}
			req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
				Height:       0,
				Hash:         modelBlocks[0].Hash,
				PreviousHash: modelBlocks[0].PreviousHash,
			})
			req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
				Height:       1,
				Hash:         modelBlocks[1].Hash,
				PreviousHash: modelBlocks[1].PreviousHash,
			})
			mockStream.On("Recv").Return(req, nil).Once()
			mockManager.On("GetLatestBlock").Return(modelBlocks[2], nil).Once()

			// Client expect to get data of block 2 that exceeds the safe distance
			// Client may encounter New or Reorg block
			// Client should sync sequentially each block at a request
			// Firstly, have to check reorg
			mockManager.On("GetBlocks", []int64{0, 1}).Return(blocks, nil).Once()
			mockManager.On("GetBlock", int64(2)).Return(modelBlocks[2], nil).Once()

			// Have no reorg, get data of block 2 from local
			mockAddressWatcher.On("GetAddresses").Return(watchedAddresses).Once()
			blocks[2] = modelBlocks[2]
			txIns[2] = modelTxIns[2]
			txOuts[2] = modelTxOuts[2]
			mockManager.On("GetBlocksData", int64(2), int64(2), watchedAddresses).Return(blocks, txIns, txOuts, nil).Once()

			// Send block 2
			mockStream.On("Send", &proto.SyncResponse{
				Response: &proto.SyncResponse_SyncBlock_{
					SyncBlock: common.BuildProtoMsg(2, blocks[2], txIns[2], txOuts[2]),
				},
			}).Return(nil).Once()

			// Client sends request with recent blocks as 1 & 2
			req = &proto.SyncRequest{}
			req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
				Height:       1,
				Hash:         modelBlocks[1].Hash,
				PreviousHash: modelBlocks[1].PreviousHash,
			})
			req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
				Height:       2,
				Hash:         modelBlocks[2].Hash,
				PreviousHash: modelBlocks[2].PreviousHash,
			})
			mockStream.On("Recv").Return(req, nil).Once()

			// Check safe distance
			mockManager.On("GetLatestBlock").Return(modelBlocks[2], nil).Once()

			// Check reorg
			mockManager.On("GetBlocks", []int64{1, 2}).Return(blocks, nil).Once()

			// Have not indexed block 3 yet, retry to get data of block 3 from local at interval in seconds
			mockManager.On("GetBlock", int64(3)).Return(nil, common.ErrNotFound)

			// Context is cancelled, streamer fails to receive msg
			mockStream.On("Recv").Return(nil, context.Canceled).Once()

			go func() {
				time.Sleep(time.Second * time.Duration(client.config.GetBlockIntervalInSec) * 2)
				cancel()
				log.L().Info("Cancel context")
			}()

			err := client.Sync()
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("failed to Receive from Streamer: " + context.Canceled.Error()))
		})

		It("Sequentially sync", func() {
			// Client sends request with recent blocks as 0 & 1
			req := &proto.SyncRequest{}
			req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
				Height:       0,
				Hash:         modelBlocks[0].Hash,
				PreviousHash: modelBlocks[0].PreviousHash,
			})
			req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
				Height:       1,
				Hash:         modelBlocks[1].Hash,
				PreviousHash: modelBlocks[1].PreviousHash,
			})
			mockStream.On("Recv").Return(req, nil).Once()

			// Check safe distance
			mockManager.On("GetLatestBlock").Return(modelBlocks[1], nil).Once()

			// Check reorg
			blocks := make(map[int64]*model.Block)
			blocks[0] = modelBlocks[0]
			blocks[1] = modelBlocks[1]
			mockManager.On("GetBlocks", []int64{0, 1}).Return(blocks, nil).Once()

			// Try to get block 2
			mockManager.On("GetBlock", int64(2)).Return(nil, common.ErrNotFound).Once()
			mockManager.On("GetBlock", int64(2)).Return(nil, common.ErrNotFound).Once()

			go func() {
				time.Sleep(time.Second*time.Duration(client.config.GetBlockIntervalInSec) - time.Microsecond*500)
				mockManager.On("GetBlock", int64(2)).Return(modelBlocks[2], nil).Once()

				// Get data of block 3 from local
				mockAddressWatcher.On("GetAddresses").Return(watchedAddresses).Once()
				txIns := make(map[int64][]*model.TxIn)
				txOuts := make(map[int64][]*model.TxOut)
				blocks[2] = modelBlocks[2]
				txIns[2] = modelTxIns[2]
				txOuts[2] = modelTxOuts[2]
				mockManager.On("GetBlocksData", int64(2), int64(2), watchedAddresses).Return(blocks, txIns, txOuts, nil)

				// Send block 2
				mockStream.On("Send", &proto.SyncResponse{
					Response: &proto.SyncResponse_SyncBlock_{
						SyncBlock: common.BuildProtoMsg(2, blocks[2], txIns[2], txOuts[2]),
					},
				}).Return(nil).Once()

				// Client sends request with recent blocks as 1 & 2
				req = &proto.SyncRequest{}
				req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
					Height:       1,
					Hash:         modelBlocks[1].Hash,
					PreviousHash: modelBlocks[1].PreviousHash,
				})
				req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
					Height:       2,
					Hash:         modelBlocks[2].Hash,
					PreviousHash: modelBlocks[2].PreviousHash,
				})
				mockStream.On("Recv").Return(req, nil).Once()

				// Check safe distance
				mockManager.On("GetLatestBlock").Return(modelBlocks[2], nil).Once()

				// Check reorg
				mockManager.On("GetBlocks", []int64{1, 2}).Return(blocks, nil).Once()

				// Have not indexed block 3 yet, retry to get data of block 3 from local at interval in seconds
				mockManager.On("GetBlock", int64(3)).Return(nil, common.ErrNotFound)

				// Context is cancelled, streamer fails to receive msg
				mockStream.On("Recv").Return(nil, context.Canceled)

				go func() {
					time.Sleep(time.Second * time.Duration(client.config.GetBlockIntervalInSec) * 2)
					cancel()
					log.L().Info("Cancel context")
				}()
			}()

			err := client.Sync()
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("failed to Receive from Streamer: " + context.Canceled.Error()))
		})

		It("Reorg", func() {
			// Client sends request with recent blocks as 2 & 3
			req := &proto.SyncRequest{}
			req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
				Height:       2,
				Hash:         modelBlocks[2].Hash,
				PreviousHash: modelBlocks[2].PreviousHash,
			})
			req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
				Height:       3,
				Hash:         modelBlocks[3].Hash,
				PreviousHash: modelBlocks[3].PreviousHash,
			})
			mockStream.On("Recv").Return(req, nil).Once()

			// Check safe distance
			mockManager.On("GetLatestBlock").Return(modelBlocks[3], nil).Once()

			// Check reorg
			blocks := make(map[int64]*model.Block)
			blocks[2] = modelBlocks[2]
			blocks[3] = modelBlocks[3]
			mockManager.On("GetBlocks", []int64{2, 3}).Return(blocks, nil).Once()

			// Try to get block 3
			mockManager.On("GetBlock", int64(4)).Return(nil, common.ErrNotFound).Once()
			mockManager.On("GetBlock", int64(4)).Return(nil, common.ErrNotFound).Once()

			go func() {
				time.Sleep(time.Second*time.Duration(client.config.GetBlockIntervalInSec) - time.Microsecond*500)
				mockManager.On("GetBlock", int64(4)).Return(reorgModelBlocks[4], nil).Once()

				// Check reorg
				blocks := make(map[int64]*model.Block)
				blocks[2] = modelBlocks[2]
				blocks[3] = reorgModelBlocks[3]
				mockManager.On("GetBlocks", []int64{2, 3}).Return(blocks, nil).Once()

				// Send reorg msg
				mockStream.On("Send", &proto.SyncResponse{
					Response: &proto.SyncResponse_ReorgBlock_{ReorgBlock: &proto.SyncResponse_ReorgBlock{
						Height:  3,
						OldHash: modelBlocks[3].Hash,
						NewHash: reorgModelBlocks[3].Hash,
					}},
				}).Return(nil).Once()

				// Client sends request with recent blocks as 1 & 2
				req := &proto.SyncRequest{}
				req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
					Height:       1,
					Hash:         modelBlocks[1].Hash,
					PreviousHash: modelBlocks[1].PreviousHash,
				})
				req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
					Height:       2,
					Hash:         modelBlocks[2].Hash,
					PreviousHash: modelBlocks[2].PreviousHash,
				})
				mockStream.On("Recv").Return(req, nil).Once()

				// Check safe distance
				mockManager.On("GetLatestBlock").Return(modelBlocks[4], nil).Once()

				// Client should rescan block 3
				mockStream.On("Send", &proto.SyncResponse{
					Response: &proto.SyncResponse_BeginStream_{},
				}).Return(nil).Once()
				mockAddressWatcher.On("GetAddresses").Return(watchedAddresses).Once()
				txIns := make(map[int64][]*model.TxIn)
				txOuts := make(map[int64][]*model.TxOut)
				txIns[3] = modelTxIns[3]
				txOuts[3] = modelTxOuts[3]
				mockManager.On("GetBlocksData", int64(3), int64(3), watchedAddresses).Return(blocks, txIns, txOuts, nil)

				// Send block 3
				mockStream.On("Send", &proto.SyncResponse{
					Response: &proto.SyncResponse_SyncBlock_{
						SyncBlock: common.BuildProtoMsg(3, blocks[3], txIns[3], txOuts[3]),
					},
				}).Return(nil).Once()

				// After sent all safe blocks data, end streaming
				mockStream.On("Send", &proto.SyncResponse{
					Response: &proto.SyncResponse_EndStream_{},
				}).Return(nil).Once()

				// Client sends request with recent blocks as 2 & 3
				req = &proto.SyncRequest{}
				req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
					Height:       2,
					Hash:         modelBlocks[2].Hash,
					PreviousHash: modelBlocks[2].PreviousHash,
				})
				req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
					Height:       3,
					Hash:         reorgModelBlocks[3].Hash,
					PreviousHash: reorgModelBlocks[3].PreviousHash,
				})
				mockStream.On("Recv").Return(req, nil).Once()
				mockManager.On("GetLatestBlock").Return(reorgModelBlocks[3], nil).Once()

				// Check reorg
				mockManager.On("GetBlocks", []int64{2, 3}).Return(blocks, nil).Once()
				mockManager.On("GetBlock", int64(4)).Return(reorgModelBlocks[4], nil).Once()

				// Have no reorg, get data of block 2 from local
				mockAddressWatcher.On("GetAddresses").Return(watchedAddresses).Once()
				blocks[4] = reorgModelBlocks[4]
				txIns[4] = reorgModelTxIns[4]
				txOuts[4] = reorgModelTxOuts[4]
				mockManager.On("GetBlocksData", int64(4), int64(4), watchedAddresses).Return(blocks, txIns, txOuts, nil).Once()

				// Send block 4
				mockStream.On("Send", &proto.SyncResponse{
					Response: &proto.SyncResponse_SyncBlock_{
						SyncBlock: common.BuildProtoMsg(4, blocks[4], txIns[4], txOuts[4]),
					},
				}).Return(nil).Once()

				// Client sends request with recent blocks as 3 & 4
				req = &proto.SyncRequest{}
				req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
					Height:       3,
					Hash:         reorgModelBlocks[3].Hash,
					PreviousHash: reorgModelBlocks[3].PreviousHash,
				})
				req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
					Height:       4,
					Hash:         reorgModelBlocks[4].Hash,
					PreviousHash: reorgModelBlocks[4].PreviousHash,
				})
				mockStream.On("Recv").Return(req, nil).Once()

				// Check safe distance
				mockManager.On("GetLatestBlock").Return(reorgModelBlocks[4], nil).Once()

				// Check reorg
				mockManager.On("GetBlocks", []int64{3, 4}).Return(blocks, nil).Once()

				// Try to get block 5 from local
				mockManager.On("GetBlock", int64(5)).Return(nil, common.ErrNotFound).Once()

				// Context is cancelled, streamer fails to receive msg
				mockStream.On("Recv").Return(nil, context.Canceled)

				go func() {
					time.Sleep(time.Second * time.Duration(client.config.GetBlockIntervalInSec) * 2)
					cancel()
					log.L().Info("Cancel context")
				}()
			}()

			err := client.Sync()
			Expect(err).ShouldNot(BeNil())
			Expect(err.Error()).Should(Equal("failed to Receive from Streamer: " + context.Canceled.Error()))
		})
	})

	Context("Errors happen", func() {
		It("Recv", func() {
			mockStream.On("Recv").Return(nil, context.Canceled).Once()

			err := client.Sync()
			Expect(err).ShouldNot(BeNil())
			errContent := fmt.Sprintf("failed to Receive from Streamer: %v", context.Canceled)
			Expect(err.Error()).Should(Equal(errContent))
		})

		It("GetLatestBlock", func() {
			// Client sends request with empty recent blocks
			req := &proto.SyncRequest{}
			mockStream.On("Recv").Return(req, nil).Once()

			// Check safe distance
			mockManager.On("GetLatestBlock").Return(nil, common.ErrNotFound).Once()

			err := client.Sync()
			Expect(err).ShouldNot(BeNil())
			errContent := fmt.Sprintf("failed to Get Latest Block: %v", common.ErrNotFound)
			errContent = fmt.Sprintf("failed to Make Handler, req '%v': %s", req, errContent)
			Expect(err.Error()).Should(Equal(errContent))
		})

		It("GetBlocksData", func() {
			// Client sends request with empty recent blocks
			req := &proto.SyncRequest{}
			mockStream.On("Recv").Return(req, nil).Once()
			mockManager.On("GetLatestBlock").Return(modelBlocks[2], nil).Once()

			// Client should rescan, start to stream safe blocks data to client
			mockStream.On("Send", &proto.SyncResponse{
				Response: &proto.SyncResponse_BeginStream_{},
			}).Return(nil).Once()
			mockAddressWatcher.On("GetAddresses").Return([]string{}).Once()
			mockManager.On("GetBlocksData", int64(0), int64(1), []string{}).Return(nil, nil, nil, common.ErrNotFound).Once()
			mockAddressWatcher.On("GetAddresses").Return([]string{}).Once()

			err := client.Sync()
			Expect(err).ShouldNot(BeNil())
			errContent := fmt.Sprintf("failed to Get Blocks Data, fromHeight '%d', toHeight '%d', No Of WatchingAddresses '%d': %v", 0, 1, len([]string{}), common.ErrNotFound)
			errContent = fmt.Sprintf("failed to Handle req '%v': %s", req, errContent)
			Expect(err.Error()).Should(Equal(errContent))
		})

		It("Send", func() {
			// Client sends request with empty recent blocks
			req := &proto.SyncRequest{}
			mockStream.On("Recv").Return(req, nil).Once()
			mockManager.On("GetLatestBlock").Return(modelBlocks[2], nil).Once()

			// Client should rescan, start to stream safe blocks data to client
			mockStream.On("Send", &proto.SyncResponse{
				Response: &proto.SyncResponse_BeginStream_{},
			}).Return(nil).Once()
			mockAddressWatcher.On("GetAddresses").Return(watchedAddresses).Once()
			blocks := make(map[int64]*model.Block)
			txIns := make(map[int64][]*model.TxIn)
			txOuts := make(map[int64][]*model.TxOut)
			blocks[0] = modelBlocks[0]
			txIns[0] = modelTxIns[0]
			txOuts[0] = modelTxOuts[0]
			blocks[1] = modelBlocks[1]
			txIns[1] = modelTxIns[1]
			txOuts[1] = modelTxOuts[1]
			mockManager.On("GetBlocksData", int64(0), int64(1), watchedAddresses).Return(blocks, txIns, txOuts, nil).Once()

			// Send each block at a time
			mockStream.On("Send", &proto.SyncResponse{
				Response: &proto.SyncResponse_SyncBlock_{
					SyncBlock: common.BuildProtoMsg(0, blocks[0], txIns[0], txOuts[0]),
				},
			}).Return(context.Canceled).Once()

			err := client.Sync()
			Expect(err).ShouldNot(BeNil())
			errContent := fmt.Sprintf("failed to Send 'Sync Block' msg from Streamer, block '%v': %v", blocks[0], context.Canceled)
			errContent = fmt.Sprintf("failed to Send Blocks Data, fromHeight '%d', toHeight '%d', No Of Blocks '%d', No Of TxIns '%d', No Of TxOuts '%d': %s", 0, 1, len(blocks), len(txIns), len(txOuts), errContent)
			errContent = fmt.Sprintf("failed to Handle req '%v': %s", req, errContent)
			Expect(err.Error()).Should(Equal(errContent))
		})

		It("GetBlocks", func() {
			// Client sends request with recent blocks as 0 & 1
			req := &proto.SyncRequest{}
			req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
				Height:       0,
				Hash:         modelBlocks[0].Hash,
				PreviousHash: modelBlocks[0].PreviousHash,
			})
			req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
				Height:       1,
				Hash:         modelBlocks[1].Hash,
				PreviousHash: modelBlocks[1].PreviousHash,
			})
			mockStream.On("Recv").Return(req, nil).Once()

			// Check safe distance
			mockManager.On("GetLatestBlock").Return(modelBlocks[1], nil).Once()

			// Check reorg
			heights := []int64{0, 1}
			mockManager.On("GetBlocks", heights).Return(nil, common.ErrNotFound).Once()
			err := client.Sync()
			Expect(err).ShouldNot(BeNil())
			errContent := fmt.Sprintf("failed to Get Blocks, block heights '%v': %v", heights, common.ErrNotFound)
			errContent = fmt.Sprintf("failed to Check Reorg: %s", errContent)
			errContent = fmt.Sprintf("failed to Handle req '%v': %s", req, errContent)
			Expect(err.Error()).Should(Equal(errContent))
		})

		It("GetBlock", func() {
			// Client sends request with recent blocks as 0 & 1
			req := &proto.SyncRequest{}
			req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
				Height:       0,
				Hash:         modelBlocks[0].Hash,
				PreviousHash: modelBlocks[0].PreviousHash,
			})
			req.RecentBlocks = append(req.RecentBlocks, &proto.Block{
				Height:       1,
				Hash:         modelBlocks[1].Hash,
				PreviousHash: modelBlocks[1].PreviousHash,
			})
			mockStream.On("Recv").Return(req, nil).Once()

			// Check safe distance
			mockManager.On("GetLatestBlock").Return(modelBlocks[1], nil).Once()

			// Check reorg
			blocks := make(map[int64]*model.Block)
			blocks[0] = modelBlocks[0]
			blocks[1] = modelBlocks[1]
			mockManager.On("GetBlocks", []int64{0, 1}).Return(blocks, nil).Once()

			// Try to get block 2
			mockManager.On("GetBlock", int64(2)).Return(nil, gorm.ErrInvalidSQL).Once()

			err := client.Sync()
			Expect(err).ShouldNot(BeNil())
			errContent := fmt.Sprintf("failed to Get Next Block, height '%d': %v", 2, gorm.ErrInvalidSQL)
			errContent = fmt.Sprintf("failed to Handle req '%v': %s", req, errContent)
			Expect(err.Error()).Should(Equal(errContent))
		})
	})
})

func TestSyncClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sync Client Test")
}
