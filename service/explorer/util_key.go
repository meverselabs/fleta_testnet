package explorer

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/binutil"
	"github.com/fletaio/fleta/process/formulator"
)

// tags
var (
	tagInitDB                = []byte{0, 0}
	tagHeight                = []byte{0, 1}
	tagTotalSupply           = []byte{0, 2}
	tagFormulatorAddress     = []byte{0, 3}
	tagHyperAddress          = []byte{0, 4}
	tagFormulatorType        = []byte{0, 5}
	tagAccountName           = []byte{0, 6}
	tagBlockHashHeight       = []byte{1, 0}
	tagHashBlock             = []byte{1, 1}
	tagFormulatorBlockList   = []byte{1, 2}
	tagTransaction           = []byte{2, 0}
	tagTransactionList       = []byte{2, 1}
	tagTransactionType       = []byte{2, 2}
	tagGatewayTxList         = []byte{2, 3}
	tagGatewayTokenInAmount  = []byte{2, 4}
	tagGatewayTokenInList    = []byte{2, 5}
	tagGatewayTokenOutAmount = []byte{2, 6}
	tagGatewayTokenOutList   = []byte{2, 7}
	tagGatewayTokenLeave     = []byte{2, 8}
	tagDelegation            = []byte{3, 0}
	tagUnstaking             = []byte{3, 1}
	tagHyperUnstaking        = []byte{3, 2}
	tagFormulatorReward      = []byte{4, 0}
	tagFormulatorRewardList  = []byte{4, 1}
	tagDelegationReward      = []byte{4, 2}
	tagDelegationRewardList  = []byte{4, 3}
	tagHyperReward           = []byte{4, 4}
	tagHyperRewardList       = []byte{4, 5}
	tagERC20TransactionCount = []byte{5, 1}
	tagERC20Transaction      = []byte{5, 1}
	tagERC20TransactionList  = []byte{5, 2}
	tagERC20LockedAddresses  = []byte{5, 3}
	tagFormulationLocked     = []byte{7, 1}
	tagDelegationLocked      = []byte{7, 2}
)

func toHashBlock(height uint32) []byte {
	bs := make([]byte, 6)
	copy(bs, tagHashBlock)
	copy(bs[2:], binutil.LittleEndian.Uint32ToBytes(height))
	return bs
}

func toFormulatorTypeKey(t formulator.FormulatorType) []byte {
	bs := make([]byte, 3)
	copy(bs, tagFormulatorType)
	bs[2] = uint8(t)
	return bs
}

func toFormulatorBlockListKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagFormulatorBlockList)
	copy(bs[2:], addr[:])
	return bs
}

func toTransactionListKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagTransactionList)
	copy(bs[2:], addr[:])
	return bs
}

func toTransactionTypeKey(t uint16) []byte {
	bs := make([]byte, 4)
	copy(bs, tagTransactionType)
	copy(bs[2:], binutil.LittleEndian.Uint16ToBytes(t))
	return bs
}

func toGatewayTxListKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagGatewayTxList)
	copy(bs[2:], addr[:])
	return bs
}

func toDelegationKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagDelegation)
	copy(bs[2:], addr[:])
	return bs
}

func toFormulatorRewardListKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagFormulatorRewardList)
	copy(bs[2:], addr[:])
	return bs
}

func toUnstakingKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagUnstaking)
	copy(bs[2:], addr[:])
	return bs
}

func toUnstakingSubKey(haddr common.Address, UnstakedHeight uint32) []byte {
	bs := make([]byte, 6+common.AddressSize)
	copy(bs, tagUnstaking)
	copy(bs[2:], binutil.LittleEndian.Uint32ToBytes(UnstakedHeight))
	copy(bs[6:], haddr[:])
	return bs
}

func toDelegationRewardListKey(haddr common.Address, addr common.Address) []byte {
	bs := make([]byte, 2+(common.AddressSize*2))
	copy(bs, tagDelegationRewardList)
	copy(bs[2:], haddr[:])
	copy(bs[16:], addr[:])
	return bs
}

func fromUnstakingSubKey(bs []byte) (common.Address, uint32) {
	UnstakedHeight := binutil.LittleEndian.Uint32(bs[2:])
	var haddr common.Address
	copy(haddr[:], bs[6:])
	return haddr, UnstakedHeight
}

func toHyperRewardListKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagHyperRewardList)
	copy(bs[2:], addr[:])
	return bs
}

/*
var (
	tagAccountAddress              = []byte{1, 1}
	tagAddressKeyHash              = []byte{1, 2}
	tagFormulatorAddress           = []byte{1, 3}
	tagHyperAddress                = []byte{1, 4}
	tagUnstaking                   = []byte{4, 0}
	tagHyperUnstaking              = []byte{4, 1}
	tagHyperStakingReward          = []byte{4, 2}
	tagCommissionFormulator        = []byte{5, 2}
	tagCommissionFormulatorList    = []byte{5, 3}
	tagRewardHyper                 = []byte{5, 4}
	tagRewardHyperList             = []byte{5, 5}
	tagRewardStaking               = []byte{5, 6}
	tagRewardStakingList           = []byte{5, 7}
	tagDependenceRewardStaking     = []byte{5, 8}
	tagDependenceRewardStakingList = []byte{5, 9}
)

func toAccountKey(pubhash common.PublicHash) []byte {
	bs := make([]byte, 2+common.PublicHashSize)
	copy(bs, tagAccountAddress)
	copy(bs[2:], pubhash[:])
	return bs
}

func toFormulatorKey(pubhash common.PublicHash) []byte {
	bs := make([]byte, 2+common.PublicHashSize)
	copy(bs, tagFormulatorAddress)
	copy(bs[2:], pubhash[:])
	return bs
}

func toCommissionFormulatorListKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagCommissionFormulatorList)
	copy(bs[2:], addr[:])
	return bs
}

func toRewardHyperListKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagRewardHyperList)
	copy(bs[2:], addr[:])
	return bs
}

func toRewardStakingListKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagRewardStakingList)
	copy(bs[2:], addr[:])
	return bs
}

func toDependenceRewardStakingListKey(haddr common.Address, addr common.Address) []byte {
	bs := make([]byte, 2+(common.AddressSize*2))
	copy(bs, tagRewardStakingList)
	copy(bs[2:], haddr[:])
	copy(bs[16:], addr[:])
	return bs
}

*/
