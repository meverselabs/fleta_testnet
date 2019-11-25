package visit

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta_testnet/process/subject"

	"github.com/fletaio/fleta_testnet/process/study"
	"github.com/fletaio/fleta_testnet/process/user"
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
)

// CreateVisit is a CreateVisit
type CreateVisit struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	UserID     string
	SubjectID  string
	VisitID    string
	FormID     string
	VisitNum   uint32
}

// Timestamp returns the timestamp of the transaction
func (tx *CreateVisit) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *CreateVisit) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *CreateVisit) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *CreateVisit) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Visit)

	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}
	if tx.UserID != "__SELF__" {
		if !sp.user.IsUserRole(loader, tx.From(), tx.UserID, []string{"CRC", "SUBI", "PI"}) {
			return user.ErrInvalidRole
		}
	}
	if !sp.subject.HasSubject(loader, tx.From(), tx.SubjectID) {
		return subject.ErrNotExistSubject
	}
	if sp.HasVisit(loader, tx.From(), tx.VisitID) {
		return ErrExistVisit
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
func (tx *CreateVisit) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Visit)

	sp.addVisit(ctw, tx.From(), tx.VisitID)

	return nil
}

// MarshalJSON is a marshaler function
func (tx *CreateVisit) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"subject_id":`)
	if bs, err := json.Marshal(tx.SubjectID); err != nil {
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
	buffer.WriteString(`"form_id":`)
	if bs, err := json.Marshal(tx.FormID); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"visit_num":`)
	if bs, err := json.Marshal(tx.VisitNum); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
