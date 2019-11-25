package query

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta_testnet/process/user"

	"github.com/fletaio/fleta_testnet/process/study"
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
)

// AppendMessage is a AppendMessage
type AppendMessage struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	UserID     string
	QueryID    string
	Message    string
}

// Timestamp returns the timestamp of the transaction
func (tx *AppendMessage) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *AppendMessage) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *AppendMessage) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *AppendMessage) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Query)

	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}
	if !sp.user.IsUserRole(loader, tx.From(), tx.UserID, []string{"CRC", "SUBI", "PI", "CRA", "DM"}) {
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
func (tx *AppendMessage) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	return nil
}

// MarshalJSON is a marshaler function
func (tx *AppendMessage) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`,`)
	buffer.WriteString(`"message":`)
	if bs, err := json.Marshal(tx.Message); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
