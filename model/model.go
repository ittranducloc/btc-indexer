package model

const (
	NonStandardAddr = "NonStandard"
)

type Block struct {
	Height       int64  `gorm:"not null;"`
	Hash         string `gorm:"type:varchar(64);not null"`
	PreviousHash string `gorm:"type:varchar(64);not null"`
}

type Tx struct {
	Height   int64  `gorm:"not null;"`
	Hash     string `gorm:"type:varchar(64);not null"`
	CoinBase *bool  `gorm:"not null;default:false"`
}

type TxIn struct {
	TxHash          string `gorm:"type:varchar(64);not null"`
	TxIndex         int32  `gorm:"not null"`
	Height          int64  `gorm:"not null"`
	Address         string `gorm:"type:varchar(62);not null;"` // max length of a bech32 address
	PreviousTxHash  string `gorm:"type:varchar(64);not null"`
	PreviousTxIndex int32  `gorm:"not null"`
}

type TxOut struct {
	Height       int64  `gorm:"not null"`
	TxHash       string `gorm:"type:varchar(64);not null"`
	TxIndex      int32  `gorm:"not null"`
	Value        int64  `gorm:"not null"`
	Address      string `gorm:"type:varchar(62);not null;"` // max length of a bech32 address
	ScriptPubKey []byte `gorm:"not null"`                   // max length 16 MB
	CoinBase     *bool  `gorm:"not null;default:false"`
}

type Reorg struct {
	FromHeight int64
	FromHash   string
	ToHeight   int64
	ToHash     string
}
