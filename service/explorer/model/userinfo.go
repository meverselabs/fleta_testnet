package model

import (
	"time"

	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/process/gateway"
)

type TransferInfo struct {
	BlockNumber uint32
	BlockHash   hash.Hash256
	TxIndex     uint16
	TxHash      hash.Hash256
	LogIndex    uint16
	FromAddress gateway.ERC20Address
	ToAddress   gateway.ERC20Address
	Amount      *amount.Amount
}

type DepositAddressInfo struct {
	Email          string
	DepositAddress string
}

type LockedAddressInfo struct {
	Address  string
	Comment  string
	CreateAt *time.Time
	updateAt *time.Time
}

type GatewayTxInfo struct {
	Name         string
	ERC20Address string
	BlockNumber  uint32
	BlockHash    hash.Hash256
	TxHash       hash.Hash256
	Amount       *amount.Amount
}
