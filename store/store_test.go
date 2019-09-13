package store

import (
	"fmt"
	"github.com/darkknightbk52/btc-indexer/common/log"
	"github.com/darkknightbk52/btc-indexer/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"os"
	"testing"
)

var (
	store Manager
	db    *gorm.DB
)

var (
	falseValue = false
)

func TestMain(m *testing.M) {
	log.Init(false)

	cfg := struct {
		Host     string
		Port     int
		User     string
		Password string
		DbName   string
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
	err = decoder.Decode(&cfg)
	if err != nil {
		log.S().Fatal(err)
	}

	dbCfg := Config{
		User:     cfg.User,
		Password: cfg.Password,
		Host:     cfg.Host,
		Port:     cfg.Port,
		DBName:   cfg.DbName,
	}
	store, err = NewPostgresManager(dbCfg.DSN())
	if err != nil {
		log.L().Info("Failed to connect to external Postgres DB", zap.String("DSN", dbCfg.DSN()), zap.Error(err))
		store, err = newManager("sqlite3", "./gorm.db")
		if err != nil {
			log.S().Fatal(err)
		}
		log.S().Info("Use Memory DB instead")
	}
	db = store.(*manager).db

	out := m.Run()
	_ = os.Remove("./gorm.db")
	os.Exit(out)
}

func clearDB(t *testing.T) {
	tables := []interface{}{
		model.Block{},
		model.Tx{},
		model.TxIn{},
		model.TxOut{},
		model.Reorg{},
	}
	err := db.DropTable(tables...).Error
	Expect(err).Should(Succeed())
	err = db.CreateTable(tables...).Error
	Expect(err).Should(Succeed())
	log.S().Info("Done clearing DB")
}

func TestNewPostgresManager_Fail(t *testing.T) {
	RegisterTestingT(t)

	dbCfg := Config{
		User:     "notExisted",
		Password: "Password",
		Host:     "Host",
		Port:     1313,
		DBName:   "DbName",
	}
	_, e := gorm.Open("postgres", dbCfg.DSN())
	errContent := fmt.Sprintf("failed to Open connection with postgres DB, DSN '%s': %s", dbCfg.DSN(), e)

	_, err := NewPostgresManager(dbCfg.DSN())
	Expect(err).ShouldNot(Succeed())
	Expect(err.Error()).Should(Equal(errContent))

	log.S().Info(err)
}

func TestManager_GetLatestBlock(t *testing.T) {
	RegisterTestingT(t)
	clearDB(t)

	err := db.Create(model.Block{
		Height:       12,
		Hash:         "12",
		PreviousHash: "11",
	}).Error
	Expect(err).Should(Succeed())

	err = db.Create(model.Block{
		Height:       13,
		Hash:         "13",
		PreviousHash: "12",
	}).Error
	Expect(err).Should(Succeed())

	block, err := store.GetLatestBlock()
	Expect(err).Should(Succeed())
	Expect(block.Height).Should(Equal(int64(13)))
	log.S().Info("latest block:", block)
}

func TestManager_GetBlock(t *testing.T) {
	RegisterTestingT(t)
	clearDB(t)

	err := db.Create(model.Block{
		Height:       13,
		Hash:         "13",
		PreviousHash: "12",
	}).Error
	Expect(err).Should(Succeed())
	block, err := store.GetBlock(13)
	Expect(err).Should(Succeed())
	Expect(block.Height).Should(Equal(int64(13)))
	log.S().Info("block:", block)
}

func TestManager_GetBlocks(t *testing.T) {
	RegisterTestingT(t)
	clearDB(t)

	err := db.Create(model.Block{
		Height:       12,
		Hash:         "12",
		PreviousHash: "11",
	}).Error
	Expect(err).Should(Succeed())

	err = db.Create(model.Block{
		Height:       13,
		Hash:         "13",
		PreviousHash: "12",
	}).Error
	Expect(err).Should(Succeed())

	blocks, err := store.GetBlocks([]int64{12, 13})
	Expect(err).Should(Succeed())
	Expect(len(blocks)).Should(Equal(2))
}

func TestManager_GetBlocksData(t *testing.T) {
	RegisterTestingT(t)
	clearDB(t)

	// ===
	err := db.Create(model.Block{
		Height:       13,
		Hash:         "13",
		PreviousHash: "12",
	}).Error
	Expect(err).Should(Succeed())
	db.Create(model.TxIn{
		Height:          13,
		TxHash:          "tx13",
		TxIndex:         0,
		Address:         "bob",
		PreviousTxHash:  "ptx13",
		PreviousTxIndex: 0,
	})
	Expect(err).Should(Succeed())
	db.Create(model.TxOut{
		Height:       13,
		TxHash:       "tx13",
		TxIndex:      0,
		Address:      "alice",
		Value:        13,
		ScriptPubKey: []byte("key"),
		CoinBase:     &falseValue,
	})
	Expect(err).Should(Succeed())
	db.Create(model.TxIn{
		Height:          13,
		TxHash:          "tx13",
		TxIndex:         1,
		Address:         "bob",
		PreviousTxHash:  "ptx13",
		PreviousTxIndex: 1,
	})
	Expect(err).Should(Succeed())
	db.Create(model.TxOut{
		Height:       13,
		TxHash:       "tx13",
		TxIndex:      1,
		Address:      "alice",
		Value:        13,
		ScriptPubKey: []byte("key"),
		CoinBase:     &falseValue,
	})
	Expect(err).Should(Succeed())

	// ===
	err = db.Create(model.Block{
		Height:       14,
		Hash:         "14",
		PreviousHash: "13",
	}).Error
	Expect(err).Should(Succeed())
	db.Create(model.TxIn{
		Height:          14,
		TxHash:          "tx14",
		TxIndex:         0,
		Address:         "mike",
		PreviousTxHash:  "ptx14",
		PreviousTxIndex: 0,
	})
	Expect(err).Should(Succeed())
	db.Create(model.TxOut{
		Height:       14,
		TxHash:       "tx14",
		TxIndex:      0,
		Address:      "john",
		Value:        13,
		ScriptPubKey: []byte("key"),
		CoinBase:     &falseValue,
	})
	Expect(err).Should(Succeed())
	db.Create(model.TxIn{
		Height:          14,
		TxHash:          "tx14",
		TxIndex:         1,
		Address:         "mike",
		PreviousTxHash:  "ptx14",
		PreviousTxIndex: 1,
	})
	Expect(err).Should(Succeed())
	db.Create(model.TxOut{
		Height:       14,
		TxHash:       "tx14",
		TxIndex:      1,
		Address:      "john",
		Value:        13,
		ScriptPubKey: []byte("key"),
		CoinBase:     &falseValue,
	})
	Expect(err).Should(Succeed())

	blocks, txIns, txOuts, err := store.GetBlocksData(13, 14, []string{"bob", "alice", "mike", "john"})
	Expect(err).Should(Succeed())
	Expect(len(blocks)).Should(Equal(2))
	Expect(txIns[13]).ShouldNot(BeNil())
	Expect(len(txIns[13])).Should(Equal(2))
	Expect(txIns[14]).ShouldNot(BeNil())
	Expect(len(txIns[14])).Should(Equal(2))
	Expect(txOuts[13]).ShouldNot(BeNil())
	Expect(len(txOuts[13])).Should(Equal(2))
	Expect(txOuts[14]).ShouldNot(BeNil())
	Expect(len(txOuts[14])).Should(Equal(2))
	for _, b := range blocks {
		log.S().Info("block:", b)
	}
	for _, ins := range txIns {
		for _, v := range ins {
			log.S().Info("txIn:", v)
		}
	}
	for _, outs := range txOuts {
		for _, v := range outs {
			log.S().Info("txOut:", v)
		}
	}

	blocks, txIns, txOuts, err = store.GetBlocksData(13, 13, []string{"bob", "alice", "mike", "john"})
	Expect(err).Should(Succeed())
	Expect(len(blocks)).Should(Equal(1))
	Expect(txIns[13]).ShouldNot(BeNil())
	Expect(len(txIns[13])).Should(Equal(2))
	Expect(txOuts[13]).ShouldNot(BeNil())
	Expect(len(txOuts[13])).Should(Equal(2))

	blocks, txIns, txOuts, err = store.GetBlocksData(13, 14, []string{"bob", "alice"})
	Expect(err).Should(Succeed())
	Expect(len(blocks)).Should(Equal(2))
	Expect(txIns[13]).ShouldNot(BeNil())
	Expect(len(txIns[13])).Should(Equal(2))
	Expect(txIns[14]).Should(BeNil())
	Expect(txOuts[13]).ShouldNot(BeNil())
	Expect(len(txOuts[13])).Should(Equal(2))
	Expect(txOuts[14]).Should(BeNil())

	blocks, txIns, txOuts, err = store.GetBlocksData(13, 14, []string{"bob", "mike"})
	Expect(err).Should(Succeed())
	Expect(len(blocks)).Should(Equal(2))
	Expect(txIns[13]).ShouldNot(BeNil())
	Expect(len(txIns[13])).Should(Equal(2))
	Expect(txIns[14]).ShouldNot(BeNil())
	Expect(len(txIns[14])).Should(Equal(2))
	Expect(txOuts[13]).Should(BeNil())
	Expect(txOuts[14]).Should(BeNil())

	blocks, txIns, txOuts, err = store.GetBlocksData(13, 14, []string{"alice", "john"})
	Expect(err).Should(Succeed())
	Expect(len(blocks)).Should(Equal(2))
	Expect(txIns[13]).Should(BeNil())
	Expect(txIns[14]).Should(BeNil())
	Expect(txOuts[13]).ShouldNot(BeNil())
	Expect(len(txOuts[13])).Should(Equal(2))
	Expect(txOuts[14]).ShouldNot(BeNil())
	Expect(len(txOuts[14])).Should(Equal(2))

	blocks, txIns, txOuts, err = store.GetBlocksData(113, 114, []string{"bob", "alice", "mike", "john"})
	Expect(err).Should(Succeed())
	Expect(len(blocks)).Should(Equal(0))
	Expect(txIns[113]).Should(BeNil())
	Expect(txOuts[113]).Should(BeNil())
	Expect(txIns[114]).Should(BeNil())
	Expect(txOuts[114]).Should(BeNil())

	blocks, txIns, txOuts, err = store.GetBlocksData(13, 14, []string{"bobCat", "aliceCat", "mikeCat", "johnCat"})
	Expect(err).Should(Succeed())
	Expect(len(blocks)).Should(Equal(2))
	Expect(txIns[13]).Should(BeNil())
	Expect(txOuts[13]).Should(BeNil())
	Expect(txIns[14]).Should(BeNil())
	Expect(txOuts[14]).Should(BeNil())
}

func TestManager_AddBlocksData(t *testing.T) {
	RegisterTestingT(t)
	clearDB(t)

	blocks := []*model.Block{
		{
			Height:       13,
			Hash:         "13",
			PreviousHash: "12",
		},
		{
			Height:       14,
			Hash:         "14",
			PreviousHash: "13",
		},
	}

	txes := []*model.Tx{
		{
			Height:   13,
			Hash:     "tx13",
			CoinBase: &falseValue,
		},
		{
			Height:   14,
			Hash:     "tx14",
			CoinBase: &falseValue,
		},
	}

	txIns := []*model.TxIn{
		{
			Height:          13,
			TxHash:          "tx13",
			TxIndex:         0,
			Address:         "bob",
			PreviousTxHash:  "ptx13",
			PreviousTxIndex: 0,
		},
		{
			Height:          13,
			TxHash:          "tx13",
			TxIndex:         1,
			Address:         "bob",
			PreviousTxHash:  "ptx13",
			PreviousTxIndex: 1,
		},
		{
			Height:          14,
			TxHash:          "tx14",
			TxIndex:         0,
			Address:         "mike",
			PreviousTxHash:  "ptx14",
			PreviousTxIndex: 0,
		},
		{
			Height:          14,
			TxHash:          "tx14",
			TxIndex:         1,
			Address:         "mike",
			PreviousTxHash:  "ptx14",
			PreviousTxIndex: 1,
		},
	}

	txOuts := []*model.TxOut{
		{
			Height:       13,
			TxHash:       "tx13",
			TxIndex:      0,
			Address:      "alice",
			Value:        13,
			ScriptPubKey: []byte("key"),
			CoinBase:     &falseValue,
		},
		{
			Height:       13,
			TxHash:       "tx13",
			TxIndex:      1,
			Address:      "alice",
			Value:        13,
			ScriptPubKey: []byte("key"),
			CoinBase:     &falseValue,
		},
		{
			Height:       14,
			TxHash:       "tx14",
			TxIndex:      0,
			Address:      "john",
			Value:        13,
			ScriptPubKey: []byte("key"),
			CoinBase:     &falseValue,
		},
		{
			Height:       14,
			TxHash:       "tx14",
			TxIndex:      1,
			Address:      "john",
			Value:        13,
			ScriptPubKey: []byte("key"),
			CoinBase:     &falseValue,
		},
	}

	err := store.AddBlocksData(blocks, txes, txIns, txOuts)
	Expect(err).Should(Succeed())

	for _, b := range blocks {
		log.S().Info("block:", b)
	}
	for _, tx := range txes {
		log.S().Info("tx:", tx)
	}
	for _, in := range txIns {
		log.S().Info("txIn:", in)
	}
	for _, out := range txOuts {
		log.S().Info("txOut:", out)
	}
}

func TestManager_Reorg(t *testing.T) {
	RegisterTestingT(t)
	clearDB(t)

	err := db.Create(model.Block{
		Height:       12,
		Hash:         "12",
		PreviousHash: "11",
	}).Error
	Expect(err).Should(Succeed())

	err = db.Create(model.Block{
		Height:       13,
		Hash:         "13",
		PreviousHash: "12",
	}).Error
	Expect(err).Should(Succeed())

	err = db.Create(model.Block{
		Height:       14,
		Hash:         "14",
		PreviousHash: "13",
	}).Error
	Expect(err).Should(Succeed())

	err = store.Reorg(&model.Reorg{
		FromHeight: 14,
		FromHash:   "14",
		ToHeight:   14,
		ToHash:     "14",
	})
	Expect(err).Should(Succeed())

	blocks, err := store.GetBlocks([]int64{12, 13, 14})
	Expect(err).Should(Succeed())
	Expect(len(blocks)).Should(Equal(2))

	err = store.Reorg(&model.Reorg{
		FromHeight: 12,
		FromHash:   "12",
		ToHeight:   13,
		ToHash:     "13",
	})
	Expect(err).Should(Succeed())

	blocks, err = store.GetBlocks([]int64{12, 13, 14})
	Expect(err).Should(Succeed())
	Expect(len(blocks)).Should(Equal(0))
}
