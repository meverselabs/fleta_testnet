package visit

import (
	"github.com/fletaio/fleta_testnet/process/subject"
	"github.com/fletaio/fleta_testnet/process/user"
	"github.com/fletaio/fleta_testnet/core/types"
)

// Visit manages balance of accounts of the chain
type Visit struct {
	*types.ProcessBase
	pid     uint8
	pm      types.ProcessManager
	cn      types.Provider
	user    *user.User
	subject *subject.Subject
}

// NewVisit returns a Visit
func NewVisit(pid uint8) *Visit {
	p := &Visit{
		pid: pid,
	}
	return p
}

// ID returns the id of the process
func (p *Visit) ID() uint8 {
	return p.pid
}

// Name returns the name of the process
func (p *Visit) Name() string {
	return "ecrf.visit"
}

// Version returns the version of the process
func (p *Visit) Version() string {
	return "0.0.1"
}

// Init initializes the process
func (p *Visit) Init(reg *types.Register, pm types.ProcessManager, cn types.Provider) error {
	p.pm = pm
	p.cn = cn

	if vp, err := pm.ProcessByName("ecrf.user"); err != nil {
		return err
	} else if v, is := vp.(*user.User); !is {
		return types.ErrInvalidProcess
	} else {
		p.user = v
	}
	if vp, err := pm.ProcessByName("ecrf.subject"); err != nil {
		return err
	} else if v, is := vp.(*subject.Subject); !is {
		return types.ErrInvalidProcess
	} else {
		p.subject = v
	}

	reg.RegisterTransaction(1, &CreateVisit{})
	reg.RegisterTransaction(2, &UpdateVisitData{})
	return nil
}

// OnLoadChain called when the chain loaded
func (p *Visit) OnLoadChain(loader types.LoaderWrapper) error {
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *Visit) BeforeExecuteTransactions(ctw *types.ContextWrapper) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *Visit) AfterExecuteTransactions(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (p *Visit) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
