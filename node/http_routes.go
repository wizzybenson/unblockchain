package node

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github/wizzybenson/unblockchain/database"
	"github/wizzybenson/unblockchain/wallet"
	"net/http"
	"strconv"
)

type ErrRes struct {
	Error string `json:"error"`
}

type BalancesRes struct {
	Hash     database.Hash             `json:"block_hash"`
	Balances map[common.Address]uint `json:"balances"`
}

type StatusRes struct {
	Hash       database.Hash       `json:"block_hash"`
	Number     uint64              `json:"block_number"`
	KnownPeers map[string]PeerNode `json:"peers_known"`
	PendingTxs []database.Tx       `json:"pending_txs"`
}

type SyncRes struct {
	Blocks []database.Block `json:"blocks"`
}

type AddPeerRes struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type TxAddRes struct {
	Success bool `json:"success"`
}

type TxAddReq struct {
	To     string `json:"to"`
	From   string `json:"from"`
	FromPwd string `json:"from_pwd"`
	Value  uint   `json:"value"`
	Reason string `json:"reason"`
}

func showStatus(w http.ResponseWriter, req *http.Request, node *Node) {
	nodeStatus := StatusRes{
		Hash:       node.state.LatestBlockHash(),
		Number:     node.state.LatestBlock().Header.Number,
		KnownPeers: node.knownPeers,
	}

	writeRes(w, nodeStatus)
}

func txAddHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	req := TxAddReq{}
	err := readReq(r, &req)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	from := database.NewAccount(req.From)

	if from.String() == common.HexToAddress("").String() {
		writeErrRes(w, fmt.Errorf("%s is an invalid 'from' sender", from.String()))
		return
	}

	if req.FromPwd == "" {
		writeErrRes(w, fmt.Errorf("password to decrypt the %s account is required. 'from_pwd' is empty", from.String()))
		return
	}

	nonce := node.state.GetNextAccountNonce(from)

	tx := database.NewTx(database.NewAccount(req.To), database.NewAccount(req.From), req.Value, nonce, req.Reason)

	signedTx, err := wallet.SignWithKeystoreAccount(tx, from, req.FromPwd, wallet.GetKeystoreDirPath(node.dataDir))
	if err != nil {
		writeErrRes(w, err)
		return
	}

	err = node.AddPendingTX(signedTx, node.info)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	writeRes(w, TxAddRes{Success: true})
}

func listBalances(w http.ResponseWriter, req *http.Request, state *database.State) {
	writeRes(w, BalancesRes{state.LatestBlockHash(), state.Balances})
}

func addPeerHandler(w http.ResponseWriter, req *http.Request, node *Node) {
	peerIp := req.URL.Query().Get(queryKeyIp)
	peerPortRaw := req.URL.Query().Get(queryKeyPort)
	minerRaw := req.URL.Query().Get(queryKeyMiner)

	peerPort, err := strconv.ParseUint(peerPortRaw, 10, 32)
	if err != nil {
		writeRes(w, AddPeerRes{false, err.Error()})
		return
	}

	peer := NewPeerNode(peerIp, peerPort, false, database.NewAccount(minerRaw),true)

	node.AddPeer(peer)

	fmt.Printf("Peer '%s' was added into knownPeers\n", peer.TcpAddress())

	writeRes(w, AddPeerRes{true, ""})
}

func syncHandler(w http.ResponseWriter, req *http.Request, node *Node) {
	reqHash := req.URL.Query().Get(querykeyFromBlock)

	hash := database.Hash{}
	err := hash.UnmarshalText([]byte(reqHash))
	if err != nil {
		writeErrRes(w, err)
		return
	}

	blocks, err := database.GetBlocksAfter(hash, node.dataDir)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	writeRes(w, SyncRes{Blocks: blocks})
}
