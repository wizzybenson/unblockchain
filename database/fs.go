package database

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func getDatabaseDirPath(dataDir string) string {
	return filepath.Join(dataDir, "database")
}

func getGenesisJsonFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "genesis.json")
}

func getBlocksDBFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "block.db")
}

func fileExist(filepath string) bool {
	_, err := os.Stat(filepath)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func InitDataDir(dataDir string, genesis []byte) error {
	if fileExist(getGenesisJsonFilePath(dataDir)) {
		return nil
	}

	if err := os.MkdirAll(getDatabaseDirPath(dataDir), os.ModePerm); err != nil {
		return err
	}

	if err := writeGenesisToDisk(getGenesisJsonFilePath(dataDir), genesis); err != nil {
		return err
	}

	if err := writeEmptyBlocksDbToDisk(getBlocksDBFilePath(dataDir)); err != nil {
		return err
	}

	return nil
}

func writeEmptyBlocksDbToDisk(path string) error {
	return ioutil.WriteFile(path, []byte(""), os.ModePerm)
}
