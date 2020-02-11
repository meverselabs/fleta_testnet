package main

import (
	"encoding/hex"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/fletaio/fleta_testnet/common/amount"

	"github.com/fletaio/fleta_testnet/cmd/app"
	"github.com/fletaio/fleta_testnet/cmd/closer"
	"github.com/fletaio/fleta_testnet/cmd/config"
	"github.com/fletaio/fleta_testnet/common"
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
	"github.com/fletaio/fleta_testnet/service/explorerservice"
	"github.com/fletaio/fleta_testnet/service/p2p"
)

// Config is a configuration for the cmd
type Config struct {
	SeedNodeMap  map[string]string
	NodeKeyHex   string
	ObserverKeys []string
	Port         int
	APIPort      int
	WebPort      int
	StoreRoot    string
	CreateMode   bool
	CustomText   string
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
	Symbol := "FLETA"
	Usage := "Mainnet"
	Version := uint16(0x0001)

	back, err := backend.Create("buntdb", cfg.StoreRoot+"/context")
	if err != nil {
		panic(err)
	}
	cdb, err := pile.Open(cfg.StoreRoot + "/chain")
	if err != nil {
		panic(err)
	}
	cdb.SetSyncMode(true)
	st, err := chain.NewStore(back, cdb, ChainID, Symbol, Usage, Version)
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
	if cfg.NodeKeyHex == "f07f3de26238cb57776556c67368665a53a969efeddf582028ae0c2344261feb" {
		e, err := explorerservice.NewBlockExplorer("_explorer", cs, cfg.WebPort)
		if err != nil {
			panic(err)
		}
		cn.MustAddService(e)
	}
	if err := cn.Init(); err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("chain", cn)
	SeedNodeMap[common.MustParsePublicHash("4YjmYcLVvBSmtjh4Z7frRZhWgdEAYTSABCoqqzhKEJa")] = "217.69.5.228:41000"
	SeedNodeMap[common.MustParsePublicHash("27n37VV3ebGWSNH5r9wX3ZhUwzxC2heY34UvXjizLDK")] = "95.179.217.127:41000"
	SeedNodeMap[common.MustParsePublicHash("4GzTnuP7Hky1Dye1AJMLzEXTX2a5kEka5h9AJVvZyTD")] = "45.32.174.70:41000"
	SeedNodeMap[common.MustParsePublicHash("4ew8HQEwwSqeepMDCnwN9PiYg1uvoeZXyudqdQZBCb3")] = "149.28.105.98:41000"
	SeedNodeMap[common.MustParsePublicHash("VbMwA5AwSfn93ks8HMv7vvSx4THuzfeefTWVoANEha")] = "207.148.27.123:41000"
	SeedNodeMap[common.MustParsePublicHash("8eDJ3h8DLW8RSovYUjxmcDi1QNvo7UW64MQxGZ9dnS")] = "140.82.6.245:41000"
	SeedNodeMap[common.MustParsePublicHash("3ZdKaqaCbGSQ5xmAphzVTeEF1eGzX6iU4LLGD2ox2g9")] = "45.76.128.131:41000"
	SeedNodeMap[common.MustParsePublicHash("3UHQyJwSSHHCw29fB5xiGk9W7GNf1DjGC284WhW6jpD")] = "104.238.187.225:41000"
	SeedNodeMap[common.MustParsePublicHash("v3GwqbQehcqNVYbRzDk3TDJ7yJ19DgwoamZnMJZuVg")] = "144.202.66.83:41000"
	SeedNodeMap[common.MustParsePublicHash("3HhrC3gPR951SjnxjnHpfhRSWH1iR3SbCSwtCHvTLuC")] = "207.148.1.215:41000"
	SeedNodeMap[common.MustParsePublicHash("4Ei1HSF3KtDfGrdzHCWfRf4NSTZ2oYCT1CNGFkjV1WB")] = "209.250.238.6:41000"
	SeedNodeMap[common.MustParsePublicHash("3u6v76WAknSq1j86Pfb6p31FsBAJztPdVmY1kkw4k66")] = "136.244.86.130:41000"
	SeedNodeMap[common.MustParsePublicHash("MP6nHXaNjZRXFfSffbRuMDhjsS8YFxEsrtrDAZ9bNW")] = "217.69.11.84:41000"
	SeedNodeMap[common.MustParsePublicHash("4FQ3TVTWQi7TPDerc8nZUBtHyPaNRccA44ushVRWCKW")] = "207.246.76.244:41000"
	SeedNodeMap[common.MustParsePublicHash("3Ue7mXou8FJouGUyn7MtmahGNgevHt7KssNB2E9wRgL")] = "207.148.25.159:41000"
	SeedNodeMap[common.MustParsePublicHash("MZtuTqpsdGLm9QXKaM68sTDwUCyitL7q4L75Vrpwbo")] = "45.77.226.165:41000"
	SeedNodeMap[common.MustParsePublicHash("2fJTp1KMwBqJRqpwGgH5kUCtfBjUBGYgd8oXEA8V9AY")] = "144.202.71.141:41000"
	SeedNodeMap[common.MustParsePublicHash("3yTFnJJqx3wCiK2Edk9f9JwdvdkC4DP4T1y8xYztMkf")] = "80.240.24.126:41000"

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
		go func() {
			/*
				switch cfg.NodeKeyHex {
				case "f07f3de26238cb57776556c67368665a53a969efeddf582028ae0c2344261feb":
					Addrs = Addrs[:16000]
				case "2e56f231189d41397a844232250276992691b9102aa53efd3a315ec2abf76094":
					Addrs = Addrs[16000:32000]
				case "74a4bb065b9553e18c5f6aab54bcb07db58f2950b09d3be024e20318512d97bb":
					Addrs = Addrs[32000:48000]
				case "f7b6a6291165b7d4cea6d16b911b6f2ba024aac6f160b230b9a04de876f3b045":
					Addrs = Addrs[48000:64000]
				case "313b356ea5411567c7237ecf257ff2335cb43a5263712fbbcf8e31ca6731e311":
					Addrs = Addrs[64000:80000]
				default:
					Addrs = []common.Address{}
				}
			*/
			switch cfg.NodeKeyHex {
			case "f07f3de26238cb57776556c67368665a53a969efeddf582028ae0c2344261feb":
				Addrs = Addrs[:16000]
			case "2e56f231189d41397a844232250276992691b9102aa53efd3a315ec2abf76094":
				Addrs = Addrs[16000:32000]
			case "74a4bb065b9553e18c5f6aab54bcb07db58f2950b09d3be024e20318512d97bb":
				Addrs = Addrs[32000:48000]
			case "f7b6a6291165b7d4cea6d16b911b6f2ba024aac6f160b230b9a04de876f3b045":
				Addrs = Addrs[48000:64000]
			case "313b356ea5411567c7237ecf257ff2335cb43a5263712fbbcf8e31ca6731e311":
				Addrs = Addrs[64000:96000]
			default:
				Addrs = []common.Address{}
			}
			//Limit := 200
			Limit := 380
			if len(Addrs) > Limit {
				Addrs = Addrs[:Limit]
			}

			for _, v := range Addrs {
				go func(Addr common.Address) {
					key, _ := key.NewMemoryKeyFromString("fd1167aad31c104c9fceb5b8a4ffd3e20a272af82176352d3b6ac236d02bafd4")
					log.Println(Addr.String(), "Start Transaction")

					for {
						/*
							if nd.TxPoolSize() > 60000 {
								time.Sleep(100 * time.Millisecond)
								continue
							}
						*/
						tx := &vault.Transfer{
							Timestamp_: uint64(time.Now().UnixNano()),
							From_:      Addr,
							To:         Addr,
							Amount:     amount.NewCoinAmount(1, 0),
						}
						sig, err := key.Sign(types.HashTransaction(ChainID, tx))
						if err != nil {
							panic(err)
						}
						if err := nd.AddTx(tx, []common.Signature{sig}); err != nil {
							if err != types.ErrInvalidTransactionTimeSlot {
								panic(err)
							}
						}
						time.Sleep(250 * time.Millisecond)
					}
				}(v)
			}
		}()
	}

	go nd.Run(":" + strconv.Itoa(cfg.Port))

	cm.Wait()
}
