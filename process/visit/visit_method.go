package visit

import (
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
)

func (p *Visit) HasVisit(loader types.Loader, Addr common.Address, VisitID string) bool {
	lw := types.NewLoaderWrapper(p.pid, loader)

	if bs := lw.AccountData(Addr, toVisitKey(VisitID)); len(bs) > 0 {
		return true
	} else {
		return false
	}
}

func (p *Visit) addVisit(ctw *types.ContextWrapper, Addr common.Address, VisitID string) {
	ctw.SetAccountData(Addr, toVisitKey(VisitID), []byte{1})
}

func (p *Visit) removeVisit(ctw *types.ContextWrapper, Addr common.Address, VisitID string) {
	ctw.SetAccountData(Addr, toVisitKey(VisitID), nil)
}
