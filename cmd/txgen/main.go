package main

import (
	"encoding/hex"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/fletaio/fleta_testnet/cmd/app"
	"github.com/fletaio/fleta_testnet/cmd/closer"
	"github.com/fletaio/fleta_testnet/cmd/config"
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/common/amount"
	"github.com/fletaio/fleta_testnet/common/key"
	"github.com/fletaio/fleta_testnet/core/backend"
	_ "github.com/fletaio/fleta_testnet/core/backend/badger_driver"
	_ "github.com/fletaio/fleta_testnet/core/backend/buntdb_driver"
	"github.com/fletaio/fleta_testnet/core/chain"
	"github.com/fletaio/fleta_testnet/core/pile"
	"github.com/fletaio/fleta_testnet/core/types"
	"github.com/fletaio/fleta_testnet/pof"
	"github.com/fletaio/fleta_testnet/process/admin"
	"github.com/fletaio/fleta_testnet/process/formulator"
	"github.com/fletaio/fleta_testnet/process/gateway"
	"github.com/fletaio/fleta_testnet/process/payment"
	"github.com/fletaio/fleta_testnet/process/vault"
	"github.com/fletaio/fleta_testnet/service/apiserver"
	"github.com/fletaio/fleta_testnet/service/p2p"
	"github.com/fletaio/testnet_explorer/explorerservice"
)

// Config is a configuration for the cmd
type Config struct {
	SeedNodeMap  map[string]string
	NodeKeyHex   string
	ObserverKeys []string
	Port         int
	APIPort      int
	StoreRoot    string
	CreateMode   bool
}

func main() {
	var cfg Config
	if err := config.LoadFile("./config.toml", &cfg); err != nil {
		panic(err)
	}
	if len(cfg.StoreRoot) == 0 {
		cfg.StoreRoot = "./ndata"
	}

	var ndkey key.Key
	if len(cfg.NodeKeyHex) > 0 {
		if bs, err := hex.DecodeString(cfg.NodeKeyHex); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
			panic(err)
		} else {
			ndkey = Key
		}
	} else {
		if bs, err := ioutil.ReadFile("./ndkey.key"); err != nil {
			k, err := key.NewMemoryKey()
			if err != nil {
				panic(err)
			}

			fs, err := os.Create("./ndkey.key")
			if err != nil {
				panic(err)
			}
			fs.Write(k.Bytes())
			fs.Close()
			ndkey = k
		} else {
			if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
				panic(err)
			} else {
				ndkey = Key
			}
		}
	}

	ObserverKeys := []common.PublicHash{}
	for _, k := range cfg.ObserverKeys {
		pubhash, err := common.ParsePublicHash(k)
		if err != nil {
			panic(err)
		}
		ObserverKeys = append(ObserverKeys, pubhash)
	}
	SeedNodeMap := map[common.PublicHash]string{}
	for k, netAddr := range cfg.SeedNodeMap {
		pubhash, err := common.ParsePublicHash(k)
		if err != nil {
			panic(err)
		}
		SeedNodeMap[pubhash] = netAddr
	}

	MaxBlocksPerFormulator := uint32(10)
	ChainID := uint8(0x01)
	Name := "FLETA Testnet"
	Version := uint16(0x0001)

	cm := closer.NewManager()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		cm.CloseAll()
	}()
	defer cm.CloseAll()

	var back backend.StoreBackend
	var cdb *pile.DB
	if true {
		contextDB, err := backend.Create("buntdb", cfg.StoreRoot+"/context")
		if err != nil {
			panic(err)
		}
		chainDB, err := pile.Open(cfg.StoreRoot + "/chain")
		if err != nil {
			panic(err)
		}
		chainDB.SetSyncMode(true)
		back = contextDB
		cdb = chainDB
	}
	st, err := chain.NewStore(back, cdb, ChainID, Name, Version)
	if err != nil {
		panic(err)
	}
	cm.Add("store", st)

	if st.Height() > 0 {
		if _, err := cdb.GetData(st.Height(), 0); err != nil {
			panic(err)
		}
	}

	cs := pof.NewConsensus(MaxBlocksPerFormulator, ObserverKeys)
	app := app.NewFletaApp()
	cn := chain.NewChain(cs, app, st)
	cn.MustAddProcess(admin.NewAdmin(1))
	cn.MustAddProcess(vault.NewVault(2))
	cn.MustAddProcess(formulator.NewFormulator(3))
	cn.MustAddProcess(gateway.NewGateway(4))
	cn.MustAddProcess(payment.NewPayment(5))
	as := apiserver.NewAPIServer()
	cn.MustAddService(as)
	ws := NewWatcher()
	if cfg.CreateMode {
		cn.MustAddService(ws)
	}
	e, err := explorerservice.NewBlockExplorer("_explorer", cs, cfg.WebPort)
	if err != nil {
		panic(err)
	}
	cn.MustAddService(e)
	if err := cn.Init(); err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("chain", cn)
	SeedNodeMap[common.MustParsePublicHash("4YjmYcLVvBSmtjh4Z7frRZhWgdEAYTSABCoqqzhKEJa")] = "199.247.2.136:41000"
	SeedNodeMap[common.MustParsePublicHash("27n37VV3ebGWSNH5r9wX3ZhUwzxC2heY34UvXjizLDK")] = "199.247.0.226:41000"
	SeedNodeMap[common.MustParsePublicHash("4ew8HQEwwSqeepMDCnwN9PiYg1uvoeZXyudqdQZBCb3")] = "149.28.240.38:41000"
	SeedNodeMap[common.MustParsePublicHash("VbMwA5AwSfn93ks8HMv7vvSx4THuzfeefTWVoANEha")] = "149.28.249.207:41000"
	SeedNodeMap[common.MustParsePublicHash("3ZdKaqaCbGSQ5xmAphzVTeEF1eGzX6iU4LLGD2ox2g9")] = "45.77.59.252:41000"
	SeedNodeMap[common.MustParsePublicHash("3UHQyJwSSHHCw29fB5xiGk9W7GNf1DjGC284WhW6jpD")] = "108.61.172.231:41000"
	SeedNodeMap[common.MustParsePublicHash("3HhrC3gPR951SjnxjnHpfhRSWH1iR3SbCSwtCHvTLuC")] = "45.32.173.38:41000"
	SeedNodeMap[common.MustParsePublicHash("4Ei1HSF3KtDfGrdzHCWfRf4NSTZ2oYCT1CNGFkjV1WB")] = "149.28.106.61:41000"
	SeedNodeMap[common.MustParsePublicHash("CV3cNk8UZxJcsLYjSgMdKuMf7VbDnbHXyqvb2rSE4y")] = "140.82.63.172:41000"
	SeedNodeMap[common.MustParsePublicHash("38qmoMNCuBht1ihjCKVV5nTWvfiDU7NBNeeHWhB7eT7")] = "45.77.76.27:41000"
	SeedNodeMap[common.MustParsePublicHash("EMLGsnW7RvSWTtmArG7aJuASvR7iFwg7uy59FmAwT2")] = "140.82.52.163:41000"
	SeedNodeMap[common.MustParsePublicHash("3Uo6d6w1Xrebq1j42Nm2TguHn42R5MgZTMHBwP4HfrX")] = "95.179.209.187:41000"
	SeedNodeMap[common.MustParsePublicHash("MP6nHXaNjZRXFfSffbRuMDhjsS8YFxEsrtrDAZ9bNW")] = "80.240.18.208:41000"
	SeedNodeMap[common.MustParsePublicHash("4FQ3TVTWQi7TPDerc8nZUBtHyPaNRccA44ushVRWCKW")] = "144.202.69.204:41000"
	SeedNodeMap[common.MustParsePublicHash("3Ue7mXou8FJouGUyn7MtmahGNgevHt7KssNB2E9wRgL")] = "78.141.196.120:41000"
	SeedNodeMap[common.MustParsePublicHash("MZtuTqpsdGLm9QXKaM68sTDwUCyitL7q4L75Vrpwbo")] = "207.246.69.195:41000"
	SeedNodeMap[common.MustParsePublicHash("2G7uZMucLN3BQYZsjhGhE8cXiJNMGccexqdb4kxHeq")] = "45.63.1.207:41000"
	SeedNodeMap[common.MustParsePublicHash("3BrmvxpBRVifN5ddEo2FDw6jPZYZwEYeMbtKfKScHbs")] = "140.82.55.177:41000"

	if err := st.IterBlockAfterContext(func(b *types.Block) error {
		if cm.IsClosed() {
			return chain.ErrStoreClosed
		}
		if err := cn.ConnectBlock(b, nil); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if err == chain.ErrStoreClosed {
			return
		}
		panic(err)
	}

	nd := p2p.NewNode(ndkey, SeedNodeMap, cn, cfg.StoreRoot+"/peer")
	if err := nd.Init(); err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("node", nd)

	if cfg.CreateMode {
		waitMap := map[common.Address]*chan struct{}{}
		go func() {
			switch cfg.NodeKeyHex {
			case "f07f3de26238cb57776556c67368665a53a969efeddf582028ae0c2344261feb":
				Addrs = Addrs[:12000]
			case "2e56f231189d41397a844232250276992691b9102aa53efd3a315ec2abf76094":
				Addrs = Addrs[12000:24000]
			case "74a4bb065b9553e18c5f6aab54bcb07db58f2950b09d3be024e20318512d97bb":
				Addrs = Addrs[24000:36000]
			case "f7b6a6291165b7d4cea6d16b911b6f2ba024aac6f160b230b9a04de876f3b045":
				Addrs = Addrs[36000:48000]
			case "4bc61ab268197c465d471d67618a3bf385651a93ed5011b8db08f5dfdca43c1d":
				Addrs = Addrs[48000:60000]
			case "bfe9f217f31f52a8e3e975c415d297ff201a4c4abfbcb921eb9013b0c21397f4":
				Addrs = Addrs[60000:72000]
			default:
				Addrs = []common.Address{}
			}

			for _, Addr := range Addrs {
				waitMap[Addr] = ws.addAddress(Addr)
			}

			for _, v := range Addrs {
				go func(Addr common.Address) {
					time.Sleep(5 * time.Second)

					Seq := st.Seq(Addr)
					key, _ := key.NewMemoryKeyFromString("fd1167aad31c104c9fceb5b8a4ffd3e20a272af82176352d3b6ac236d02bafd4")
					log.Println(Addr.String(), "Start Transaction", Seq)

					for i := 0; i < 1; i++ {
						Seq++
						tx := &vault.Transfer{
							Timestamp_: uint64(time.Now().UnixNano()),
							Seq_:       Seq,
							From_:      Addr,
							To:         Addr,
							Amount:     amount.NewCoinAmount(1, 0),
						}
						sig, err := key.Sign(chain.HashTransaction(ChainID, tx))
						if err != nil {
							panic(err)
						}
						if err := nd.AddTx(tx, []common.Signature{sig}); err != nil {
							panic(err)
						}
						time.Sleep(100 * time.Millisecond)
					}

					pCh := waitMap[Addr]

					if pCh == nil {
						log.Println(Addr)
					}

					ticker := time.NewTicker(9000 * time.Millisecond)
					for {
						select {
						case <-ticker.C:
							Seq++
							tx := &vault.Transfer{
								Timestamp_: uint64(time.Now().UnixNano()),
								Seq_:       Seq,
								From_:      Addr,
								To:         Addr,
								Amount:     amount.NewCoinAmount(1, 0),
							}
							sig, err := key.Sign(chain.HashTransaction(ChainID, tx))
							if err != nil {
								panic(err)
							}
							if err := nd.AddTx(tx, []common.Signature{sig}); err != nil {
							}
						case <-*pCh:
							tx := &vault.Transfer{
								Timestamp_: uint64(time.Now().UnixNano()),
								Seq_:       st.Seq(Addr) + 1,
								From_:      Addr,
								To:         Addr,
								Amount:     amount.NewCoinAmount(1, 0),
							}
							sig, err := key.Sign(chain.HashTransaction(ChainID, tx))
							if err != nil {
								panic(err)
							}
							if err := nd.AddTx(tx, []common.Signature{sig}); err != nil {
							}
						}
					}
				}(v)
			}
		}()

		go func() {
			for {
				b := <-ws.blockCh
				for _, t := range b.Transactions {
					if tx, is := t.(*vault.Transfer); is {
						pCh, has := waitMap[tx.From()]
						if has {
							(*pCh) <- struct{}{}
						}
					}
				}
			}
		}()
	}

	go nd.Run(":" + strconv.Itoa(cfg.Port))
	go as.Run(":" + strconv.Itoa(cfg.APIPort))

	cm.Wait()
}

// Watcher provides json rpc and web service for the chain
type Watcher struct {
	sync.Mutex
	types.ServiceBase
	waitMap map[common.Address]*chan struct{}
	blockCh chan *types.Block
}

// NewWatcher returns a Watcher
func NewWatcher() *Watcher {
	s := &Watcher{
		waitMap: map[common.Address]*chan struct{}{},
		blockCh: make(chan *types.Block, 1000),
	}
	return s
}

// Name returns the name of the service
func (s *Watcher) Name() string {
	return "fleta.watcher"
}

// Init called when initialize service
func (s *Watcher) Init(pm types.ProcessManager, cn types.Provider) error {
	return nil
}

// OnLoadChain called when the chain loaded
func (s *Watcher) OnLoadChain(loader types.Loader) error {
	return nil
}

func (s *Watcher) addAddress(addr common.Address) *chan struct{} {
	ch := make(chan struct{})
	s.waitMap[addr] = &ch
	return &ch
}

// OnBlockConnected called when a block is connected to the chain
func (s *Watcher) OnBlockConnected(b *types.Block, events []types.Event, loader types.Loader) {
	s.blockCh <- b
}
