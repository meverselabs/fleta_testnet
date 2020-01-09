package vault

import (
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/common/amount"
	"github.com/fletaio/fleta_testnet/core/types"
)

type FeeTransaction interface {
	From() common.Address
	Fee(p types.Process, lw types.LoaderWrapper) *amount.Amount
}
