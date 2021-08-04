package database

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"io/ioutil"
)

var genesisJson = `
{
  "genesis_time": "2020-11-12T00:00:00.000000000Z",
  "chain_id": "the-unblockchain-bar-ledger",
  "balances": {
    "0xEa15CaddDA4238E9727ae985c19550DF468B373D": 1000000
  }

}`

type Genesis struct {
	GenesisTime string           `json:"genesis_time"`
	ChainId     string           `json:"chain_id"`
	Balances    map[common.Address]uint `json:"balances"`
}

func loadGenesis(path string) (Genesis, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return Genesis{}, err
	}
	var genesis Genesis
	err = json.Unmarshal(content, &genesis)
	if err != nil {
		return Genesis{}, err
	}
	return genesis, nil

}

func writeGenesisToDisk(path string, genesis []byte) error {
	return ioutil.WriteFile(path, genesis, 0644)
}
