package study

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
	"github.com/fletaio/fleta_testnet/process/admin"
)

// DeleteSite is used to remove site account
type DeleteSite struct {
	Timestamp_  uint64
	Seq_        uint64
	From_       common.Address
	SiteAddress common.Address
}

// Timestamp returns the timestamp of the transaction
func (tx *DeleteSite) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *DeleteSite) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *DeleteSite) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *DeleteSite) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Study)

	if tx.From() != sp.admin.AdminAddress(loader, p.Name()) {
		return admin.ErrUnauthorizedTransaction
	}
	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	acc, err := loader.Account(tx.SiteAddress)
	if err != nil {
		return err
	}
	if _, is := acc.(*SiteAccount); !is {
		return types.ErrInvalidAccountType
	}

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
func (tx *DeleteSite) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	acc, err := ctw.Account(tx.SiteAddress)
	if err != nil {
		return err
	}
	siteAcc := acc.(*SiteAccount)
	if err := ctw.DeleteAccount(siteAcc); err != nil {
		return err
	}
	return nil
}

// MarshalJSON is a marshaler function
func (tx *DeleteSite) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"site_address":`)
	if bs, err := tx.SiteAddress.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
