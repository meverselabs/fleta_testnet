package study

import (
	"github.com/fletaio/fleta_testnet/core/types"
)

// TextData returns text data of the id
func (p *Study) TextData(loader types.Loader, ID string) string {
	lw := types.NewLoaderWrapper(p.pid, loader)

	return string(lw.ProcessData(toTextDataKey(ID)))
}

func (p *Study) setTextData(ctw *types.ContextWrapper, ID string, TextData string) {
	ctw.SetProcessData(toTextDataKey(ID), []byte(TextData))
}
