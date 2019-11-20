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

	"github.com/fletaio/fleta_testnet/encoding"

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
	"github.com/fletaio/fleta_testnet/core/txpool"
	"github.com/fletaio/fleta_testnet/core/types"
	"github.com/fletaio/fleta_testnet/pof"
	"github.com/fletaio/fleta_testnet/process/admin"
	"github.com/fletaio/fleta_testnet/process/formulator"
	"github.com/fletaio/fleta_testnet/process/gateway"
	"github.com/fletaio/fleta_testnet/process/payment"
	"github.com/fletaio/fleta_testnet/process/vault"
	"github.com/fletaio/fleta_testnet/service/apiserver"
)

// Config is a configuration for the cmd
type Config struct {
	SeedNodeMap    map[string]string
	ObserverKeyMap map[string]string
	GenKeyHex      string
	NodeKeyHex     string
	Formulator     string
	Port           int
	APIPort        int
	StoreRoot      string
	InsertMode     bool
}

func main() {
	var cfg Config
	if err := config.LoadFile("./config.toml", &cfg); err != nil {
		panic(err)
	}
	cfg.NodeKeyHex = cfg.GenKeyHex //TEMP
	if len(cfg.StoreRoot) == 0 {
		cfg.StoreRoot = "./fdata"
	}

	var frkey key.Key
	if bs, err := hex.DecodeString(cfg.GenKeyHex); err != nil {
		panic(err)
	} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
		panic(err)
	} else {
		frkey = Key
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

	NetAddressMap := map[common.PublicHash]string{}
	ObserverKeys := []common.PublicHash{}
	for k, netAddr := range cfg.ObserverKeyMap {
		pubhash, err := common.ParsePublicHash(k)
		if err != nil {
			panic(err)
		}
		NetAddressMap[pubhash] = "ws://" + netAddr
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

	MaxBlocksPerFormulator := uint32(10)
	ChainID := uint8(0x01)
	Name := "FLETA Testnet"
	Version := uint16(0x0001)

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
	cn.MustAddService(ws)
	if err := cn.Init(); err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("chain", cn)

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

	if cfg.InsertMode {
		switch cfg.GenKeyHex {
		case "f9d8e80d688c8b79a0470eaf418d0b6d0adac0648af9481f6d58b69ecebeb82c":
			Addrs = Addrs[:5500]
		case "7b5c05c6a87f650900dafd05fcbdb184c8d5b5f81cb31fff49f9212b44848340":
			Addrs = Addrs[5500:11000]
		case "e85029d11cdfc8595331bec963977a410fdeca1c36dbd89e2ec7c2985a15ac78":
			Addrs = Addrs[11000:16500]
		case "e2ec6a295d63d9bf312367773efe0b164d55554a61880741b072c87cd66298ae":
			Addrs = Addrs[16500:22000]
		case "bb3f0d6b24dce5d5b4d539a814ba23ff429c1dfacde9a83e72cb4049a38ca113":
			Addrs = Addrs[22000:27500]
		case "f322fa429a627772b76249c96d9e4525eb7c7ab66fc8ff16e7f141c1ddd61b6b":
			Addrs = Addrs[27500:33000]
		case "a3bcc459e90b575d75a64aa7f8a0e45b610057d2132112f9d5876b358d95609b":
			Addrs = Addrs[33000:38500]
		case "0f72009df8bbbf78aed3adbadb31d89410e7a4d4d0b104421b72b5d2e5343577":
			Addrs = Addrs[38500:44000]
		case "a0c7e7ab4bb90e55c4e8d6fde2f7e9c18d9e1a9a8ba8cdf8e48caa2e6003f252":
			Addrs = Addrs[44000:49500]
		case "16e0381a755ea31b5567db0557d173fca57396f54ba734ade9f7a8e420e446b3":
			Addrs = Addrs[49500:55000]
		case "e5db5c29bdfb784f7235f86bfc9ac28e5e6e0507aaacc4b0e1d7db73ee20a1f5":
			Addrs = Addrs[55000:60500]
		case "ea060ebefabb620500080461d438748e967965c210991b8e1a7b7435f96585e1":
			Addrs = Addrs[60500:66000]
		default:
			Addrs = []common.Address{}
		}
	} else {
		Addrs = []common.Address{}
	}

	PoolItems := []*txpool.PoolItem{}
	if true {
		key, _ := key.NewMemoryKeyFromString("fd1167aad31c104c9fceb5b8a4ffd3e20a272af82176352d3b6ac236d02bafd4")
		signer := common.NewPublicHash(key.PublicKey())
		fc := encoding.Factory("transaction")
		for _, Addr := range Addrs {
			tx := &vault.TransferUnsafe{
				Timestamp_: uint64(time.Now().UnixNano()),
				From_:      Addr,
				To:         Addr,
				Amount:     amount.NewCoinAmount(1, 0),
			}
			t, err := fc.TypeOf(tx)
			if err != nil {
				panic(err)
			}
			TxHash := chain.HashTransactionByType(ChainID, t, tx)
			sig, err := key.Sign(TxHash)
			if err != nil {
				panic(err)
			}
			item := &txpool.PoolItem{
				TxType:      t,
				TxHash:      TxHash,
				Transaction: tx,
				Signatures:  []common.Signature{sig},
				Signers:     []common.PublicHash{signer},
			}
			PoolItems = append(PoolItems, item)
		}
	}

	fr := pof.NewFormulatorNode(&pof.FormulatorConfig{
		Formulator:              common.MustParseAddress(cfg.Formulator),
		MaxTransactionsPerBlock: 7000,
		PoolItems:               PoolItems,
	}, frkey, ndkey, NetAddressMap, SeedNodeMap, cs, cfg.StoreRoot+"/peer")
	if err := fr.Init(); err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("formulator", fr)

	waitMap := map[common.Address]*chan struct{}{}
	if false {
		go func() {
			for _, Addr := range Addrs {
				waitMap[Addr] = ws.addAddress(Addr)
			}

			for _, v := range Addrs {
				go func(Addr common.Address) {
					for {
						time.Sleep(5 * time.Second)

						Seq := st.Seq(Addr)
						key, _ := key.NewMemoryKeyFromString("fd1167aad31c104c9fceb5b8a4ffd3e20a272af82176352d3b6ac236d02bafd4")
						log.Println(Addr.String(), "Start Transaction", Seq)

						for i := 0; i < 2; i++ {
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
							if err := fr.AddTx(tx, []common.Signature{sig}); err != nil {
								//panic(err)
							}
							time.Sleep(100 * time.Millisecond)
						}

						pCh := waitMap[Addr]

						if pCh == nil {
							log.Println(Addr)
						}

						for range *pCh {
							Seq++
							//log.Println(Addr.String(), "Execute Transaction", Seq)
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
							if err := fr.AddTx(tx, []common.Signature{sig}); err != nil {
								switch err {
								case txpool.ErrExistTransaction:
								case txpool.ErrTooFarSeq:
									Seq--
								}
								time.Sleep(100 * time.Millisecond)
								continue
							}
							time.Sleep(10 * time.Millisecond)
						}
					}
				}(v)
			}
		}()
	}

	go func() {
		for {
			b := <-ws.blockCh
			for i, t := range b.Transactions {
				res := b.TransactionResults[i]
				if res == 1 {
					if tx, is := t.(chain.AccountTransaction); is {
						CreatedAddr := common.NewAddress(b.Header.Height, uint16(i), 0)
						switch tx.(type) {
						case (*vault.IssueAccount):
							log.Println("Created", CreatedAddr.String())
						//case (*vault.Transfer):
						//	log.Println("Transfered", tx.(*vault.Transfer).To)
						default:
							pCh, has := waitMap[tx.From()]
							if has {
								(*pCh) <- struct{}{}
							}
						}
					}
				}
			}
		}
	}()

	go fr.Run(":" + strconv.Itoa(cfg.Port))
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
