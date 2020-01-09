package main

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"

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
	"github.com/gorilla/websocket"
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
	InsertTxCount  int
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
	SeedNodeMap[common.MustParsePublicHash("4YjmYcLVvBSmtjh4Z7frRZhWgdEAYTSABCoqqzhKEJa")] = "45.77.147.144:41000"
	SeedNodeMap[common.MustParsePublicHash("27n37VV3ebGWSNH5r9wX3ZhUwzxC2heY34UvXjizLDK")] = "108.61.82.48:41000"
	SeedNodeMap[common.MustParsePublicHash("4GzTnuP7Hky1Dye1AJMLzEXTX2a5kEka5h9AJVvZyTD")] = "107.191.43.224:41000"
	SeedNodeMap[common.MustParsePublicHash("4ew8HQEwwSqeepMDCnwN9PiYg1uvoeZXyudqdQZBCb3")] = "140.82.7.91:41000"
	SeedNodeMap[common.MustParsePublicHash("VbMwA5AwSfn93ks8HMv7vvSx4THuzfeefTWVoANEha")] = "149.28.57.20:41000"
	SeedNodeMap[common.MustParsePublicHash("8eDJ3h8DLW8RSovYUjxmcDi1QNvo7UW64MQxGZ9dnS")] = "45.76.2.218:41000"
	SeedNodeMap[common.MustParsePublicHash("3ZdKaqaCbGSQ5xmAphzVTeEF1eGzX6iU4LLGD2ox2g9")] = "45.63.10.124:41000"
	SeedNodeMap[common.MustParsePublicHash("3UHQyJwSSHHCw29fB5xiGk9W7GNf1DjGC284WhW6jpD")] = "149.28.229.121:41000"
	SeedNodeMap[common.MustParsePublicHash("v3GwqbQehcqNVYbRzDk3TDJ7yJ19DgwoamZnMJZuVg")] = "66.55.159.135:41000"
	SeedNodeMap[common.MustParsePublicHash("3HhrC3gPR951SjnxjnHpfhRSWH1iR3SbCSwtCHvTLuC")] = "208.167.233.43:41000"
	SeedNodeMap[common.MustParsePublicHash("4Ei1HSF3KtDfGrdzHCWfRf4NSTZ2oYCT1CNGFkjV1WB")] = "144.202.0.171:41000"
	SeedNodeMap[common.MustParsePublicHash("3u6v76WAknSq1j86Pfb6p31FsBAJztPdVmY1kkw4k66")] = "208.167.239.236:41000"
	SeedNodeMap[common.MustParsePublicHash("MP6nHXaNjZRXFfSffbRuMDhjsS8YFxEsrtrDAZ9bNW")] = "45.76.6.45:41000"
	SeedNodeMap[common.MustParsePublicHash("4FQ3TVTWQi7TPDerc8nZUBtHyPaNRccA44ushVRWCKW")] = "45.76.0.241:41000"
	SeedNodeMap[common.MustParsePublicHash("3Ue7mXou8FJouGUyn7MtmahGNgevHt7KssNB2E9wRgL")] = "45.77.100.83:41000"
	SeedNodeMap[common.MustParsePublicHash("MZtuTqpsdGLm9QXKaM68sTDwUCyitL7q4L75Vrpwbo")] = "207.148.18.155:41000"
	SeedNodeMap[common.MustParsePublicHash("2fJTp1KMwBqJRqpwGgH5kUCtfBjUBGYgd8oXEA8V9AY")] = "207.246.127.38:41000"
	SeedNodeMap[common.MustParsePublicHash("3yTFnJJqx3wCiK2Edk9f9JwdvdkC4DP4T1y8xYztMkf")] = "45.63.13.183:41000"

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
			Addrs = Addrs[:10000]
		case "7b5c05c6a87f650900dafd05fcbdb184c8d5b5f81cb31fff49f9212b44848340":
			Addrs = Addrs[10000:20000]
		case "e85029d11cdfc8595331bec963977a410fdeca1c36dbd89e2ec7c2985a15ac78":
			Addrs = Addrs[20000:30000]
		case "e2ec6a295d63d9bf312367773efe0b164d55554a61880741b072c87cd66298ae":
			Addrs = Addrs[30000:40000]
		case "bb3f0d6b24dce5d5b4d539a814ba23ff429c1dfacde9a83e72cb4049a38ca113":
			Addrs = Addrs[40000:50000]
		case "f322fa429a627772b76249c96d9e4525eb7c7ab66fc8ff16e7f141c1ddd61b6b":
			Addrs = Addrs[50000:60000]
		case "a3bcc459e90b575d75a64aa7f8a0e45b610057d2132112f9d5876b358d95609b":
			Addrs = Addrs[60000:70000]
		case "0f72009df8bbbf78aed3adbadb31d89410e7a4d4d0b104421b72b5d2e5343577":
			Addrs = Addrs[70000:80000]
		case "a0c7e7ab4bb90e55c4e8d6fde2f7e9c18d9e1a9a8ba8cdf8e48caa2e6003f252":
			Addrs = Addrs[80000:90000]
		case "16e0381a755ea31b5567db0557d173fca57396f54ba734ade9f7a8e420e446b3":
			Addrs = Addrs[90000:100000]
		case "e5db5c29bdfb784f7235f86bfc9ac28e5e6e0507aaacc4b0e1d7db73ee20a1f5":
			Addrs = Addrs[10000:110000]
		case "ea060ebefabb620500080461d438748e967965c210991b8e1a7b7435f96585e1":
			Addrs = Addrs[110000:120000]
		default:
			Addrs = []common.Address{}
		}
		if cfg.InsertTxCount > len(Addrs) {
			cfg.InsertTxCount = len(Addrs)
		}
		Addrs = Addrs[:cfg.InsertTxCount]
	} else {
		Addrs = []common.Address{}
	}

	fr := pof.NewFormulatorNode(&pof.FormulatorConfig{
		Formulator:              common.MustParseAddress(cfg.Formulator),
		MaxTransactionsPerBlock: 7000,
	}, frkey, ndkey, NetAddressMap, SeedNodeMap, cs, cfg.StoreRoot+"/peer")
	if err := fr.Init(); err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("formulator", fr)

	go fr.Run(":" + strconv.Itoa(cfg.Port))

	cm.Wait()
}

func GetStudyMeta(c *websocket.Conn, addr string) (string, error) {
	res, err := DoRequest(c, "study.meta", []interface{}{addr})
	if err != nil {
		return "", err
	} else {
		bs, err := json.MarshalIndent(res, "", "\t")
		if err != nil {
			return "", err
		} else {
			return string(bs), nil
		}
	}
}
