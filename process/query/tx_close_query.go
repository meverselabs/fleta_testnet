package query

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta_testnet/process/study"
	"github.com/fletaio/fleta_testnet/process/user"
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
)

// CloseQuery is a CloseQuery
type CloseQuery struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	UserID     string
	QueryID    string
}

// Timestamp returns the timestamp of the transaction
func (tx *CloseQuery) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *CloseQuery) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *CloseQuery) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *CloseQuery) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Query)

	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}
	if !sp.user.IsUserRole(loader, tx.From(), tx.UserID, []string{"CRA", "DM"}) {
		return user.ErrInvalidRole
	}
	if is, err := sp.IsOpenQuery(loader, tx.From(), tx.QueryID); err != nil {
		return err
	} else if !is {
		return ErrClosedQuery
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
func (tx *CloseQuery) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Query)

	sp.closeQuery(ctw, tx.From(), tx.QueryID)

	return nil
}

// MarshalJSON is a marshaler function
func (tx *CloseQuery) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"query_id":`)
	if bs, err := json.Marshal(tx.QueryID); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
