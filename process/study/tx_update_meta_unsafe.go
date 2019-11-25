package study

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
)

// UpdateMetaUnsafe is a UpdateMetaUnsafe
type UpdateMetaUnsafe struct {
	Timestamp_ uint64
	From_      common.Address
	Forms      []*Form
}

// Timestamp returns the timestamp of the transaction
func (tx *UpdateMetaUnsafe) Timestamp() uint64 {
	return tx.Timestamp_
}

// From returns the from address of the transaction
func (tx *UpdateMetaUnsafe) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *UpdateMetaUnsafe) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	fromAcc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	if _, is := fromAcc.(*StudyAccount); !is {
		return ErrNotStudyAccount
	}
	if err := fromAcc.Validate(loader, signers); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *UpdateMetaUnsafe) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	return nil
}

// MarshalJSON is a marshaler function
func (tx *UpdateMetaUnsafe) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"timestamp":`)
	if bs, err := json.Marshal(tx.Timestamp_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"from":`)
	if bs, err := tx.From_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"forms":`)
	buffer.WriteString(`[`)
	for i, f := range tx.Forms {
		if i > 0 {
			buffer.WriteString(`,`)
		}
		if bs, err := json.Marshal(f); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
	}
	buffer.WriteString(`]`)
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
