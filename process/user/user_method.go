package user

import (
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
)

func (p *User) HasUser(loader types.Loader, Addr common.Address, UserID string) bool {
	lw := types.NewLoaderWrapper(p.pid, loader)

	if bs := lw.AccountData(Addr, toUserKey(UserID)); len(bs) > 0 {
		return true
	} else {
		return false
	}
}

func (p *User) IsUserRole(loader types.Loader, Addr common.Address, UserID string, Roles []string) bool {
	lw := types.NewLoaderWrapper(p.pid, loader)

	bs := lw.AccountData(Addr, toUserKey(UserID))
	if len(bs) == 0 {
		return false
	}
	roleMap := map[string]bool{}
	for _, role := range Roles {
		roleMap[role] = true
	}
	return roleMap[string(bs)]
}

func (p *User) addUser(ctw *types.ContextWrapper, Addr common.Address, UserID string, Role string) {
	ctw.SetAccountData(Addr, toUserKey(UserID), []byte(Role))
}

func (p *User) updateUserRole(ctw *types.ContextWrapper, Addr common.Address, UserID string, Role string) {
	ctw.SetAccountData(Addr, toUserKey(UserID), []byte(Role))
}
