package main

import (
	"encoding/hex"
	"log"
	"strconv"
	"time"

	"github.com/fletaio/fleta_testnet/common/amount"

	"github.com/fletaio/fleta_testnet/cmd/app"
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
	"github.com/fletaio/fleta_testnet/service/p2p"
)

func main() {
	MaxBlocksPerFormulator := uint32(10)
	ChainID := uint8(0x01)
	Symbol := "FLETA"
	Usage := "Mainnet"
	Version := uint16(0x0001)

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
			log.Println(i, pubhash.String())
			NetAddressMap[pubhash] = ":400" + strconv.Itoa(i)
			FrNetAddressMap[pubhash] = "ws://localhost:500" + strconv.Itoa(i)
		}
	}

	for i, obkey := range obkeys {
		back, err := backend.Create("buntdb", "./_test/odata_"+strconv.Itoa(i)+"/context")
		if err != nil {
			panic(err)
		}
		cdb, err := pile.Open("./_test/odata_" + strconv.Itoa(i) + "/chain")
		if err != nil {
			panic(err)
		}
		cdb.SetSyncMode(true)
		st, err := chain.NewStore(back, cdb, ChainID, Symbol, Usage, Version)
		if err != nil {
			panic(err)
		}
		defer st.Close()

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
		if err := st.IterBlockAfterContext(func(b *types.Block) error {
			if err := cn.ConnectBlock(b, nil); err != nil {
				panic(err)
			}
			log.Println(b.Header.Height, "Connect block From local", b.Header.Generator.String(), b.Header.Height)
			return nil
		}); err != nil {
			panic(err)
		}

		ob := pof.NewObserverNode(obkey, NetAddressMap, cs)
		if err := ob.Init(); err != nil {
			panic(err)
		}

		go ob.Run(":400"+strconv.Itoa(i), ":500"+strconv.Itoa(i))
	}

	ndstrs := []string{
		"43fdd20672a54ac9efb8723f85fa7acd8ec5636dfdcd130afa9a1dff6a8f8f04",
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
		back, err := backend.Create("buntdb", "./_test/fdata_"+strconv.Itoa(i)+"/context")
		if err != nil {
			panic(err)
		}
		cdb, err := pile.Open("./_test/fdata_" + strconv.Itoa(i) + "/chain")
		if err != nil {
			panic(err)
		}
		cdb.SetSyncMode(true)
		st, err := chain.NewStore(back, cdb, ChainID, Symbol, Usage, Version)
		if err != nil {
			panic(err)
		}
		defer st.Close()

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
		if err := st.IterBlockAfterContext(func(b *types.Block) error {
			if err := cn.ConnectBlock(b, nil); err != nil {
				panic(err)
			}
			log.Println(b.Header.Height, "Connect block From local", b.Header.Generator.String(), b.Header.Height)
			return nil
		}); err != nil {
			panic(err)
		}

		fr := pof.NewFormulatorNode(&pof.FormulatorConfig{
			Formulator:              common.MustParseAddress(fi.addr),
			MaxTransactionsPerBlock: 10000,
		}, fi.mkey, fi.mkey, FrNetAddressMap, NdNetAddressMap, cs, "./_test/fdata_"+strconv.Itoa(i)+"/peer")
		if err := fr.Init(); err != nil {
			panic(err)
		}
		go func() {
			currentHeight := st.Height()
			for {
				time.Sleep(time.Second * 60)
				if currentHeight == st.Height() {
					panic("make no further progress")
				}
				currentHeight = st.Height()
			}

		}()

		go fr.Run(":600" + strconv.Itoa(i))
	}

	for i, nk := range ndkeys {
		back, err := backend.Create("buntdb", "./_test/ndata_"+strconv.Itoa(i)+"/context")
		if err != nil {
			panic(err)
		}
		cdb, err := pile.Open("./_test/ndata_" + strconv.Itoa(i) + "/chain")
		if err != nil {
			panic(err)
		}
		cdb.SetSyncMode(true)
		st, err := chain.NewStore(back, cdb, ChainID, Symbol, Usage, Version)
		if err != nil {
			panic(err)
		}
		defer st.Close()

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
		if err := st.IterBlockAfterContext(func(b *types.Block) error {
			if err := cn.ConnectBlock(b, nil); err != nil {
				panic(err)
			}
			log.Println(b.Header.Height, "Connect block From local", b.Header.Generator.String(), b.Header.Height)
			return nil
		}); err != nil {
			panic(err)
		}

		nd := p2p.NewNode(nk, NdNetAddressMap, cn, "./_test/ndata_"+strconv.Itoa(i)+"/peer")
		if err := nd.Init(); err != nil {
			panic(err)
		}

		go nd.Run(":601" + strconv.Itoa(i))

		go func() {
			Limit := 200
			if len(Addrs) > Limit {
				Addrs = Addrs[:Limit]
			}

			for _, v := range Addrs {
				go func(Addr common.Address) {
					key, _ := key.NewMemoryKeyFromString("fd1167aad31c104c9fceb5b8a4ffd3e20a272af82176352d3b6ac236d02bafd4")
					//log.Println(Addr.String(), "Start Transaction")

					for {
						if nd.TxPoolSize() > 20000 {
							time.Sleep(100 * time.Millisecond)
							continue
						}
						tx := &vault.Transfer{
							Timestamp_: uint64(time.Now().UnixNano()),
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
				}(v)
			}
		}()
	}

	select {}
}
