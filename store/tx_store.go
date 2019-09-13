package store

import (
	"fmt"
	"github.com/darkknightbk52/btc-indexer/common"
	"github.com/darkknightbk52/btc-indexer/common/log"
	"github.com/darkknightbk52/btc-indexer/model"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
	"strings"
)

type txManager struct {
	db        *gorm.DB
	committed bool
}

func (txm *txManager) commit() error {
	err := txm.db.Commit().Error
	if err != nil {
		return fmt.Errorf("failed to Commit DB transaction: %v", err)
	}
	txm.committed = true
	return nil
}

func (txm *txManager) maybeRollback() {
	if !txm.committed {
		err := txm.db.Rollback().Error
		if err != nil {
			log.L().Warn("failed to Roll Back after error", zap.Error(err))
		}
	}
}

func (txm *txManager) createBlocks(blocks []*model.Block) error {
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		model.Block{}.TableName(),
		strings.Join(model.Block{}.ColumnNames(), ","),
		common.GenerateSqlValuesPart(len(model.Block{}.ColumnNames()), len(blocks)))

	values := make([]interface{}, 0, len(blocks)*len(model.Block{}.ColumnNames()))
	for _, b := range blocks {
		values = append(values, b.Height)
		values = append(values, b.Hash)
		values = append(values, b.PreviousHash)
	}

	return txm.db.Exec(sql, values...).Error
}

func (txm *txManager) createTxs(txs []*model.Tx) error {
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		model.Tx{}.TableName(),
		strings.Join(model.Tx{}.ColumnNames(), ","),
		common.GenerateSqlValuesPart(len(model.Tx{}.ColumnNames()), len(txs)))

	values := make([]interface{}, 0, len(txs)*len(model.Tx{}.ColumnNames()))
	for _, b := range txs {
		values = append(values, b.Height)
		values = append(values, b.Hash)
		values = append(values, b.CoinBase)
	}

	return txm.db.Exec(sql, values...).Error
}

func (txm *txManager) createTxIns(txIns []*model.TxIn) error {
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		model.TxIn{}.TableName(),
		strings.Join(model.TxIn{}.ColumnNames(), ","),
		common.GenerateSqlValuesPart(len(model.TxIn{}.ColumnNames()), len(txIns)))

	values := make([]interface{}, 0, len(txIns)*len(model.TxIn{}.ColumnNames()))
	for _, b := range txIns {
		values = append(values, b.Height)
		values = append(values, b.TxHash)
		values = append(values, b.TxIndex)
		values = append(values, b.Address)
		values = append(values, b.PreviousTxHash)
		values = append(values, b.PreviousTxIndex)
	}

	return txm.db.Exec(sql, values...).Error
}

func (txm *txManager) createTxOuts(txOuts []*model.TxOut) error {
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		model.TxOut{}.TableName(),
		strings.Join(model.TxOut{}.ColumnNames(), ","),
		common.GenerateSqlValuesPart(len(model.TxOut{}.ColumnNames()), len(txOuts)))

	values := make([]interface{}, 0, len(txOuts)*len(model.TxOut{}.ColumnNames()))
	for _, b := range txOuts {
		values = append(values, b.Height)
		values = append(values, b.TxHash)
		values = append(values, b.TxIndex)
		values = append(values, b.Value)
		values = append(values, b.Address)
		values = append(values, b.ScriptPubKey)
		values = append(values, b.CoinBase)
	}

	return txm.db.Exec(sql, values...).Error
}
