package app

import (
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
	"github.com/fletaio/fleta_testnet/process/admin"
	"github.com/fletaio/fleta_testnet/process/study"
)

// ECRFApp is app
type ECRFApp struct {
	*types.ApplicationBase
	pm      types.ProcessManager
	cn      types.Provider
	addrMap map[string]common.Address
}

// NewECRFApp returns a ECRFApp
func NewECRFApp() *ECRFApp {
	return &ECRFApp{
		addrMap: map[string]common.Address{
			"ecrf.study": common.MustParseAddress("3CUsUpv9v"),
		},
	}
}

// Name returns the name of the application
func (app *ECRFApp) Name() string {
	return "ECRFApp"
}

// Version returns the version of the application
func (app *ECRFApp) Version() string {
	return "v1.0.0"
}

// Init initializes the consensus
func (app *ECRFApp) Init(reg *types.Register, pm types.ProcessManager, cn types.Provider) error {
	app.pm = pm
	app.cn = cn
	return nil
}

// InitGenesis initializes genesis data
func (app *ECRFApp) InitGenesis(ctw *types.ContextWrapper) error {
	if p, err := app.pm.ProcessByName("fleta.admin"); err != nil {
		return err
	} else if ap, is := p.(*admin.Admin); !is {
		return types.ErrNotExistProcess
	} else {
		if err := ap.InitAdmin(ctw, app.addrMap); err != nil {
			return err
		}
	}
	if p, err := app.pm.ProcessByName("ecrf.study"); err != nil {
		return err
	} else if _, is := p.(*study.Study); !is {
		return types.ErrNotExistProcess
	} else {
	}

	addStudyAccount(ctw, common.MustParsePublicHash("2RqGkxiHZ4NopN9QxKgw93RuSrxX2NnLjv1q1aFDdV9"), common.MustParsePublicHash("2RqGkxiHZ4NopN9QxKgw93RuSrxX2NnLjv1q1aFDdV9"), common.MustParseAddress("3CUsUpv9v"), "ecrf.study")
	return nil
}

// OnLoadChain called when the chain loaded
func (app *ECRFApp) OnLoadChain(loader types.LoaderWrapper) error {
	return nil
}

func addStudyAccount(ctw *types.ContextWrapper, KeyHash common.PublicHash, GenHash common.PublicHash, addr common.Address, name string) {
	acc := &study.StudyAccount{
		Address_: addr,
		Name_:    name,
		KeyHash:  KeyHash,
		GenHash:  GenHash,
	}
	if err := ctw.CreateAccount(acc); err != nil {
		panic(err)
	}
}
