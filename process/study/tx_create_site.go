package study

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/common/hash"
	"github.com/fletaio/fleta_testnet/core/types"
	"github.com/fletaio/fleta_testnet/process/admin"
)

// CreateSite is used to make alpha study account
type CreateSite struct {
	Timestamp_   uint64
	Seq_         uint64
	From_        common.Address
	SiteID       string
	KeyHash      common.PublicHash
	GenHash      common.PublicHash
	SiteName     string
	AdminID      string
	PasswordHash hash.Hash256
}

// Timestamp returns the timestamp of the transaction
func (tx *CreateSite) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *CreateSite) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *CreateSite) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *CreateSite) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Study)

	if tx.From() != sp.admin.AdminAddress(loader, p.Name()) {
		return admin.ErrUnauthorizedTransaction
	}
	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	if has, err := loader.HasAccountName(tx.SiteID); err != nil {
		return err
	} else if has {
		return types.ErrExistAccountName
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
func (tx *CreateSite) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	acc := &SiteAccount{
		Address_: common.NewAddress(ctw.TargetHeight(), index, 0),
		Name_:    tx.SiteID,
		KeyHash:  tx.KeyHash,
		GenHash:  tx.GenHash,
	}
	if err := ctw.CreateAccount(acc); err != nil {
		return err
	}
	return nil
}

// MarshalJSON is a marshaler function
func (tx *CreateSite) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"site_id":`)
	if bs, err := json.Marshal(tx.SiteID); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"key_hash":`)
	if bs, err := tx.KeyHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"gen_hash":`)
	if bs, err := tx.GenHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"site_name":`)
	if bs, err := json.Marshal(tx.SiteName); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"admin_id":`)
	if bs, err := json.Marshal(tx.AdminID); err != nil {
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
