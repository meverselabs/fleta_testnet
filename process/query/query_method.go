package query

import (
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
)

func (p *Query) IsOpenQuery(loader types.Loader, Addr common.Address, QueryID string) (bool, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	if bs := lw.AccountData(Addr, toQueryKey(QueryID)); len(bs) > 0 {
		return bs[0] == 1, nil
	} else {
		return false, ErrNotExistQuery
	}
}

func (p *Query) addQuery(ctw *types.ContextWrapper, Addr common.Address, QueryID string) {
	ctw.SetAccountData(Addr, toQueryKey(QueryID), []byte{1})
}

func (p *Query) closeQuery(ctw *types.ContextWrapper, Addr common.Address, QueryID string) error {
	if is, err := p.IsOpenQuery(ctw, Addr, QueryID); err != nil {
		return err
	} else if !is {
		return ErrClosedQuery
	} else {
		ctw.SetAccountData(Addr, toQueryKey(QueryID), []byte{0})
		return nil
	}
}
