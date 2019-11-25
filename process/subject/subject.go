package subject

import (
	"github.com/fletaio/fleta_testnet/process/user"
	"github.com/fletaio/fleta_testnet/core/types"
)

// Subject manages balance of accounts of the chain
type Subject struct {
	*types.ProcessBase
	pid  uint8
	pm   types.ProcessManager
	cn   types.Provider
	user *user.User
}

// NewSubject returns a Subject
func NewSubject(pid uint8) *Subject {
	p := &Subject{
		pid: pid,
	}
	return p
}

// ID returns the id of the process
func (p *Subject) ID() uint8 {
	return p.pid
}

// Name returns the name of the process
func (p *Subject) Name() string {
	return "ecrf.subject"
}

// Version returns the version of the process
func (p *Subject) Version() string {
	return "0.0.1"
}

// Init initializes the process
func (p *Subject) Init(reg *types.Register, pm types.ProcessManager, cn types.Provider) error {
	p.pm = pm
	p.cn = cn

	if vp, err := pm.ProcessByName("ecrf.user"); err != nil {
		return err
	} else if v, is := vp.(*user.User); !is {
		return types.ErrInvalidProcess
	} else {
		p.user = v
	}

	reg.RegisterTransaction(1, &CreateSubject{})
	reg.RegisterTransaction(2, &UpdateSubject{})
	reg.RegisterTransaction(3, &UpdateSubjectPassword{})
	return nil
}

// OnLoadChain called when the chain loaded
func (p *Subject) OnLoadChain(loader types.LoaderWrapper) error {
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *Subject) BeforeExecuteTransactions(ctw *types.ContextWrapper) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *Subject) AfterExecuteTransactions(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (p *Subject) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
