package store

import (
	"fmt"
	"github.com/darkknightbk52/btc-indexer/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Manager interface {
	GetLatestBlock() (*model.Block, error)
	GetBlock(height int64) (*model.Block, error)
	GetBlocks(heights []int64) (map[int64]*model.Block, error)
	Reorg(event *model.Reorg) error
	AddBlocksData([]*model.Block, []*model.Tx, []*model.TxIn, []*model.TxOut) error
	GetBlocksData(fromHeight, toHeight int64, interestedAddresses []string) (map[int64]*model.Block, map[int64][]*model.TxIn, map[int64][]*model.TxOut, error)
}

type manager struct {
	db *gorm.DB
}

func NewPostgresManager(connectionString string) (Manager, error) {
	return newManager("postgres", connectionString)
}

func newManager(dialect, dsn string) (Manager, error) {
	db, err := gorm.Open(dialect, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to Open connection with postgres DB, DSN '%s': %v", dsn, err)
	}

	tables := []interface{}{
		model.Block{},
		model.Tx{},
		model.TxIn{},
		model.TxOut{},
		model.Reorg{},
	}
	err = db.AutoMigrate(tables...).Error
	if err != nil {
		var tableNames string
		for _, t := range tables {
			tableNames = fmt.Sprintf("%s %T", tableNames, t)
		}
		return nil, fmt.Errorf("failed to Auto Migrate tables '%s': %v", tableNames, err)
	}
	return &manager{db: db}, nil
}

func (m *manager) GetLatestBlock() (*model.Block, error) {
	b := new(model.Block)
	err := m.db.Order("height DESC").Limit(1).First(b).Error
	return b, err
}

func (m *manager) GetBlock(height int64) (*model.Block, error) {
	b := new(model.Block)
	err := m.db.Where(model.Block{Height: height}).First(b).Error
	return b, err
}

func (m *manager) GetBlocks(heights []int64) (map[int64]*model.Block, error) {
	blocks := make([]*model.Block, 0, len(heights))
	err := m.db.Where("height IN (?)", heights).Find(&blocks).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*model.Block, len(blocks))
	for _, b := range blocks {
		result[b.Height] = b
	}
	return result, nil
}

func (m *manager) newTxManager() (*txManager, error) {
	tx := m.db.Begin()
	if err := tx.Error; err != nil {
		return nil, fmt.Errorf("failed to Begin DB transaction: %v", err)
	}
	return &txManager{
		db:        tx,
		committed: false,
	}, nil
}

func (m *manager) Reorg(event *model.Reorg) error {
	txm, err := m.newTxManager()
	if err != nil {
		return err
	}
	defer txm.maybeRollback()

	for _, table := range []interface{}{
		model.Block{},
		model.Tx{},
		model.TxIn{},
		model.TxOut{},
	} {
		err = txm.db.Delete(table, "height >= (?)", event.FromHeight).Error
		if err != nil {
			return fmt.Errorf("failed to Delete %T from height '%d': %v", table, event.FromHeight, err)
		}
	}

	err = txm.db.Create(event).Error
	if err != nil {
		return fmt.Errorf("failed to Create Reorg event '%v': %v", event, err)
	}

	return txm.commit()
}

func (m *manager) AddBlocksData(blocks []*model.Block, txs []*model.Tx, txIns []*model.TxIn, txOuts []*model.TxOut) error {
	txm, err := m.newTxManager()
	if err != nil {
		return err
	}
	defer txm.maybeRollback()

	err = txm.createBlocks(blocks)
	if err != nil {
		return fmt.Errorf("failed to Create Blocks, Blocks No '%d': %v", len(blocks), err)
	}

	err = txm.createTxs(txs)
	if err != nil {
		return fmt.Errorf("failed to Create Txs, Txs No '%d': %v", len(txs), err)
	}

	err = txm.createTxIns(txIns)
	if err != nil {
		return fmt.Errorf("failed to Create TxIns, TxIns No '%d': %v", len(txIns), err)
	}

	err = txm.createTxOuts(txOuts)
	if err != nil {
		return fmt.Errorf("failed to Create TxOuts, TxOuts No '%d': %v", len(txOuts), err)
	}

	return txm.commit()
}

func (m *manager) GetBlocksData(fromHeight, toHeight int64, interestedAddresses []string) (map[int64]*model.Block, map[int64][]*model.TxIn, map[int64][]*model.TxOut, error) {
	heights := make([]int64, 0, toHeight-fromHeight+1)
	for i := fromHeight; i <= toHeight; i++ {
		heights = append(heights, i)
	}
	blocks, err := m.GetBlocks(heights)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to Get Blocks for Heights '%v': %v", heights, err)
	}

	var txIns []*model.TxIn
	err = m.db.Where("height >= (?) AND height <= (?) AND address in (?)", fromHeight, toHeight, interestedAddresses).Find(&txIns).Error
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to Get TxIns from '%d' to '%d' height and number of addresses '%d': %v", fromHeight, toHeight, len(interestedAddresses), err)
	}
	txInsResult := make(map[int64][]*model.TxIn, len(blocks))
	for _, txIn := range txIns {
		txInsResult[txIn.Height] = append(txInsResult[txIn.Height], txIn)
	}

	var txOuts []*model.TxOut
	err = m.db.Where("height >= (?) AND height <= (?) AND address in (?)", fromHeight, toHeight, interestedAddresses).Find(&txOuts).Error
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to Get TxOuts from '%d' to '%d' height and number of addresses '%d': %v", fromHeight, toHeight, len(interestedAddresses), err)
	}
	txOutsResult := make(map[int64][]*model.TxOut, len(blocks))
	for _, txOut := range txOuts {
		txOutsResult[txOut.Height] = append(txOutsResult[txOut.Height], txOut)
	}

	return blocks, txInsResult, txOutsResult, nil
}
