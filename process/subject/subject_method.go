package subject

import (
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
)

func (p *Subject) HasSubject(loader types.Loader, Addr common.Address, SubjectID string) bool {
	lw := types.NewLoaderWrapper(p.pid, loader)

	if bs := lw.AccountData(Addr, toSubjectKey(SubjectID)); len(bs) > 0 {
		return true
	} else {
		return false
	}
}

func (p *Subject) addSubject(ctw *types.ContextWrapper, Addr common.Address, SubjectID string) {
	ctw.SetAccountData(Addr, toSubjectKey(SubjectID), []byte{1})
}

func (p *Subject) removeSubject(ctw *types.ContextWrapper, Addr common.Address, SubjectID string) {
	ctw.SetAccountData(Addr, toSubjectKey(SubjectID), nil)
}
