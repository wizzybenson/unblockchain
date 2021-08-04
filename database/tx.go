package database

import (
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"time"
)

func NewAccount(value string) common.Address {
	return common.HexToAddress(value)
}

type Tx struct {
	To     common.Address `json:"to"`
	From   common.Address `json:"from"`
	Nonce uint `json:"nonce"`
	Value  uint    `json:"value"`
	Reason string  `json:"reason"`
	Time   uint64  `json:"time"`
}

type SignedTx struct {
	Tx
	Sig []byte `json:"signature"`
}


func NewTx(to common.Address, from common.Address, value uint, nonce uint, reason string) Tx {
	return Tx{to, from, value, nonce, reason, uint64(time.Now().Unix())}
}

func NewSignedTx(tx Tx, sig []byte) SignedTx {
	return SignedTx{tx, sig}
}

func (tx Tx) IsReward() bool {
	return tx.Reason == "reward"
}

func (tx Tx) Hash() (Hash, error) {
	txJson, err := tx.Encode()
	if err != nil {
		return Hash{}, err
	}

	return sha256.Sum256(txJson), nil
}

func (tx Tx) Encode() ([]byte, error) {
	return json.Marshal(tx)
}

func (tx SignedTx) IsAuthentic() (bool,error) {
	txHash, err := tx.Tx.Hash()
	if err != nil {
		return false, err
	}

	recoveredPubKey, err := crypto.SigToPub(txHash[:], tx.Sig)
	if err != nil {
		return false, err
	}

	recoveredPubKeyBytes := elliptic.Marshal(crypto.S256(), recoveredPubKey.X, recoveredPubKey.Y)
	recoveredPubKeyBytesHash := crypto.Keccak256(recoveredPubKeyBytes[1:])
	recoveredAccount := common.BytesToAddress(recoveredPubKeyBytesHash[12:])

	return recoveredAccount.Hex() == tx.From.Hex(), nil
}


