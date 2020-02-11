package explorerservice

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fletaio/fleta_testnet/core/backend"
	_ "github.com/fletaio/fleta_testnet/core/backend/badger_driver"
	"github.com/fletaio/webserver"

	"github.com/fletaio/fleta_testnet/pof"

	"github.com/fletaio/fleta_testnet/common/binutil"
	"github.com/fletaio/fleta_testnet/common/factory"
	"github.com/fletaio/fleta_testnet/core/types"
	"github.com/fletaio/fleta_testnet/encoding"
	"github.com/labstack/echo/v4"
)

var (
	libPath string
)

func init() {
	var pwd string
	{
		pc := make([]uintptr, 10) // at least 1 entry needed
		runtime.Callers(1, pc)
		f := runtime.FuncForPC(pc[0])
		pwd, _ = f.FileLine(pc[0])

		path := strings.Split(pwd, "/")
		pwd = strings.Join(path[:len(path)-1], "/")
	}

	libPath = pwd
}

//Block explorer error list
var (
	ErrDbNotClear               = errors.New("Db is not clear")
	ErrNotEnoughParameter       = errors.New("Not enough parameter")
	ErrNotTransactionHash       = errors.New("This hash is not a transaction hash")
	ErrAlreadyRegistrationBlock = errors.New("Already registration block")
	ErrNotBlockHash             = errors.New("This hash is not a block hash")
	ErrInvalidHeightFormat      = errors.New("Invalid height format")
)

// BlockExplorer struct
type BlockExplorer struct {
	types.ServiceBase
	provider               types.Provider
	transactionCountList   []*countInfo
	CurrentChainInfo       currentChainInfo
	lastestTransactionList []txInfos
	lastestTransactionMap  map[string]bool

	db backend.StoreBackend

	cs *pof.Consensus

	initURLFlag      bool
	e                *echo.Echo
	webChecker       echo.MiddlewareFunc
	dataHandlerPacks []DataHandlerPack

	MaximumTps int

	port int
}

type countInfo struct {
	Time  int64 `json:"time"`
	Count int   `json:"count"`
}

//NewBlockExplorer TODO
func NewBlockExplorer(dbPath string, cs *pof.Consensus, port int) (*BlockExplorer, error) {
	DB, err := backend.Create("badger", "./ndata/explorer")
	if err != nil {
		panic(err)
	}

	e := &BlockExplorer{
		transactionCountList:   []*countInfo{},
		lastestTransactionList: []txInfos{},
		db:                     DB,
		cs:                     cs,
		dataHandlerPacks:       []DataHandlerPack{},
		port:                   port,
	}

	if err := e.db.View(func(txn backend.StoreReader) error {
		value, err := txn.Get(blockChainInfoBytes)
		if err != nil {
			if err != backend.ErrNotExistKey {
				return err
			}
		} else {
			buf := bytes.NewBuffer(value)
			e.CurrentChainInfo.ReadFrom(buf)
		}

		value, err = txn.Get(MaximumTpsBytes)
		if err != nil {
			if err != backend.ErrNotExistKey {
				return err
			}
		} else {
			tps := binutil.LittleEndian.Uint32(value)
			e.MaximumTps = int(tps)
		}

		return nil
	}); err != nil {
		return nil, ErrDbNotClear
	}

	return e, nil
}

func (e *BlockExplorer) Name() string {
	return "BlockExplorer"
}
func (e *BlockExplorer) Init(pm types.ProcessManager, p types.Provider) error {
	e.provider = p
	return nil
}

var reverseOrderedTx = []byte("orderedTx")

func (e *BlockExplorer) OnLoadChain(loader types.Loader) error {
	go e.StartExplorer(e.port)

	func(e *BlockExplorer) {
		initHeightKey := []byte("initHeightKey")
		var startHeight uint32
		e.db.View(func(txn backend.StoreReader) error {
			value, err := txn.Get(initHeightKey)
			if err != nil {
				if err != backend.ErrNotExistKey {
					return err
				}
				startHeight = 0
			} else {
				startHeight = binutil.LittleEndian.Uint32(value)
			}

			return nil
		})

		fc := encoding.Factory("transaction")
		currHeight := e.provider.Height()
		for i := startHeight; i < currHeight; i++ {
			b, err := e.provider.Block(i)
			if err != nil {
				continue
			}
			e.updateChain(b, fc, nil)

			if err := e.db.Update(func(txn backend.StoreWriter) error {
				txn.Set(initHeightKey, binutil.LittleEndian.Uint32ToBytes(i+1))
				return nil
			}); err != nil {
				fmt.Errorf("err : %v", err)
			}
		}

	}(e)

	e.db.View(func(txn backend.StoreReader) error {
		prefix := []byte(reverseOrderedTx)
		//Iterate(prefix []byte, fn func(key []byte, value []byte) error) error
		err := txn.Iterate(prefix, func(key []byte, value []byte) error {
			buf := bytes.NewBuffer(value)
			ti := &txInfos{}
			ti.ReadFrom(buf)
			e.txinfoInsertSort(*ti)
			return nil
		})
		return err
	})
	return nil
}

func (e *BlockExplorer) txinfoInsertSort(el txInfos) {
	index := sort.Search(len(e.lastestTransactionList), func(i int) bool { return e.lastestTransactionList[i].Time < el.Time })
	e.lastestTransactionList = append(e.lastestTransactionList, txInfos{})
	copy(e.lastestTransactionList[index+1:], e.lastestTransactionList[index:])
	e.lastestTransactionList[index] = el
	if len(e.lastestTransactionList) > 500 {
		e.lastestTransactionList = e.lastestTransactionList[0:500]
	}
}

func (e *BlockExplorer) countinfoInsertSort(el *countInfo) {
	index := sort.Search(len(e.transactionCountList), func(i int) bool {
		return e.transactionCountList[i].Time > el.Time
	})

	if index >= 0 && index <= len(e.transactionCountList) {
		dasv := func(index int) bool {
			t := time.Unix(el.Time/1000000000, 0)
			t1 := time.Unix(e.transactionCountList[index].Time/1000000000, 0)
			if t.Second() == t1.Second() && t.Minute() == t1.Minute() {
				e.transactionCountList[index].Count += el.Count
				e.updateMaximumTps(e.transactionCountList[index].Count)
				return true
			}
			return false
		}
		var result bool
		if index < len(e.transactionCountList) {
			result = dasv(index) || result
		}
		if index < len(e.transactionCountList)-1 {
			result = dasv(index+1) || result
		}
		if index > 0 {
			result = dasv(index-1) || result
		}
		if result {
			return
		}
	}

	e.transactionCountList = append(e.transactionCountList, &countInfo{})
	copy(e.transactionCountList[index+1:], e.transactionCountList[index:])
	e.transactionCountList[index] = el
	if len(e.transactionCountList) > 500 {
		e.transactionCountList = e.transactionCountList[len(e.transactionCountList)-500 : len(e.transactionCountList)]
	}
}

func (e *BlockExplorer) updateMaximumTps(count int) {
	if e.MaximumTps < count {
		e.MaximumTps = count
		if err := e.db.Update(func(txn backend.StoreWriter) error {
			txn.Set(MaximumTpsBytes, binutil.LittleEndian.Uint32ToBytes(uint32(e.MaximumTps)))
			return nil
		}); err != nil {
			fmt.Errorf("err : %v", err)
		}
	}

}

func (e *BlockExplorer) OnBlockConnected(b *types.Block, events []types.Event, loader types.Loader) {
	fc := encoding.Factory("transaction")
	e.updateChain(b, fc, e.txinfoInsertSort)
}

func (e *BlockExplorer) updateChain(b *types.Block, fc *factory.Factory, insertTx func(el txInfos)) {
	e.db.Update(func(txn backend.StoreWriter) error {
		// txn.Set(MaximumTpsBytes, binutil.LittleEndian.Uint32ToBytes(uint32(e.MaximumTps)))
		_, err := txn.Get([]byte(encoding.Hash(b.Header).String()))
		if err != backend.ErrNotExistKey {
			return ErrAlreadyRegistrationBlock
		}

		e.CurrentChainInfo.currentTransactions = len(b.Transactions)
		if e.CurrentChainInfo.Blocks < b.Header.Height {
			e.CurrentChainInfo.Blocks = b.Header.Height
		}

		e.countinfoInsertSort(&countInfo{
			Time:  int64(b.Header.Timestamp),
			Count: len(b.Transactions),
		})

		value := binutil.LittleEndian.Uint32ToBytes(b.Header.Height | 0xFFFFFFFF)

		txs := b.Transactions
		for i, tx := range txs {
			t := b.TransactionTypes[i]
			name, err := fc.TypeName(t)
			if err != nil {
				name = "UNKNOWN"
			} else {
				strs := strings.Split(name, "/")
				name = strs[len(strs)-1]
			}
			ti := txInfos{
				TxHash:    types.HashTransactionByType(e.provider.ChainID(), t, tx).String(),
				BlockHash: encoding.Hash(b.Header).String(),
				Time:      tx.Timestamp(),
				TxType:    name,
			}
			if insertTx != nil {
				insertTx(ti)
			}
			iBs := binutil.LittleEndian.Uint32ToBytes(uint32(i) | 0xFFFFFFFF)
			v := append(value, iBs...)

			buf := &bytes.Buffer{}
			ti.WriteTo(buf)

			if err := txn.Set(append(reverseOrderedTx, v...), buf.Bytes()); err != nil {
				return err
			}

		}

		err = e.updateHashs(txn, b, fc)
		if err != nil {
			return err
		}

		buf := &bytes.Buffer{}
		_, err = e.CurrentChainInfo.WriteTo(buf)
		if err != nil {
			return err
		}

		e.CurrentChainInfo.Transactions += e.CurrentChainInfo.currentTransactions

		cs := e.cs.Candidates()
		e.CurrentChainInfo.Foumulators = len(cs)
		txn.Set(blockChainInfoBytes, buf.Bytes())
		return nil
	})
}

var blockChainInfoBytes = []byte("blockChainInfo")
var MaximumTpsBytes = []byte("MaximumTps")

// LastestTransactionLen is returned length of lastest txs
func (e *BlockExplorer) LastestTransactionLen() int {
	return len(e.lastestTransactionList)
}

func (e *BlockExplorer) updateHashs(txn backend.StoreWriter, b *types.Block, fc *factory.Factory) error {
	value := binutil.LittleEndian.Uint32ToBytes(b.Header.Height)

	h := encoding.Hash(b.Header).String()
	if err := txn.Set([]byte(h), value); err != nil {
		return err
	}

	formulatorAddr := []byte("formulator" + b.Header.Generator.String()) //FIXME
	value, err := txn.Get(formulatorAddr)
	if err != nil {
		if err != backend.ErrNotExistKey {
			return err
		}
		txn.Set(formulatorAddr, binutil.LittleEndian.Uint32ToBytes(1))
	} else {
		height := binutil.LittleEndian.Uint32(value)
		txn.Set(formulatorAddr, binutil.LittleEndian.Uint32ToBytes(height+1))
	}

	txs := b.Transactions
	for i, tx := range txs {
		t := b.TransactionTypes[i]

		h := types.HashTransactionByType(e.provider.ChainID(), t, tx)
		txid := types.TransactionID(b.Header.Height, uint16(i))
		if err := txn.Set(h[:], []byte(txid)); err != nil {
			return err
		}

	}
	return nil
}

// GetBlockCount return block height
func (e *BlockExplorer) GetBlockCount(formulatorAddr string) (height uint32) {
	formulatorKey := []byte("formulator" + formulatorAddr)
	e.db.View(func(txn backend.StoreReader) error {
		value, err := txn.Get(formulatorKey)
		if err != nil {
			if err != backend.ErrNotExistKey {
				return err
			}
			height = 0
		} else {
			height = binutil.LittleEndian.Uint32(value)
		}

		return nil
	})
	return
}

// InitURL is initialization urls
func (e *BlockExplorer) InitURL() {
	e.initURLFlag = true

	assets := webserver.NewFileAsset(Assets, "webfiles")
	web := webserver.NewWebServer(assets, "webfiles", func() {})
	e.e = web.Echo
	e.e.Renderer = web

	ec := NewExplorerController(e.db, e)

	fs := http.FileServer(assets)
	e.e.GET("/resource/*", echo.WrapHandler(fs))

	e.webChecker = func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			web.CheckWatch()
			return next(c)
		}
	}

	e.e.Any("/data/:order", e.dataHandler)
	e.e.GET("/", func(c echo.Context) error {
		args := map[string]string{
			"MaximumTps": fmt.Sprintln(e.MaximumTps),
		}
		err := c.Render(http.StatusOK, "index.html", args)
		if err != nil {
			log.Println(err)
		}
		return err
	}, e.webChecker)
	e.e.GET("/blocks", func(c echo.Context) error {
		args, err := ec.Blocks(c.Request())
		if err != nil {
			log.Println(err)
		}
		err = c.Render(http.StatusOK, "blocks.html", args)
		if err != nil {
			log.Println(err)
		}
		return err
	}, e.webChecker)
	e.e.GET("/blockDetail", func(c echo.Context) error {
		args, err := ec.BlockDetail(c.Request())
		if err != nil {
			log.Println(err)
		}
		err = c.Render(http.StatusOK, "blockDetail.html", args)
		if err != nil {
			log.Println(err)
		}
		return err
	}, e.webChecker)
	e.e.GET("/transactions", func(c echo.Context) error {
		args, err := ec.Transactions(c.Request())
		if err != nil {
			log.Println(err)
		}
		err = c.Render(http.StatusOK, "transactions.html", args)
		if err != nil {
			log.Println(err)
		}
		return err
	}, e.webChecker)
	e.e.GET("/transactionDetail", func(c echo.Context) error {
		args, err := ec.TransactionDetail(c.Request())
		if err != nil {
			log.Println(err)
		}
		err = c.Render(http.StatusOK, "transactionDetail.html", args)
		if err != nil {
			log.Println(err)
		}
		return err
	}, e.webChecker)
	e.e.GET("/formulators", func(c echo.Context) error {
		args, err := ec.Formulators(c.Request())
		if err != nil {
			log.Println(err)
		}
		err = c.Render(http.StatusOK, "formulators.html", args)
		if err != nil {
			log.Println(err)
		}
		return err
	}, e.webChecker)

}

// AddURL add homepage url
func (e *BlockExplorer) AddURL(url string, method string, handler func(c echo.Context) error) {
	switch method {
	case "CONNECT":
		e.e.CONNECT(url, handler, e.webChecker)
	case "DELETE":
		e.e.DELETE(url, handler, e.webChecker)
	case "GET":
		e.e.GET(url, handler, e.webChecker)
	case "HEAD":
		e.e.HEAD(url, handler, e.webChecker)
	case "OPTIONS":
		e.e.OPTIONS(url, handler, e.webChecker)
	case "PATCH":
		e.e.PATCH(url, handler, e.webChecker)
	case "POST":
		e.e.POST(url, handler, e.webChecker)
	case "PUT":
		e.e.PUT(url, handler, e.webChecker)
	case "TRACE":
		e.e.TRACE(url, handler, e.webChecker)
	case "ANY":
		e.e.Any(url, handler, e.webChecker)
	}
}

// StartExplorer is start web server
func (e *BlockExplorer) StartExplorer(port int) {
	if e.initURLFlag != true {
		e.InitURL()
	}
	e.e.Start(":" + strconv.Itoa(port))
}

// AddDataHandler add data handler
func (e *BlockExplorer) AddDataHandler(d DataHandlerPack) {
	e.dataHandlerPacks = append(e.dataHandlerPacks, d)
}

// DataHandlerPack interface of handler
type DataHandlerPack interface {
	DataHandler(c echo.Context) (interface{}, error)
}

func (e *BlockExplorer) dataHandler(c echo.Context) error {
	order := c.Param("order")
	var result interface{}

	switch order {
	case "transactions.data":
		result = e.transactions()
	case "currentChainInfo.data":
		result = e.CurrentChainInfo
	case "lastestBlocks.data":
		result = e.lastestBlocks()
	case "lastestTransactions.data":
		result = e.lastestTransactions()
	case "paginationBlocks.data":
		startStr := c.QueryParam("start")
		result = e.paginationBlocks(startStr)
	case "paginationTxs.data":
		startStr := c.QueryParam("start")
		result = e.paginationTxs(startStr)
	default:
		for _, v := range e.dataHandlerPacks {
			var err error
			result, err = v.DataHandler(c)
			if err == nil {
				break
			}
		}
	}
	return c.JSON(http.StatusOK, result)
}
