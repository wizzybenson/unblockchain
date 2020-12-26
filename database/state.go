package database

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type State struct {
	Balances  map[Account]uint
	TxMempool []Tx
	DbFile    *os.File
	snapshot Snapshot
}

type Snapshot [32]byte

func (s *State) LatestSnapshot() Snapshot {
	return s.snapshot
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
	f, err := os.OpenFile(filepath.Join(cwd, "database", "tx.db"), os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)

	state := &State{balances, make([]Tx, 0), f, Snapshot{}}
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		var tx Tx
		json.Unmarshal(scanner.Bytes(), &tx)

		if err := state.Apply(tx); err != nil {
			return nil, err
		}

	}
	err = state.takeSnapshot()
	if err != nil {
		return nil, err
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

func (s *State) Persist() (Snapshot, error) {
	mempool := make([]Tx, len(s.TxMempool))
	copy(mempool, s.TxMempool)
	for i := 0; i < len(mempool); i++ {
		jsonStr, err := json.Marshal(s.TxMempool[i])
		if err != nil {
			return Snapshot{}, err
		}
		fmt.Printf("Persisting new tx to disk:\n")
		fmt.Printf("\t%s\n", jsonStr)
		if _, err := s.DbFile.Write(append(jsonStr, '\n')); err != nil {
			return Snapshot{}, err
		}
		err = s.takeSnapshot()
		if err != nil {
			return Snapshot{}, err
		}

		fmt.Printf("New DB snapshot: %x\n", s.snapshot)
		s.TxMempool = append(s.TxMempool[:i], s.TxMempool[i+1:]...)
	}

	return s.snapshot, nil
}

func (s *State) Close() error {
	return s.DbFile.Close()
}

func (s *State) takeSnapshot() error {
	_, err := s.DbFile.Seek(0,0)
	if err != nil {
		return err
	}

	txsData, err := ioutil.ReadAll(s.DbFile)
	if err != nil {
		return nil
	}
	s.snapshot = sha256.Sum256(txsData)

	return nil
}
