package chain

import (
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/common/hash"
)

// SignerProvider provides signers of the hash
type SignerProvider interface {
	GetSigners(TxHash hash.Hash256) []common.PublicHash
	UnsafeGetSigners(TxHash hash.Hash256) []common.PublicHash
	Lock()
	Unlock()
}
