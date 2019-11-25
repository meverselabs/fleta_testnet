package explorer

import (
	"log"
	"strconv"
	"sync"

	"github.com/fletaio/ecrf/process/study"

	lediscfg "github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"

	"github.com/fletaio/ecrf/pof"
	"github.com/fletaio/ecrf/service/ecrf"
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/p2p"
	"github.com/fletaio/webserver"
)

type Explorer struct {
	sync.Mutex
	pm       types.ProcessManager
	cn       types.Provider
	cs       *pof.Consensus
	db       *ledis.DB
	nd       *p2p.Node
	web      *webserver.WebServer
	ecrf     *ecrf.ECRF
	path     string
	port     int
	infoLock sync.Mutex
}

func NewExplorer(ecrf *ecrf.ECRF, path string, port int) *Explorer {
	cfg := lediscfg.NewConfigDefault()
	cfg.DataDir = path
	l, err := ledis.Open(cfg)
	if err != nil {
		panic(err)
	}

	db, err := l.Select(0)
	if err != nil {
		panic(err)
	}

	return &Explorer{
		db:   db,
		port: port,
		ecrf: ecrf,
	}
}

func (ex *Explorer) Name() string {
	return "ExplorerService"
}

func (ex *Explorer) InitFromStore(st *chain.Store) error {
	if !ex.isInitDB() {
		if err := ex.setTotalSupply(amount.NewCoinAmount(2000000000, 0)); err != nil {
			return err
		}
		accs, err := st.Accounts()
		if err != nil {
			return err
		}
		for _, a := range accs {
			if err := ex.setAccountName(a.Address(), a.Name()); err != nil {
				return err
			}
			switch acc := a.(type) {
			case *study.StudyAccount:
				//ex.addStudy(acc.Address())
				log.Println(acc)
			case *study.SiteAccount:
				//ex.addSite(acc.Address())
			}
		}
		ex.setInitDB()
	}
	return nil
}

func (ex *Explorer) Init(pm types.ProcessManager, cn types.Provider) error {
	ex.pm = pm
	ex.cn = cn

	if ex.Height() < cn.Height() {
		for i := ex.Height() + 1; i <= cn.Height(); i++ {
			b, err := cn.Block(i)
			if err != nil {
				panic(err)
			}
			es, err := cn.Events(i, i)
			if err != nil {
				panic(err)
			}
			ex.UpdateState(b, es)
			if i%1000 == 0 {
				log.Println("explorer block process " + strconv.Itoa(int(i)))
			}
		}
	}

	return nil
}

func (ex *Explorer) SetConsensus(Consensus *pof.Consensus) {
	ex.cs = Consensus
}

func (ex *Explorer) OnLoadChain(loader types.Loader) error {
	go ex.RunWeb()
	return nil
}

func (ex *Explorer) OnBlockConnected(b *types.Block, events []types.Event, loader types.Loader) {
	ex.UpdateState(b, events)
}

func (ex *Explorer) UpdateState(b *types.Block, events []types.Event) error {
	if b.Header.Height != ex.Height()+1 {
		return ErrInvalidHeight
	}

	if err := ex.setHeight(b.Header.Height); err != nil {
		log.Println(err)
		panic(err)
	}
	if err := ex.setBlockHash(b.Header.Height, encoding.Hash(b.Header)); err != nil {
		log.Println(err)
		panic(err)
	}
	if err := ex.addFormulatorBlock(&b.Header); err != nil {
		log.Println(err)
		panic(err)
	}

	for i, t := range b.Transactions {
		TXID := types.TransactionID(b.Header.Height, uint16(i))

		res := b.TransactionResults[i]
		if at, is := t.(chain.AccountTransaction); is {
			if err := ex.addTransaction(TXID, b.TransactionTypes[i], at, res); err != nil {
				log.Println(err)
				panic(err)
			}
		}

		CreatedAddr := common.NewAddress(b.Header.Height, uint16(i), 0)
		//TxHash := chain.HashTransactionByType(ex.cn.ChainID(), b.TransactionTypes[i], t)

		if res == 1 {
			switch tx := t.(type) {
			default:
				log.Println(CreatedAddr, tx)
			}
		}
	}
	return nil
}
