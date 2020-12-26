package database

import (
	"encoding/json"
	"io/ioutil"
)

type Genesis struct {
	GenesisTime string           `json:"genesis_time"`
	ChainId     string           `json:"chain_id"`
	Balances    map[Account]uint `json:"balances"`
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
