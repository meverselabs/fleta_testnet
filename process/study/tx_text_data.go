package study

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
)

// TextData is a TextData
type TextData struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	ID         string
	TextData   string
}

// Timestamp returns the timestamp of the transaction
func (tx *TextData) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *TextData) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *TextData) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *TextData) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	fromAcc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	if err := fromAcc.Validate(loader, signers); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *TextData) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Study)
	sp.setTextData(ctw, tx.ID, tx.TextData)
	return nil
}

// MarshalJSON is a marshaler function
func (tx *TextData) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"timestamp":`)
	if bs, err := json.Marshal(tx.Timestamp_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"seq":`)
	if bs, err := json.Marshal(tx.Seq_); err != nil {
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
	buffer.WriteString(`"id":`)
	if bs, err := json.Marshal(tx.ID); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"text_data":`)
	if bs, err := json.Marshal(tx.TextData); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
