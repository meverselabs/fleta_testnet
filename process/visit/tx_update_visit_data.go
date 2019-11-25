package visit

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/fletaio/fleta_testnet/process/user"

	"github.com/fletaio/fleta_testnet/process/study"
	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/core/types"
)

// UpdateVisitData is a UpdateVisitData
type UpdateVisitData struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	UserID     string
	VisitID    string
	Data       *types.StringBytesMap
}

// Timestamp returns the timestamp of the transaction
func (tx *UpdateVisitData) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *UpdateVisitData) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *UpdateVisitData) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *UpdateVisitData) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Visit)

	var inErr error
	tx.Data.EachAll(func(Key string, Value []byte) bool {
		ls := strings.Split(Key, "_")
		if len(ls) > 1 {
			if _, err := strconv.ParseUint(ls[1], 10, 32); err != nil {
				inErr = err
				return false
			}
		}
		return true
	})
	if inErr != nil {
		return inErr
	}

	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}
	if tx.UserID != "__SELF__" {
		if !sp.user.IsUserRole(loader, tx.From(), tx.UserID, []string{"CRC", "SUBI", "PI"}) {
			return user.ErrInvalidRole
		}
	}
	if !sp.HasVisit(loader, tx.From(), tx.VisitID) {
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
func (tx *UpdateVisitData) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	return nil
}

// MarshalJSON is a marshaler function
func (tx *UpdateVisitData) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"data":`)
	buffer.WriteString(`{`)
	isFirst := true
	var inErr error
	tx.Data.EachAll(func(key string, value []byte) bool {
		if isFirst {
			isFirst = false
		} else {
			buffer.WriteString(`,`)
		}
		if bs, err := json.Marshal(key); err != nil {
			inErr = err
			return false
		} else {
			buffer.Write(bs)
		}
		buffer.WriteString(`:`)
		if bs, err := json.Marshal(string(value)); err != nil {
			inErr = err
			return false
		} else {
			buffer.Write(bs)
		}
		return true
	})
	if inErr != nil {
		return nil, inErr
	}
	buffer.WriteString(`}`)
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
