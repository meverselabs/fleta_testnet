package study

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
)

// SiteAccount is a site account
type SiteAccount struct {
	Address_ common.Address
	Name_    string
	KeyHash  common.PublicHash
	GenHash  common.PublicHash
}

// Address returns the address of the account
func (acc *SiteAccount) Address() common.Address {
	return acc.Address_
}

// Name returns the name of the account
func (acc *SiteAccount) Name() string {
	return acc.Name_
}

// IsFormulator returns it is formulator or not
func (acc *SiteAccount) IsFormulator() bool {
	return true
}

// GeneratorHash returns a generator public hash
func (acc *SiteAccount) GeneratorHash() common.PublicHash {
	return acc.GenHash
}

// IsActivated returns it is activated or not
func (acc *SiteAccount) IsActivated() bool {
	return true
}

// Clone returns the clonend value of it
func (acc *SiteAccount) Clone() types.Account {
	c := &SiteAccount{
		Address_: acc.Address_,
		Name_:    acc.Name_,
		KeyHash:  acc.KeyHash.Clone(),
		GenHash:  acc.GenHash.Clone(),
	}
	return c
}

// Validate validates account signers
func (acc *SiteAccount) Validate(loader types.LoaderWrapper, signers []common.PublicHash) error {
	if len(signers) != 1 {
		return types.ErrInvalidSignerCount
	}
	if acc.KeyHash != signers[0] {
		return types.ErrInvalidAccountSigner
	}
	return nil
}

// MarshalJSON is a marshaler function
func (acc *SiteAccount) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"address":`)
	if bs, err := acc.Address_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"name":`)
	if bs, err := json.Marshal(acc.Name_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"key_hash":`)
	if bs, err := acc.KeyHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"gen_hash":`)
	if bs, err := acc.GenHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
