package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type State struct {
	Balances        map[Account]uint
	TxMempool       []Tx
	DbFile          *os.File
	latestBlockHash Hash
}

func (s *State) LatestBlockHash() Hash {
	return s.latestBlockHash
}

func NewStateFromDisk() (*State, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	genesis, err := loadGenesis(filepath.Join(cwd, "database", "genesis.json"))
	if err != nil {
		return nil, err
	}
	balances := genesis.Balances
	f, err := os.OpenFile(filepath.Join(cwd, "database", "block.db"), os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)

	state := &State{balances, make([]Tx, 0), f, Hash{}}
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		var blockFs BlockFs
		json.Unmarshal(scanner.Bytes(), &blockFs)

		if err := state.applyBlock(blockFs.Value); err != nil {
			return nil, err
		}
		state.latestBlockHash = blockFs.Key
	}
	return state, nil
}

func (s *State) Apply(tx Tx) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if s.Balances[tx.From] < tx.Value {
		return fmt.Errorf("insufficeint fund")
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value
	return nil
}

func (s *State) Add(tx Tx) error {
	if err := s.Apply(tx); err != nil {
		return err
	}
	s.TxMempool = append(s.TxMempool, tx)
	return nil
}

func (s *State) Persist() (Hash, error) {
	block := NewBlock(s.latestBlockHash, uint64(time.Now().Unix()), s.TxMempool)
	blockHash, err := block.Hash()
	if err != nil {
		return Hash{}, err
	}

	blockFs := BlockFs{blockHash, block}
	blockFsJsonStr, err := json.Marshal(blockFs)
	if err != nil {
		return Hash{}, err
	}
	fmt.Printf("Persisting new block to disk:\n")
	fmt.Printf("\t%s\n", blockFsJsonStr)
	if _, err := s.DbFile.Write(append(blockFsJsonStr, '\n')); err != nil {
		return Hash{}, err
	}
	s.latestBlockHash = blockHash
	s.TxMempool = []Tx{}

	return s.latestBlockHash, nil
}

func (s *State) Close() error {
	return s.DbFile.Close()
}

func (s *State) applyBlock(b Block) error {
	for _, tx := range b.Txs {
		if err := s.Apply(tx); err != nil {
			return err
		}
	}
	return nil
}

func (s *State) AddBlock(b Block) error {
	for _, tx := range b.Txs {
		if err := s.Add(tx); err != nil {
			return nil
		}
	}
	return nil
}
