package user

import (
	"github.com/fletaio/fleta_testnet/process/study"
	"github.com/fletaio/fleta_testnet/core/types"
)

// User manages balance of accounts of the chain
type User struct {
	*types.ProcessBase
	pid   uint8
	pm    types.ProcessManager
	cn    types.Provider
	study *study.Study
}

// NewUser returns a User
func NewUser(pid uint8) *User {
	p := &User{
		pid: pid,
	}
	return p
}

// ID returns the id of the process
func (p *User) ID() uint8 {
	return p.pid
}

// Name returns the name of the process
func (p *User) Name() string {
	return "ecrf.user"
}

// Version returns the version of the process
func (p *User) Version() string {
	return "0.0.1"
}

// Init initializes the process
func (p *User) Init(reg *types.Register, pm types.ProcessManager, cn types.Provider) error {
	p.pm = pm
	p.cn = cn

	if vp, err := pm.ProcessByName("ecrf.study"); err != nil {
		return err
	} else if v, is := vp.(*study.Study); !is {
		return types.ErrInvalidProcess
	} else {
		p.study = v
	}

	reg.RegisterTransaction(1, &CreateUser{})
	reg.RegisterTransaction(2, &UpdateUser{})
	return nil
}

// OnLoadChain called when the chain loaded
func (p *User) OnLoadChain(loader types.LoaderWrapper) error {
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *User) BeforeExecuteTransactions(ctw *types.ContextWrapper) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *User) AfterExecuteTransactions(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (p *User) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
