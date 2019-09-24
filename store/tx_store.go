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

const postgresParamsLimit = 65535

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
	sql := fmt.Sprintf("INSERT INTO %s (%s)",
		model.Block{}.TableName(),
		strings.Join(model.Block{}.ColumnNames(), ","))

	values := make([]interface{}, 0, len(blocks)*len(model.Block{}.ColumnNames()))
	for _, b := range blocks {
		values = append(values, b.Height)
		values = append(values, b.Hash)
		values = append(values, b.PreviousHash)
	}

	return txm.execSql(sql, values, len(model.Block{}.ColumnNames()))
}

func (txm *txManager) createTxs(txs []*model.Tx) error {
	sql := fmt.Sprintf("INSERT INTO %s (%s)",
		model.Tx{}.TableName(),
		strings.Join(model.Tx{}.ColumnNames(), ","))

	values := make([]interface{}, 0, len(txs)*len(model.Tx{}.ColumnNames()))
	for _, b := range txs {
		values = append(values, b.Height)
		values = append(values, b.Hash)
		values = append(values, b.CoinBase)
	}

	return txm.execSql(sql, values, len(model.Tx{}.ColumnNames()))
}

func (txm *txManager) createTxIns(txIns []*model.TxIn) error {
	sql := fmt.Sprintf("INSERT INTO %s (%s)",
		model.TxIn{}.TableName(),
		strings.Join(model.TxIn{}.ColumnNames(), ","))

	values := make([]interface{}, 0, len(txIns)*len(model.TxIn{}.ColumnNames()))
	for _, b := range txIns {
		values = append(values, b.Height)
		values = append(values, b.TxHash)
		values = append(values, b.TxIndex)
		values = append(values, b.Address)
		values = append(values, b.PreviousTxHash)
		values = append(values, b.PreviousTxIndex)
	}

	return txm.execSql(sql, values, len(model.TxIn{}.ColumnNames()))
}

func (txm *txManager) createTxOuts(txOuts []*model.TxOut) error {
	sql := fmt.Sprintf("INSERT INTO %s (%s)",
		model.TxOut{}.TableName(),
		strings.Join(model.TxOut{}.ColumnNames(), ","))

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

	return txm.execSql(sql, values, len(model.TxOut{}.ColumnNames()))
}

func (txm *txManager) execSql(sql string, values []interface{}, columnNo int) error {
	var sqlParts []string
	var paramsParts [][]interface{}
	for len(values) >= postgresParamsLimit {
		// calculate affordable number of rows be used to build each sql command
		rowNo := postgresParamsLimit / columnNo
		limitIndex := rowNo * columnNo
		sqlParts = append(sqlParts, fmt.Sprintf("%s VALUES %s", sql, common.GenerateSqlValuesPart(columnNo, rowNo)))
		paramsParts = append(paramsParts, values[0:limitIndex])
		values = values[limitIndex:]
	}
	sqlParts = append(sqlParts, fmt.Sprintf("%s VALUES %s", sql, common.GenerateSqlValuesPart(columnNo, len(values)/columnNo)))
	paramsParts = append(paramsParts, values)

	for i, paramsPart := range paramsParts {
		sqlPart := sqlParts[i]
		err := txm.db.Exec(sqlPart, paramsPart...).Error
		if err != nil {
			return err
		}
	}

	return nil
}
