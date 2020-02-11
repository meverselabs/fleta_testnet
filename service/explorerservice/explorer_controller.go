package explorerservice

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/fletaio/fleta_testnet/common/binutil"
	"github.com/fletaio/fleta_testnet/common/hash"
	"github.com/fletaio/fleta_testnet/core/backend"
	"github.com/fletaio/fleta_testnet/core/chain"
	"github.com/fletaio/fleta_testnet/core/types"
	"github.com/fletaio/fleta_testnet/encoding"
)

type ExplorerController struct {
	db    backend.StoreBackend
	block *BlockExplorer
}

func NewExplorerController(db backend.StoreBackend, block *BlockExplorer) *ExplorerController {
	return &ExplorerController{
		db:    db,
		block: block,
	}
}

func (e *ExplorerController) Blocks(r *http.Request) (map[string]string, error) {
	data := e.block.blocks(0, e.block.provider.Height())
	j, _ := json.Marshal(data)
	return map[string]string{
		"blockData": string(j),
	}, nil
}
func (e *ExplorerController) Transactions(r *http.Request) (map[string]string, error) {
	data := e.block.txs(0, 10)
	j, _ := json.Marshal(data)
	return map[string]string{
		"txsData":  string(j),
		"txLength": strconv.Itoa(e.block.LastestTransactionLen()),
	}, nil
}
func (e *ExplorerController) BlockDetail(r *http.Request) (map[string]string, error) {
	param := r.URL.Query()
	// hash := param.Get("hash")
	heightStr := param.Get("height")
	var height uint32
	if heightStr == "" {
		hash := param.Get("hash")
		if hash == "" {
			return nil, ErrNotEnoughParameter
		}

		if err := e.db.View(func(txn backend.StoreReader) error {
			v, err := txn.Get([]byte(hash))
			if err != nil {
				return err
			}
			if len(v) != 4 {
				return ErrNotBlockHash
			}
			height = binutil.LittleEndian.Uint32(v)
			return nil
		}); err != nil {
			return nil, err
		}

	} else {
		heightInt, err := strconv.Atoi(heightStr)
		if err != nil {
			return nil, ErrInvalidHeightFormat
		}
		height = uint32(heightInt)
	}

	m, err := e.block.blockDetailMap(height)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (e *ExplorerController) TransactionDetail(r *http.Request) (map[string]string, error) {
	param := r.URL.Query()
	hashStr := param.Get("hash")
	h, err := hash.ParseHash(hashStr)
	if err != nil {
		return nil, err
	}
	var v []byte
	if err := e.db.View(func(txn backend.StoreReader) error {
		var err error
		v, err = txn.Get(h[:])
		return err
	}); err != nil {
		return nil, err
	}

	blockHeight, txIndex, err := types.ParseTransactionID(string(v))
	if err != nil {
		return nil, err
	}

	if m, err := e.block.txDetailMap(blockHeight, txIndex); err == nil {
		return m, nil
	} else {
		return nil, err
	}
}

func (e *BlockExplorer) txDetailMap(height uint32, txIndex uint16) (map[string]string, error) {
	m := map[string]interface{}{}

	b, err := e.provider.Block(height)
	if err != nil {
		return nil, err
	}
	t := b.TransactionTypes[int(txIndex)]
	tx := b.Transactions[int(txIndex)]

	fc := encoding.Factory("transaction")

	name, err := fc.TypeName(t)
	if err != nil {
		m["err"] = "현재 지원하지 않는 transaction 입니다."
	}
	m["Type"] = name
	m["Result"] = "Success"

	m["Block Hash"] = encoding.Hash(b.Header).String()

	tm := time.Unix(int64(b.Header.Timestamp/uint64(time.Second)), 0)
	m["Block Timestamp"] = tm.Format("2006-01-02 15:04:05")
	m["Tx Hash"] = chain.HashTransactionByType(e.provider.ChainID(), t, tx).String()
	tm = time.Unix(int64(tx.Timestamp()/uint64(time.Second)), 0)
	m["Tx TimeStamp"] = tm.Format("2006-01-02 15:04:05")

	bs, err := json.Marshal(&m)
	if err != nil {
		return nil, err
	}

	txbs, err := tx.MarshalJSON()
	if err != nil {
		return nil, err
	}
	bs = append(bs[:len(bs)-1], byte(','))
	bs = append(bs, txbs[1:]...)
	return map[string]string{"TxInfo": string(bs)}, nil
}

func (e *BlockExplorer) blockDetailMap(height uint32) (map[string]string, error) {
	h, err := e.provider.Header(height)
	if err != nil {
		return nil, err
	}
	b, err := e.provider.Block(height)
	if err != nil {
		return nil, err
	}

	tm := time.Unix(int64(b.Header.Timestamp/uint64(time.Second)), 0)
	m := map[string]interface{}{}

	tc, err := e.cs.DecodeConsensusData(h.ConsensusData)
	if err != nil {
		return nil, err
	}
	m["Hash"] = encoding.Hash(b.Header).String()
	m["Height"] = strconv.Itoa(int(b.Header.Height))
	m["Version"] = strconv.Itoa(int(b.Header.Version))
	m["HashPrevBlock"] = b.Header.PrevHash.String()
	m["HashLevelRoot"] = b.Header.LevelRootHash.String()
	m["Timestamp"] = tm.Format("2006-01-02 15:04:05")
	m["TimeoutCount"] = strconv.Itoa(int(tc))
	m["FormulationAddress"] = b.Header.Generator.String()
	m["Transaction Count"] = strconv.Itoa(len(b.Transactions))

	txs := []string{}
	for i, tx := range b.Transactions {
		t := b.TransactionTypes[i]
		txs = append(txs, chain.HashTransactionByType(e.provider.ChainID(), t, tx).String())
	}
	m["Transactions"] = txs
	bs, err := json.Marshal(&m)
	return map[string]string{"TxInfo": string(bs)}, nil
}

func (e *ExplorerController) Formulators(r *http.Request) (map[string]string, error) {
	data := e.block.formulators()
	j, _ := json.Marshal(data)
	return map[string]string{
		"formulatorData": string(j),
	}, nil
}
