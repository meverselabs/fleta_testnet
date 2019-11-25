package study

import (
	"github.com/fletaio/fleta_testnet/core/types"
	"github.com/fletaio/fleta_testnet/process/admin"
	"github.com/fletaio/fleta_testnet/service/apiserver"
)

// Study manages balance of accounts of the chain
type Study struct {
	*types.ProcessBase
	pid   uint8
	pm    types.ProcessManager
	cn    types.Provider
	admin *admin.Admin
}

// NewStudy returns a study
func NewStudy(pid uint8) *Study {
	p := &Study{
		pid: pid,
	}
	return p
}

// ID returns the id of the process
func (p *Study) ID() uint8 {
	return p.pid
}

// Name returns the name of the process
func (p *Study) Name() string {
	return "ecrf.study"
}

// Version returns the version of the process
func (p *Study) Version() string {
	return "0.0.1"
}

// Init initializes the process
func (p *Study) Init(reg *types.Register, pm types.ProcessManager, cn types.Provider) error {
	p.pm = pm
	p.cn = cn

	if vp, err := pm.ProcessByName("fleta.admin"); err != nil {
		return err
	} else if v, is := vp.(*admin.Admin); !is {
		return types.ErrInvalidProcess
	} else {
		p.admin = v
	}
	if vs, err := pm.ServiceByName("fleta.apiserver"); err != nil {
		//ignore when not loaded
	} else if v, is := vs.(*apiserver.APIServer); !is {
		//ignore when not loaded
	} else {
		s, err := v.JRPC("study")
		if err != nil {
			return err
		}
		s.Set("meta", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			loader := cn.NewLoaderWrapper(p.ID())
			Forms, err := p.MetaData(loader)
			if err != nil {
				return nil, err
			}
			return Forms, nil
		})
	}

	reg.RegisterAccount(1, &StudyAccount{})
	reg.RegisterAccount(2, &SiteAccount{})
	reg.RegisterTransaction(1, &UpdateMeta{})
	reg.RegisterTransaction(2, &CreateSite{})
	reg.RegisterTransaction(3, &DeleteSite{})
	reg.RegisterTransaction(4, &UpdateSite{})
	return nil
}

// OnLoadChain called when the chain loaded
func (p *Study) OnLoadChain(loader types.LoaderWrapper) error {
	p.admin.AdminAddress(loader, p.Name())
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (p *Study) BeforeExecuteTransactions(ctw *types.ContextWrapper) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *Study) AfterExecuteTransactions(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (p *Study) OnSaveData(b *types.Block, ctw *types.ContextWrapper) error {
	return nil
}
