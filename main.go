package main // imports "github.com/fletaio/fleta_testnet"

import (
	"log"

	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
	"github.com/fletaio/fleta_testnet/encoding"
	"github.com/fletaio/fleta_testnet/process/vault"
	"github.com/fletaio/fleta_testnet/service/p2p"
)

// message types
var (
	TransactionMessageType = types.DefineHashedType("p2p.TransactionMessage")
)

func main() {
	reg := types.NewRegister(1)
	reg.RegisterTransaction(1, &vault.Transfer{})

	fc := encoding.Factory("message")
	fc.Register(TransactionMessageType, []*p2p.TransactionMessage{})

	fct := encoding.Factory("transaction")
	t, _ := fct.TypeOf(&vault.Transfer{})

	bs, _ := encoding.Marshal([]*p2p.TransactionMessage{
		&p2p.TransactionMessage{
			ChainID:    1,
			Type:       t,
			Tx:         &vault.Transfer{},
			Signatures: []common.Signature{common.Signature{1, 2, 3}},
		},
	})
	a := []*p2p.TransactionMessage{}
	encoding.Unmarshal(bs, &a)
	log.Println(a[0].Signatures[0])
	return
}
