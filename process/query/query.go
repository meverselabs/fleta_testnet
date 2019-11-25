package query

import (
	"github.com/fletaio/fleta_testnet/process/subject"
	"github.com/fletaio/fleta_testnet/process/user"
	"github.com/fletaio/fleta_testnet/process/visit"
	"github.com/fletaio/fleta_testnet/core/types"
)

// Query manages balance of accounts of the chain
type Query struct {
	*types.ProcessBase
	pid     uint8
	pm      types.ProcessManager
	cn      types.Provider
	user    *user.User
	subject *subject.Subject
	visit   *visit.Visit
}

// NewQuery returns a Query
func NewQuery(pid uint8) *Query {
	p := &Query{
		pid: pid,
	}
	return p
}

// ID returns the id of the process
func (p *Query) ID() uint8 {
	return p.pid
}

// Name returns the name of the process
func (p *Query) Name() string {
	return "ecrf.query"
}

// Version returns the version of the process
func (p *Query) Version() string {
	return "0.0.1"
}

// Init initializes the process
func (p *Query) Init(reg *types.Register, pm types.ProcessManager, cn types.Provider) error {
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
	if vp, err := pm.ProcessByName("ecrf.visit"); err != nil {
		return err
	} else if v, is := vp.(*visit.Visit); !is {
		return types.ErrInvalidProcess
	} else {
		p.visit = v
	}

	reg.RegisterTransaction(1, &CreateQuery{})
	reg.RegisterTransaction(2, &AppendMessage{})
	reg.RegisterTransaction(3, &CloseQuery{})
	return nil
}

// OnLoadChain called when the chain loaded
func (p *Query) OnLoadChain(loader types.LoaderWrapper) error {
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *Query) BeforeExecuteTransactions(ctw *types.ContextWrapper) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *Query) AfterExecuteTransactions(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (p *Query) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
