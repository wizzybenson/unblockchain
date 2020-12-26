package database

type Tx struct {
	To Account `json:"to"`
	From Account `json:"from"`
	Value uint `json:"value"`
	Reason string `json:"reason"`
}

func NewTx(to Account, from Account, value uint, reason string) Tx {
	return Tx{to, from, value, reason}
}

func (tx Tx) IsReward() bool {
	return tx.Reason == "reward"
}