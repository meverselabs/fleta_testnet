package query

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta_testnet/process/study"
	"github.com/fletaio/fleta_testnet/process/user"
	"github.com/fletaio/fleta_testnet/process/visit"
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
)

// CreateQuery is a CreateQuery
type CreateQuery struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	UserID     string
	VisitID    string
	QueryID    string
	Type       string
	ItemID     string
	Rowindex   uint32
	Message    string
}

// Timestamp returns the timestamp of the transaction
func (tx *CreateQuery) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *CreateQuery) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *CreateQuery) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *CreateQuery) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Query)

	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}
	if !sp.user.IsUserRole(loader, tx.From(), tx.UserID, []string{"CRC", "SUBI", "PI", "CRA", "DM"}) {
		return user.ErrInvalidRole
	}
	if !sp.visit.HasVisit(loader, tx.From(), tx.VisitID) {
		return visit.ErrNotExistVisit
	}
	if _, err := sp.IsOpenQuery(loader, tx.From(), tx.QueryID); err != nil {
	} else {
		return ErrExistQuery
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
func (tx *CreateQuery) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Query)

	sp.addQuery(ctw, tx.From(), tx.QueryID)

	return nil
}

// MarshalJSON is a marshaler function
func (tx *CreateQuery) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"visit_id":`)
	if bs, err := json.Marshal(tx.VisitID); err != nil {
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
	buffer.WriteString(`"type":`)
	if bs, err := json.Marshal(tx.Type); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"item_id":`)
	if bs, err := json.Marshal(tx.ItemID); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"rowindex":`)
	if bs, err := json.Marshal(tx.Rowindex); err != nil {
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
