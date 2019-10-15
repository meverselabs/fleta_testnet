package main // imports "github.com/fletaio/fleta_testnet"

import (
	"encoding/hex"
	"log"
	"os"
	"strconv"

	"github.com/fletaio/fleta_testnet/core/types"

	"github.com/fletaio/fleta_testnet/core/pile"

	"github.com/fletaio/fleta_testnet/cmd/app"
	"github.com/fletaio/fleta_testnet/cmd/config"
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/common/key"
	"github.com/fletaio/fleta_testnet/core/backend"
	_ "github.com/fletaio/fleta_testnet/core/backend/badger_driver"
	_ "github.com/fletaio/fleta_testnet/core/backend/buntdb_driver"
	"github.com/fletaio/fleta_testnet/core/chain"
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
	SeedNodeMap     map[string]string
	ObserverKeyMap  map[string]string
	KeyHex          string
	NodeKeyHex      string
	GatewayKeyHex   string
	ObserverKeys    []string
	Port            int
	APIPort         int
	WebPort         int
	WebPortExplorer int
	StoreRoot       string
	BackendVersion  int
	Backend         string
	ForceRecover    bool
	Domain          string
}

func main() {
	if err := test(); err != nil {
		panic(err)
	}
}

func test() error {
	var cfg Config
	if err := config.LoadFile("./config.toml", &cfg); err != nil {
		panic(err)
	}

	os.RemoveAll("./_test")
	defer os.RemoveAll("./_test")

	obstrs := []string{
		"73c80ff5f0fd053ab12fdecbedf9693620470be1f9d3f3fdee28a2dc2e200803",
		"289174d46cac3985a81a77e157dc441088344432ff3a5628478fd0f631aaae76",
		"adecc23d0cd3c361cf7e391458fd13588e84e904aa99133b54eec47f4ec0d1da",
		"f686cfb2f1e77f1cbea78f59e85fb98fa8df8187c8b8186e2bde6324195206f1",
		"ff7bca3d2b0481e0e6f2af4fb541b01ed5d163334815e645511869dcff70d4aa",
	}
	obkeys := make([]key.Key, 0, len(obstrs))
	NetAddressMap := map[common.PublicHash]string{}
	FrNetAddressMap := map[common.PublicHash]string{}
	ObserverKeys := make([]common.PublicHash, 0, len(obstrs))
	for i, v := range obstrs {
		if bs, err := hex.DecodeString(v); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
			panic(err)
		} else {
			obkeys = append(obkeys, Key)
			pubhash := common.NewPublicHash(Key.PublicKey())
			ObserverKeys = append(ObserverKeys, pubhash)
			NetAddressMap[pubhash] = ":400" + strconv.Itoa(i)
			FrNetAddressMap[pubhash] = "ws://localhost:500" + strconv.Itoa(i)
		}
	}

	MaxBlocksPerFormulator := uint32(10)
	ChainID := uint8(0x01)
	Name := "FLETA Testnet"
	Version := uint16(0x0001)

	for i, obkey := range obkeys {
		var back backend.StoreBackend
		var cdb *pile.DB
		switch cfg.BackendVersion {
		case 0:
			contextDB, err := backend.Create("badger", cfg.StoreRoot)
			if err != nil {
				panic(err)
			}
			back = contextDB
		case 1:
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
		if err := cn.Init(); err != nil {
			panic(err)
		}

		if err := st.IterBlockAfterContext(func(b *types.Block) error {
			if err := cn.ConnectBlock(b); err != nil {
				return err
			}
			return nil
		}); err != nil {
			if err == chain.ErrStoreClosed {
				return nil
			}
			panic(err)
		}

		ob := pof.NewObserverNode(obkey, NetAddressMap, cs)
		if err := ob.Init(); err != nil {
			panic(err)
		}

		go ob.Run(":400"+strconv.Itoa(i), ":500"+strconv.Itoa(i))
		log.Println(`go ob.Run(":400"+strconv.Itoa(i), ":500"+strconv.Itoa(i))`)
	}

	ndstrs := []string{
		"44bc87d266348e96f68ecce817ad358bfae0a796653084bdf4a079c31d7381c7",
	}
	NdNetAddressMap := map[common.PublicHash]string{}
	ndkeys := make([]key.Key, 0, len(ndstrs))
	for i, v := range ndstrs {
		if bs, err := hex.DecodeString(v); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
			panic(err)
		} else {
			ndkeys = append(ndkeys, Key)
			pubhash := common.NewPublicHash(Key.PublicKey())
			NdNetAddressMap[pubhash] = ":601" + strconv.Itoa(i)
		}
	}

	type frinfo struct {
		key  string
		addr string
		mkey *key.MemoryKey
	}

	fris := []*frinfo{}
	fris = append(fris, &frinfo{"f9d8e80d688c8b79a0470eaf418d0b6d0adac0648af9481f6d58b69ecebeb82c", "5CyLcFhpyN", nil})

	for i, fi := range fris {
		if bs, err := hex.DecodeString(fi.key); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
			panic(err)
		} else {
			fi.mkey = Key
		}
		var back backend.StoreBackend
		var cdb *pile.DB
		switch cfg.BackendVersion {
		case 0:
			contextDB, err := backend.Create("badger", cfg.StoreRoot)
			if err != nil {
				panic(err)
			}
			back = contextDB
		case 1:
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
		if err := cn.Init(); err != nil {
			panic(err)
		}

		if err := st.IterBlockAfterContext(func(b *types.Block) error {
			if err := cn.ConnectBlock(b); err != nil {
				return err
			}
			return nil
		}); err != nil {
			if err == chain.ErrStoreClosed {
				return nil
			}
			panic(err)
		}

		fr := pof.NewFormulatorNode(&pof.FormulatorConfig{
			Formulator:              common.MustParseAddress(fi.addr),
			MaxTransactionsPerBlock: 10000,
		}, fi.mkey, fi.mkey, FrNetAddressMap, NdNetAddressMap, cs, "./_test/fdata_"+strconv.Itoa(i)+"/peer")
		if err := fr.Init(); err != nil {
			panic(err)
		}
		go fr.Run(":600" + strconv.Itoa(i))
		log.Println(`go fr.Run(":600" + strconv.Itoa(i))`)
	}

	select {}
	return nil
}
