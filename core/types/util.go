package types

import (
	"bytes"
	"encoding/hex"
	"strings"
	"time"

	"github.com/fletaio/fleta_testnet/common"
	"github.com/fletaio/fleta_testnet/common/binutil"
	"github.com/fletaio/fleta_testnet/common/factory"
	"github.com/fletaio/fleta_testnet/common/hash"
	"github.com/fletaio/fleta_testnet/encoding"
	"github.com/petar/GoLLRB/llrb"
)

var (
	ninf = nInf{}
	pinf = pInf{}
)

type nInf struct{}

func (nInf) Less(llrb.Item) bool {
	return true
}

type pInf struct{}

func (pInf) Less(llrb.Item) bool {
	return false
}

func cmpAddressASC(a interface{}, b interface{}) bool {
	ai := a.(common.Address)
	bi := b.(common.Address)
	return bytes.Compare(ai[:], bi[:]) < 0
}

func cmpStringASC(a interface{}, b interface{}) bool {
	ai := a.(string)
	bi := b.(string)
	return strings.Compare(ai[:], bi[:]) < 0
}

func cmpUint64ASC(a interface{}, b interface{}) bool {
	ai := a.(uint64)
	bi := b.(uint64)
	return ai < bi
}

// IsAllowedAccountName returns it is allowed account name or not
func IsAllowedAccountName(Name string) bool {
	if len(Name) < 8 || len(Name) > 40 {
		return false
	}
	if _, err := common.ParseAddress(Name); err == nil {
		return false
	}
	for i := 0; i < len(Name); i++ {
		c := Name[i]
		if (c < '0' || '9' < c) && (c < 'a' || 'z' < c) && (c < 'A' || 'Z' < c) && c != '.' && c != '-' && c != '_' && c != '@' {
			return false
		}
	}
	return true
}

// UnmarshalID returns the block height, the transaction index in the block, the output index in the transaction
func UnmarshalID(id uint64) (uint32, uint16, uint16) {
	return uint32(id >> 32), uint16(id >> 16), uint16(id)
}

// MarshalID returns the packed id
func MarshalID(height uint32, index uint16, n uint16) uint64 {
	return uint64(height)<<32 | uint64(index)<<16 | uint64(n)
}

// TransactionID returns the id of the transaction
func TransactionID(Height uint32, Index uint16) string {
	bs := make([]byte, 6)
	binutil.BigEndian.PutUint32(bs, Height)
	binutil.BigEndian.PutUint16(bs[4:], Index)
	return hex.EncodeToString(bs)
}

// ParseTransactionID returns the id of the transaction
func ParseTransactionID(TXID string) (uint32, uint16, error) {
	if len(TXID) != 12 {
		return 0, 0, ErrInvalidTransactionIDFormat
	}
	bs, err := hex.DecodeString(TXID)
	if err != nil {
		return 0, 0, err
	}
	Height := binutil.BigEndian.Uint32(bs)
	Index := binutil.BigEndian.Uint16(bs[4:])
	return Height, Index, nil
}

// ToTimeSlot returns the timeslot of the timestamp
func ToTimeSlot(timestamp uint64) uint32 {
	return uint32(timestamp / uint64(5*time.Second))
}

// HashTransaction returns the hash of the transaction
func HashTransaction(ChainID uint8, tx Transaction) hash.Hash256 {
	fc := encoding.Factory("transaction")
	t, err := fc.TypeOf(tx)
	if err != nil {
		panic(err)
	}
	return HashTransactionByType(ChainID, t, tx)
}

// HashTransactionByType returns the hash of the transaction using the type
func HashTransactionByType(ChainID uint8, t uint16, tx Transaction) hash.Hash256 {
	bs, err := EncodeTransaction(ChainID, t, tx)
	if err != nil {
		panic(err)
	}
	return hash.Hash(bs)
}

// EncodeTransaction returns the encoded transaction
func EncodeTransaction(ChainID uint8, t uint16, tx Transaction) ([]byte, error) {
	var buffer bytes.Buffer
	enc := encoding.NewEncoder(&buffer)
	if err := enc.EncodeUint8(ChainID); err != nil {
		return nil, err
	}
	if err := enc.EncodeUint16(t); err != nil {
		return nil, err
	}
	if err := enc.Encode(tx); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// DecodeTransaction returns the encoded transaction
func DecodeTransaction(fc *factory.Factory, bs []byte) (uint8, Transaction, uint16, error) {
	dec := encoding.NewDecoder(bytes.NewReader(bs))
	ChainID, err := dec.DecodeUint8()
	if err != nil {
		return 0, nil, 0, err
	}
	t, err := dec.DecodeUint16()
	if err != nil {
		return 0, nil, 0, err
	}
	tx, err := fc.Create(t)
	if err != nil {
		return 0, nil, 0, err
	}
	if err := dec.Decode(&tx); err != nil {
		return 0, nil, 0, err
	}
	return ChainID, tx.(Transaction), t, nil
}
