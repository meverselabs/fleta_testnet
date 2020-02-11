package explorerservice

import (
	"io"
	"strconv"
	"time"

	"github.com/fletaio/fleta_testnet/common/binutil"

	"github.com/fletaio/fleta_testnet/core/types"
	"github.com/fletaio/fleta_testnet/encoding"
)

func (e *BlockExplorer) transactions() []*countInfo {
	if len(e.transactionCountList) > 50 {
		return e.transactionCountList[len(e.transactionCountList)-50:]
	} else {
		return e.transactionCountList
	}
}
func (e *BlockExplorer) chainInfo() currentChainInfo {
	return e.CurrentChainInfo
}

type typePerBlock struct {
	BlockTime uint64 `json:"blockTime"`
	Symbol    string `json:"symbol"`
	TxCount   string `json:"txCount"`
	// Types     map[string]int `json:"types"`
}

type blockInfos struct {
	BlockHeight  uint32   `json:"Block Height"`
	BlockHash    string   `json:"Block Hash"`
	Time         string   `json:"Time"`
	Status       string   `json:"Status"`
	Txs          string   `json:"Txs"`
	Formulator   string   `json:"Formulator"`
	Msg          string   `json:"Msg"`
	Signs        []string `json:"Signs"`
	BlockCount   uint32   `json:"BlockCount"`
	TimeoutCount uint32   `json:"TimeoutCount"`
}
type blockInfosCase struct {
	ITotalRecords        int          `json:"iTotalRecords"`
	ITotalDisplayRecords int          `json:"iTotalDisplayRecords"`
	SEcho                int          `json:"sEcho"`
	SColumns             string       `json:"sColumns"`
	AaData               []blockInfos `json:"aaData"`
}

func (e *BlockExplorer) lastestBlocks() (result blockInfosCase) {
	currHeight := e.provider.Height()

	result.AaData = []blockInfos{}

	for i := currHeight; i > 0 && i > currHeight-8; i-- {
		b, err := e.provider.Block(i)
		if err != nil {
			continue
		}

		bs := types.BlockSign{
			HeaderHash:         encoding.Hash(b.Header),
			GeneratorSignature: b.Signatures[0],
		}
		Signs := []string{
			b.Signatures[1].String(),
			b.Signatures[2].String(),
			b.Signatures[3].String(),
		}

		tm := time.Unix(int64(b.Header.Timestamp/uint64(time.Second)), 0)

		result.AaData = append(result.AaData, blockInfos{
			BlockHeight: i,
			BlockHash:   encoding.Hash(b.Header).String(),
			Time:        tm.Format("2006-01-02 15:04:05"),
			Txs:         strconv.Itoa(len(b.Transactions)),
			Formulator:  b.Header.Generator.String(),
			Msg:         encoding.Hash(bs).String(),
			Status:      "1",
			Signs:       Signs,
			BlockCount:  e.GetBlockCount(b.Header.Generator.String()),
		})
	}

	result.ITotalRecords = len(result.AaData)
	result.ITotalDisplayRecords = len(result.AaData)

	return
}

type txInfos struct {
	TxHash    string `json:"TxHash"`
	BlockHash string `json:"BlockHash"`
	ChainID   string `json:"ChainID"`
	Time      uint64 `json:"Time"`
	TxType    string `json:"TxType"`
}

// WriteTo is a serialization function
func (c *txInfos) WriteTo(w io.Writer) (wrote int64, err error) {
	var wn int64
	{
		wn, err = writeString(w, c.TxHash)
		if err != nil {
			return
		}
		wrote += wn
		wn, err = writeString(w, c.BlockHash)
		if err != nil {
			return
		}
		wrote += wn
		wn, err = writeString(w, c.ChainID)
		if err != nil {
			return
		}
		wrote += wn
		n, err2 := w.Write(binutil.LittleEndian.Uint64ToBytes(c.Time))
		if err2 != nil {
			return wrote, err2
		}
		wrote += int64(n)
		wn, err = writeString(w, c.TxType)
		if err != nil {
			return
		}
		wrote += wn
	}

	return
}

// ReadFrom is a deserialization function
func (c *txInfos) ReadFrom(r io.Reader) (read int64, err error) {
	var rn int64
	{
		c.TxHash, rn, err = readString(r)
		read += rn
		if err != nil {
			return
		}
		c.BlockHash, rn, err = readString(r)
		read += rn
		if err != nil {
			return
		}
		c.ChainID, rn, err = readString(r)
		read += rn
		if err != nil {
			return
		}

		bs := make([]byte, 8)
		n, err2 := r.Read(bs)
		if err2 != nil {
			return read, err2
		}
		read += int64(n)
		c.Time = binutil.LittleEndian.Uint64(bs)

		c.TxType, rn, err = readString(r)
		read += rn
		if err != nil {
			return
		}

	}

	return read, nil
}
func writeString(w io.Writer, str string) (int64, error) {
	var wrote int64
	bs := []byte(str)
	n, err := w.Write(binutil.LittleEndian.Uint16ToBytes(uint16(len(bs))))
	if err != nil {
		return wrote, err
	}
	wrote += int64(n)

	nint, err := w.Write(bs)
	if err != nil {
		return wrote, err
	}
	wrote += int64(nint)

	return wrote, nil
}

func readString(r io.Reader) (string, int64, error) {
	var read int64
	bs := make([]byte, 2)
	n, err := r.Read(bs)
	if err != nil {
		return "", read, err
	}
	read += int64(n)
	bsBs := make([]byte, binutil.LittleEndian.Uint16(bs))

	nInt, err := r.Read(bsBs)
	if err != nil {
		return "", read, err
	}
	read += int64(nInt)

	return string(bsBs), read, nil
}

func (e *BlockExplorer) lastestTransactions() []txInfos {
	if len(e.lastestTransactionList) < 8 {
		return e.lastestTransactionList[0:len(e.lastestTransactionList)]
	}

	return e.lastestTransactionList[0:8]
}

func (e *BlockExplorer) blocks(start int, currHeight uint32) []blockInfos {
	length := 10
	aaData := []blockInfos{}

	for i, j := currHeight-uint32(start), 0; i > 0 && j < length; i, j = i-1, j+1 {
		b, err := e.provider.Block(i)
		if err != nil {
			continue
		}
		tm := time.Unix(int64(b.Header.Timestamp/uint64(time.Second)), 0)

		TimeoutCount, err := e.cs.DecodeConsensusData(b.Header.ConsensusData)
		if err != nil {
			continue
		}
		aaData = append(aaData, blockInfos{
			BlockHeight:  i,
			BlockHash:    encoding.Hash(b.Header).String(),
			Formulator:   b.Header.Generator.String(),
			TimeoutCount: TimeoutCount,
			Time:         tm.Format("2006-01-02 15:04:05"),
			Status:       strconv.Itoa(1),
			Txs:          strconv.Itoa(len(b.Transactions)),
		})
	}

	return aaData
}

func (e *BlockExplorer) paginationBlocks(startStr string) (result blockInfosCase) {
	start, err := strconv.Atoi(startStr)
	if err != nil {
		return
	}
	currHeight := e.provider.Height()

	result.ITotalRecords = int(currHeight)
	result.ITotalDisplayRecords = int(currHeight)

	result.AaData = e.blocks(start, currHeight)

	return
}

type txInfosCase struct {
	ITotalRecords        int       `json:"iTotalRecords"`
	ITotalDisplayRecords int       `json:"iTotalDisplayRecords"`
	SEcho                int       `json:"sEcho"`
	SColumns             string    `json:"sColumns"`
	AaData               []txInfos `json:"aaData"`
}

func (e *BlockExplorer) txs(start int, length int) []txInfos {
	max := start + length
	if max > len(e.lastestTransactionList) {
		max = len(e.lastestTransactionList)
	}

	return e.lastestTransactionList[start:max]
}

func (e *BlockExplorer) paginationTxs(startStr string) (result txInfosCase) {
	start, err := strconv.Atoi(startStr)
	if err != nil {
		return
	}
	length := 10

	result.ITotalRecords = len(e.lastestTransactionList)
	result.ITotalDisplayRecords = len(e.lastestTransactionList)

	result.AaData = e.txs(start, length)

	return
}

type formulatorInfos struct {
	Address    string `json:"Address"`
	Name       string `json:"Name"`
	BlockCount uint32 `json:"BlockCount"`
}

func (e *BlockExplorer) formulators() []formulatorInfos {
	aaData := []formulatorInfos{}

	cs := e.cs.Candidates()
	for _, c := range cs {
		addr := c.Address.String()
		acc, err := e.provider.NewLoaderWrapper(0).Account(c.Address)
		if err != nil {
			aaData = append(aaData, formulatorInfos{
				Address:    addr,
				Name:       "Deleted",
				BlockCount: e.GetBlockCount(addr),
			})
		} else {
			aaData = append(aaData, formulatorInfos{
				Address:    addr,
				Name:       acc.Name(),
				BlockCount: e.GetBlockCount(addr),
			})
		}
	}

	return aaData
}
