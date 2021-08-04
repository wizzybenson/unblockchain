package node

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github/wizzybenson/unblockchain/database"
	"github/wizzybenson/unblockchain/fs"
	"github/wizzybenson/unblockchain/wallet"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testKsMawAccount = "07a91aaceee215cbaccf5407dbbabbe92d160be7"
const testKsMawFile = "test_maw--07a91aaceee215cbaccf5407dbbabbe92d160be7"
const testKsThanosAccount = "625f385c6b56d03e1eb38c7e34313a3e1898f62c"
const testKsThanosFile = "test_thanos--625f385c6b56d03e1eb38c7e34313a3e1898f62c"
const testKsAccountsPwd = "ernest"

func getTestDataDirPath() (string, error) {
	return ioutil.TempDir(os.TempDir(), ".tub_test")
}

func TestNode_Run(t *testing.T) {
	datadir, err := getTestDataDirPath()
	if err != nil {
		t.Fatal(err)
	}

	err = fs.RemoveDir(datadir)
	if err != nil {
		t.Fatal(err)
	}

	n := New(
		datadir,
		"127.0.0.1",
		8086,
		database.NewAccount(DefaultMiner),
		PeerNode{},
	)

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	err = n.Run(ctx)
	if err.Error() != "http: Server closed" {
		t.Fatal("node server was supposed to close after 5s")
	}
}

func TestNode_Mining(t *testing.T) {
	datadir, thanos, maw, err := setUpTestNodeDir()
	if err != nil {
		t.Error(err)
	}
	defer fs.RemoveDir(datadir)

	nInfo := NewPeerNode("127.0.0.1", 8087, false, database.NewAccount(""), true)

	n := New(datadir, nInfo.IP, nInfo.Port, thanos, nInfo)
	ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*30)

	go func() {
		time.Sleep(time.Second * miningIntervalSeconds / 3)
		tx := database.NewTx(maw, thanos, 1, 1,"")
		signedTx, err := wallet.SignWithKeystoreAccount(tx, thanos, testKsAccountsPwd, wallet.GetKeystoreDirPath(datadir))
		if err != nil {
			t.Error(err)
			return
		}

		_ = n.AddPendingTX(signedTx, nInfo)
	}()

	go func() {
		time.Sleep(time.Second*miningIntervalSeconds + 2)

		tx := database.NewTx(maw, thanos, 2, 1,"")
		signedTx, err := wallet.SignWithKeystoreAccount(tx, thanos, testKsAccountsPwd, wallet.GetKeystoreDirPath(datadir))
		if err != nil {
			t.Error(err)
			return
		}

		_ = n.AddPendingTX(signedTx, nInfo)
	}()

	go func() {
		ticker := time.NewTicker(10 * time.Second)

		for {
			select {
			case <-ticker.C:
				if n.state.LatestBlock().Header.Number == 2 {
					closeNode()
					return
				}
			}
		}
	}()

	_ = n.Run(ctx)

	if n.state.LatestBlock().Header.Number != 1 {
		t.Fatal("Was supposed to mine 2 pending TX into 2 valid blocks under 30 minutes")
	}
}

func TestNode_ForgedTx(t *testing.T) {
	datadir, thanos, maw, err := setUpTestNodeDir()
	if err != nil {
		t.Error(err)
	}
	defer fs.RemoveDir(datadir)

	n := New(datadir, "127.0.0.1", 8087, thanos, PeerNode{})
	ctx, _ := context.WithTimeout(context.Background(), time.Minute * 15)
	thanosPeerNode := NewPeerNode("127.0.0.1", 8087, false, thanos,true)

	txValue := uint(5)
	txNonce := uint(1)
	tx := database.NewTx(maw, thanos, txValue, txNonce, "")

	signedTx, err := wallet.SignWithKeystoreAccount(tx, thanos, testKsAccountsPwd, wallet.GetKeystoreDirPath(datadir))
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		time.Sleep(time.Second * 1)

		_ = n.AddPendingTX(signedTx, thanosPeerNode)
	}()

	go func() {
		time.Sleep(time.Second * (miningIntervalSeconds + 1))
		forgedTx := database.NewTx(maw, thanos, txValue, txNonce, "")
		forgedSignedTx := database.NewSignedTx(forgedTx, signedTx.Sig)

		_ = n.AddPendingTX(forgedSignedTx, thanosPeerNode)
	}()

	_ = n.Run(ctx)

	if n.state.LatestBlock().Header.Number != 0 {
		t.Fatal("was supposed to mine only one TX. The second was forged")
	}
}

func TestNode_ReplayedTx(t *testing.T) {
	datadir, thanos, maw, err := setUpTestNodeDir()
	if err != nil {
		t.Error(err)
	}
	defer fs.RemoveDir(datadir)

	n := New(datadir, "127.0.0.1", 8087, thanos, PeerNode{})
	ctx, closeNode := context.WithCancel(context.Background())
	thanosPeerNode := NewPeerNode("127.0.0.1", 8087, false, thanos,true)
	mawPeerNode := NewPeerNode("127.0.0.1", 8088, false, maw,true)

	txValue := uint(5)
	txNonce := uint(1)
	tx := database.NewTx(maw, thanos, txValue, txNonce, "")

	signedTx, err := wallet.SignWithKeystoreAccount(tx, thanos, testKsAccountsPwd, wallet.GetKeystoreDirPath(datadir))
	if err != nil {
		t.Error(err)
		return
	}

	_ = n.AddPendingTX(signedTx, thanosPeerNode)

	go func() {
		ticker := time.NewTicker(time.Second * (miningIntervalSeconds - 3))
		wasReplayedTxAdded := false

		for {
			select {
			case <-ticker.C:
				if n.state.LatestBlock().Header.Number == 0 {
					if wasReplayedTxAdded && !n.isMining {
						closeNode()
						return
					}

					n.archivedTXs = make(map[string]database.SignedTx)

					_ = n.AddPendingTX(signedTx, mawPeerNode)
					wasReplayedTxAdded = true
				}

				if n.state.LatestBlock().Header.Number == 1 {
					closeNode()
					return
				}
			}
		}
	}()

	_ = n.Run(ctx)

	if n.state.Balances[maw] == txValue * 2 {
		t.Error("replayed attack was successful :( damn digital signatures!")
		return
	}
}

func TestNode_MiningStopsOnNewSyncedBlock(t *testing.T) {
	datadir, thanos, maw, err := setUpTestNodeDir()
	if err != nil {
		t.Error(err)
	}

	defer fs.RemoveDir(datadir)

	nInfo := NewPeerNode("127.0.0.1", 8087, false, database.NewAccount(""), true)

	n := New(datadir, nInfo.IP, nInfo.Port, thanos, nInfo)
	ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*30)

	tx := database.NewTx(maw, thanos, 1, 1,"")
	tx2 := database.NewTx(maw, thanos, 2, 1,"")

	signedTx, err := wallet.SignWithKeystoreAccount(tx, thanos, testKsAccountsPwd, wallet.GetKeystoreDirPath(datadir))
	if err != nil {
		t.Error(err)
		return
	}

	signedTx2, err := wallet.SignWithKeystoreAccount(tx, thanos, testKsAccountsPwd, wallet.GetKeystoreDirPath(datadir))
	if err != nil {
		t.Error(err)
		return
	}
	tx2Hash, _ := signedTx2.Hash()

	validPreMinedPb := NewPendingBlock(
		database.Hash{},
		1,
		thanos,
		[]database.SignedTx{signedTx},
	)
	validSyncedBlock, err := Mine(ctx, validPreMinedPb)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(time.Second * (miningIntervalSeconds - 2))

		err := n.AddPendingTX(signedTx, nInfo)
		if err != nil {
			t.Fatal(err)
		}

		err = n.AddPendingTX(signedTx2, nInfo)
		if err != nil {
			t.Fatal(err)
		}
	}()

	go func() {
		time.Sleep(time.Second * (miningIntervalSeconds + 2))
		if !n.isMining {
			t.Fatal("should be mining")
		}

		_, err := n.state.AddBlock(validSyncedBlock)
		if err != nil {
			t.Fatal(err)
		}
		n.newSyncedBlocks <- validSyncedBlock

		time.Sleep(time.Second * 2)
		if n.isMining {
			t.Fatal("new received block should have cancelled mining")
		}

		_, onlyTX2IsPending := n.pendingTXs[tx2Hash.Hex()]

		if len(n.pendingTXs) != 1 && !onlyTX2IsPending {
			t.Fatal("new received block should have cancelled mining of already mined transaction")
		}

		time.Sleep(time.Second * (miningIntervalSeconds + 2))
		if !n.isMining {
			t.Fatal("should be mining again the 1 tx not included in synced block")
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second * 10)

		for {
			select {
			case <-ticker.C:
				if n.state.LatestBlock().Header.Number == 1 {
					closeNode()
					return
				}
			}
		}
	}()

	go func() {
		time.Sleep(time.Second * 2)

		startingThanosBalance := n.state.Balances[thanos]
		startingMawBalance := n.state.Balances[maw]

		<-ctx.Done()

		endThanosBalance := n.state.Balances[thanos]
		endMawBalances := n.state.Balances[maw]

		expectedEndThanosBalance := startingThanosBalance - tx.Value - tx2.Value + database.BlockReward
		expectedEndMawBalance := startingMawBalance + tx.Value + tx2.Value + database.BlockReward

		if endThanosBalance != expectedEndThanosBalance {
			t.Fatalf("Thanos expected end balance is %d not %d", expectedEndThanosBalance, endThanosBalance)
		}

		if endMawBalances != expectedEndMawBalance {
			t.Fatalf("BabaYaga expected end balance is %d not %d", expectedEndMawBalance, endMawBalances)
		}

		t.Logf("Starting Thanos balance: %d", startingThanosBalance)
		t.Logf("Starting Maw balance: %d", startingMawBalance)
		t.Logf("Ending Thanos balance: %d", endThanosBalance)
		t.Logf("Ending Maw balance: %d", endMawBalances)
	}()

	_ = n.Run(ctx)

	if n.state.LatestBlock().Header.Number != 1 {
		t.Fatal("was suppose to mine 2 pending TX into 2 valid blocks under 30m")
	}

	if len(n.pendingTXs) != 0 {
		t.Fatal("no pending TXs should be left to mine")
	}

}

func copyKeystoreFilesIntoTestDataDirPath(datadir string) error {
	thanosSrcKs, err := os.Open(testKsThanosFile)
	if err != nil {
		return err
	}
	
	defer thanosSrcKs.Close()
	
	ksDir := filepath.Join(wallet.GetKeystoreDirPath(datadir))
	
	err = os.Mkdir(ksDir, 0777)
	if err != nil {
		return err
	}
	
	thanosDstKs, err := os.Create(filepath.Join(ksDir, testKsThanosFile))
	if err != nil {
		return err
	}
	
	defer thanosDstKs.Close()
	
	_, err = io.Copy(thanosDstKs, thanosSrcKs)
	if err != nil {
		return err
	}

	mawSrcKs, err := os.Open(testKsMawFile)
	if err != nil {
		return err
	}

	defer mawSrcKs.Close()
	
	mawDstKs, err := os.Create(filepath.Join(ksDir, testKsMawFile))
	if err != nil {
		return err
	}

	defer mawDstKs.Close()

	_, err = io.Copy(mawDstKs, mawSrcKs)
	if err != nil {
		return err
	}

	return nil

}

func setUpTestNodeDir() (datadir string, thanos, maw common.Address, err error) {
	thanos = database.NewAccount(testKsThanosAccount)
	maw = database.NewAccount(testKsMawAccount)

	genesisBalances := make(map[common.Address]uint)
	genesisBalances[thanos] = 1000000
	genesis := database.Genesis{Balances: genesisBalances}
	genesisJson, err := json.Marshal(genesis)
	if err != nil {
		return "", common.Address{}, common.Address{}, err
	}

	datadir, err = getTestDataDirPath()
	if err != nil {
		return "", common.Address{}, common.Address{}, err
	}

	err = database.InitDataDir(datadir, genesisJson)
	if err != nil {
		return "", common.Address{}, common.Address{}, err
	}

	err = copyKeystoreFilesIntoTestDataDirPath(datadir)
	if err != nil {
		return "", common.Address{}, common.Address{}, err
	}

	return datadir, thanos, maw, nil
}
