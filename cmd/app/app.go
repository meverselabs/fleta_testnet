package app

import (
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/common/amount"
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

	addStudyAccount(ctw, common.MustParsePublicHash("iUqb4PxXQ12JShdtEsb6SLipFFPHmSLW29zqHKGjvB"), common.MustParsePublicHash("4YjmYcLVvBSmtjh4Z7frRZhWgdEAYTSABCoqqzhKEJa"), common.MustParseAddress("5CyLcFhpyN"), "ecrf.study01")
	addStudyAccount(ctw, common.MustParsePublicHash("2Jid4fJm3Kf2GD2hvSMTyCbvW5gGCuo2p2oDWo5GhKT"), common.MustParsePublicHash("27n37VV3ebGWSNH5r9wX3ZhUwzxC2heY34UvXjizLDK"), common.MustParseAddress("4wayWtvQuB"), "ecrf.study02")
	addStudyAccount(ctw, common.MustParsePublicHash("4oV8S1dEuTKQrsac7CS81jZdQQpiG31CgoUd66eHXsk"), common.MustParsePublicHash("4GzTnuP7Hky1Dye1AJMLzEXTX2a5kEka5h9AJVvZyTD"), common.MustParseAddress("4sC1mwGabR"), "ecrf.study03")
	addStudyAccount(ctw, common.MustParsePublicHash("324QLx4QrYrh9hE7dQb8xbmy4anyCvn6cGaE5jt3qE"), common.MustParsePublicHash("4ew8HQEwwSqeepMDCnwN9PiYg1uvoeZXyudqdQZBCb3"), common.MustParseAddress("58aNsJ3zfc"), "ecrf.study04")
	addStudyAccount(ctw, common.MustParsePublicHash("mVssPMvS4RnSK6LmpYrWbXVxxhhE5AAyRbuU8Br74r"), common.MustParsePublicHash("VbMwA5AwSfn93ks8HMv7vvSx4THuzfeefTWVoANEha"), common.MustParseAddress("4gCcRY8zq4"), "ecrf.study05")
	addStudyAccount(ctw, common.MustParsePublicHash("2Egzma6KP4yERrhEAeBdFiBEhCQHFyDaaJ1vGR1DYKf"), common.MustParsePublicHash("8eDJ3h8DLW8RSovYUjxmcDi1QNvo7UW64MQxGZ9dnS"), common.MustParseAddress("5mwYfT6aH5"), "ecrf.study06")
	addStudyAccount(ctw, common.MustParsePublicHash("3z1S6ZzWKGfSHmW519sDBgSvoWJthzcprhJziofdNHQ"), common.MustParsePublicHash("3ZdKaqaCbGSQ5xmAphzVTeEF1eGzX6iU4LLGD2ox2g9"), common.MustParseAddress("51ywFraFCw"), "ecrf.study07")
	addStudyAccount(ctw, common.MustParsePublicHash("4X8Fbz4HurLjpbdBsmhmqNbd8an7aPmCrRPRDLGkqVe"), common.MustParsePublicHash("3UHQyJwSSHHCw29fB5xiGk9W7GNf1DjGC284WhW6jpD"), common.MustParseAddress("4uPVeR6VkL"), "ecrf.study08")
	addStudyAccount(ctw, common.MustParsePublicHash("49DaZWMvaiJU5DuZGwTJn99sMn4UuTEVU1CUKhHSPSi"), common.MustParsePublicHash("v3GwqbQehcqNVYbRzDk3TDJ7yJ19DgwoamZnMJZuVg"), common.MustParseAddress("4no42yckHf"), "ecrf.study09")
	addStudyAccount(ctw, common.MustParsePublicHash("3yADixVW3KxWFhf1dNHkoDFbJCsKLPArQyg5btbh6nB"), common.MustParsePublicHash("3HhrC3gPR951SjnxjnHpfhRSWH1iR3SbCSwtCHvTLuC"), common.MustParseAddress("54BR8LQAMr"), "ecrf.study10")
	addStudyAccount(ctw, common.MustParsePublicHash("2tqyyKDje5iiTD8Wvm6VyRSagN1QRzGDwrevhq1kmaJ"), common.MustParsePublicHash("4Ei1HSF3KtDfGrdzHCWfRf4NSTZ2oYCT1CNGFkjV1WB"), common.MustParseAddress("5p92XvvVRz"), "ecrf.study11")
	addStudyAccount(ctw, common.MustParsePublicHash("4aDsBd3UCd74BsMYpfeZmxeUWeRph9WnDi14HHMFCKT"), common.MustParsePublicHash("3u6v76WAknSq1j86Pfb6p31FsBAJztPdVmY1kkw4k66"), common.MustParseAddress("5hYavVSjyK"), "ecrf.study12")

	setupSingleAccunt(ctw)
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

func addSingleAccount(ctw *types.ContextWrapper, KeyHash common.PublicHash, addr common.Address, name string, am *amount.Amount) {
	acc := &study.StudyAccount{
		Address_: addr,
		Name_:    name,
		KeyHash:  KeyHash,
		GenHash:  KeyHash,
	}
	if err := ctw.CreateAccount(acc); err != nil {
		panic(err)
	}
}
