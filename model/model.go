package model

const (
	NonStandardAddr = "NonStandard"
)

type Block struct {
	Height       int64  `gorm:"not null;"`
	Hash         string `gorm:"type:varchar(64);not null"`
	PreviousHash string `gorm:"type:varchar(64);not null"`
}

func (m Block) TableName() string {
	return "blocks"
}

func (m Block) ColumnNames() []string {
	return []string{
		"height",
		"hash",
		"previous_hash",
	}
}

type Tx struct {
	Height   int64  `gorm:"not null;"`
	Hash     string `gorm:"type:varchar(64);not null"`
	CoinBase *bool  `gorm:"not null;default:false"`
}

func (m Tx) TableName() string {
	return "txes"
}

func (m Tx) ColumnNames() []string {
	return []string{
		"height",
		"hash",
		"coin_base",
	}
}

type TxIn struct {
	Height          int64  `gorm:"not null"`
	TxHash          string `gorm:"type:varchar(64);not null"`
	TxIndex         int32  `gorm:"not null"`
	Address         string `gorm:"type:varchar(62);not null;"` // max length of a bech32 address
	PreviousTxHash  string `gorm:"type:varchar(64);not null"`
	PreviousTxIndex int32  `gorm:"not null"`
}

func (m TxIn) TableName() string {
	return "tx_ins"
}

func (m TxIn) ColumnNames() []string {
	return []string{
		"height",
		"tx_hash",
		"tx_index",
		"address",
		"previous_tx_hash",
		"previous_tx_index",
	}
}

type TxOut struct {
	Height       int64  `gorm:"not null"`
	TxHash       string `gorm:"type:varchar(64);not null"`
	TxIndex      int32  `gorm:"not null"`
	Value        int64  `gorm:"not null"`
	Address      string `gorm:"type:varchar(62);not null"` // max length of a bech32 address
	ScriptPubKey []byte `gorm:"not null"`                  // max length 16 MB
	CoinBase     *bool  `gorm:"not null;default:false"`
}

func (m TxOut) TableName() string {
	return "tx_outs"
}

func (m TxOut) ColumnNames() []string {
	return []string{
		"height",
		"tx_hash",
		"tx_index",
		"value",
		"address",
		"script_pub_key",
		"coin_base",
	}
}

type Reorg struct {
	Id         int64  `gorm:"primary"`
	FromHeight int64  `gorm:"not null"`
	FromHash   string `gorm:"not null"`
	ToHeight   int64  `gorm:"not null"`
	ToHash     string `gorm:"not null"`
}
