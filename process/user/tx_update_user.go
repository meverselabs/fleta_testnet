package user

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta_testnet/process/study"
	"github.com/fletaio/fleta_testnet/common/hash"

	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
)

// UpdateUser is a UpdateUser
type UpdateUser struct {
	Timestamp_   uint64
	Seq_         uint64
	From_        common.Address
	UserID       string
	UserName     string
	PasswordHash hash.Hash256
}

// Timestamp returns the timestamp of the transaction
func (tx *UpdateUser) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *UpdateUser) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *UpdateUser) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *UpdateUser) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*User)

	if len(tx.UserID) == 0 {
		return ErrInvalidUserID
	}
	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}
	if !sp.HasUser(loader, tx.From(), tx.UserID) {
		return ErrNotExistUser
	}

	fromAcc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	if _, is := fromAcc.(*study.SiteAccount); !is {
		return study.ErrNotSiteAccount
	}
	if err := fromAcc.Validate(loader, signers); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *UpdateUser) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	return nil
}

// MarshalJSON is a marshaler function
func (tx *UpdateUser) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"user_id":`)
	if bs, err := json.Marshal(tx.UserID); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"user_name":`)
	if bs, err := json.Marshal(tx.UserName); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"password_hash":`)
	if bs, err := tx.PasswordHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
