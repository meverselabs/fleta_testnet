package study

import (
	"encoding/json"

	"github.com/fletaio/fleta_testnet/core/types"
)

// MetaData returns text data of the id
func (p *Study) MetaData(loader types.Loader) ([]*Form, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	bs := lw.ProcessData(tagMetaData)
	if len(bs) == 0 {
		return nil, ErrNotExistMeta
	}
	var Forms []*Form
	if err := json.Unmarshal(bs, &Forms); err != nil {
		return nil, err
	}
	return Forms, nil
}

func (p *Study) setMetaData(ctw *types.ContextWrapper, MetaData []*Form) error {
	bs, err := json.Marshal(MetaData)
	if err != nil {
		return err
	}
	ctw.SetProcessData(tagMetaData, bs)
	return nil
}

// TextData returns text data of the id
func (p *Study) TextData(loader types.Loader, ID string) string {
	lw := types.NewLoaderWrapper(p.pid, loader)

	return string(lw.ProcessData(toTextDataKey(ID)))
}

func (p *Study) setTextData(ctw *types.ContextWrapper, ID string, TextData string) {
	ctw.SetProcessData(toTextDataKey(ID), []byte(TextData))
}
