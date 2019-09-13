package common

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/darkknightbk52/btc-indexer/model"
	proto "github.com/darkknightbk52/btc-indexer/proto"
	"strings"
)

func ToBlock(header *btcjson.GetBlockHeaderVerboseResult) *model.Block {
	return &model.Block{
		Height:       int64(header.Height),
		Hash:         header.Hash,
		PreviousHash: header.PreviousHash,
	}
}

// IsASCII is a helper method that checks whether all bytes in `data` would be
// printable ASCII characters if interpreted as a string.
func IsASCII(s string) bool {
	for _, c := range s {
		if c < 32 || c > 126 {
			return false
		}
	}
	return true
}

func GetChainParams(network string) (*chaincfg.Params, error) {
	switch network {
	case "TestNet3":
		return &chaincfg.TestNet3Params, nil
	case "MainNet":
		return &chaincfg.MainNetParams, nil
	default:
		return nil, fmt.Errorf("unsupported Network '%s'", network)
	}
}

func GetAddrFromTxIn(in *wire.TxIn, chainParams *chaincfg.Params) (string, error) {
	pkScript, err := txscript.ComputePkScript(in.SignatureScript, in.Witness)
	if err != nil {
		return model.NonStandardAddr, fmt.Errorf("failed to Compute Public Key Script: %v", err)
	}

	addr, err := pkScript.Address(chainParams)
	if err != nil {
		return model.NonStandardAddr, fmt.Errorf("failed to Encode Public Key Script to Address: %v", err)
	}

	return addr.String(), nil
}

func GetAddrFromTxOut(out *wire.TxOut, chainParams *chaincfg.Params) (string, error) {
	_, addrs, _, err := txscript.ExtractPkScriptAddrs(out.PkScript, chainParams)
	if err != nil {
		return model.NonStandardAddr, fmt.Errorf("failed to Extract Public Key Script: %v", err)
	}

	if len(addrs) == 0 {
		return model.NonStandardAddr, nil
	}
	return addrs[0].String(), nil
}

func BuildProtoMsg(height int64, block *model.Block, txIns []*model.TxIn, txOuts []*model.TxOut) *proto.SyncResponse_SyncBlock {
	msg := new(proto.SyncResponse_SyncBlock)
	msg.Block = &proto.Block{
		Height:       block.Height,
		Hash:         block.Hash,
		PreviousHash: block.PreviousHash,
	}

	for _, txIn := range txIns {
		msg.TxIns = append(msg.TxIns, &proto.TxIn{
			TxHash:          txIn.TxHash,
			TxIndex:         txIn.TxIndex,
			Height:          txIn.Height,
			Address:         txIn.Address,
			PreviousTxHash:  txIn.PreviousTxHash,
			PreviousTxIndex: txIn.PreviousTxIndex,
		})
	}

	for _, txOut := range txOuts {
		msg.TxOuts = append(msg.TxOuts, &proto.TxOut{
			TxHash:       txOut.TxHash,
			TxIndex:      txOut.TxIndex,
			Height:       txOut.Height,
			Value:        txOut.Value,
			Address:      txOut.Address,
			ScriptPubKey: hex.EncodeToString(txOut.ScriptPubKey),
			CoinBase:     *txOut.CoinBase,
		})
	}

	return msg
}

func GenerateSqlValuesPart(columnNo, rowNo int) string {
	questionMasks := make([]string, 0, columnNo)
	for i := 0; i < columnNo; i++ {
		questionMasks = append(questionMasks, "?")
	}
	partForOneColumn := fmt.Sprintf("(%s)", strings.Join(questionMasks, ","))
	columnParts := make([]string, 0, rowNo)
	for i := 0; i < rowNo; i++ {
		columnParts = append(columnParts, partForOneColumn)
	}
	return strings.Join(columnParts, ",")
}
