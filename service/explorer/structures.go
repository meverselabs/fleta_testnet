package explorer

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

type Delegation struct {
	HyperAddr common.Address
	Address   common.Address
	Amount    *amount.Amount
}

type Transaction struct {
	Type   uint16
	Data   []byte
	Result uint8
}

type Unstaking struct {
	UnstakedHeight uint32
	HyperAddr      common.Address
	Address        common.Address
	Amount         *amount.Amount
}

type Reward struct {
	Amount    *amount.Amount
	Height    uint32
	Timestamp uint64
}

type HyperReward struct {
	Formulator *amount.Amount
	Commission *amount.Amount
	Delegation *amount.Amount
	Height     uint32
	Timestamp  uint64
}
type HyperRewardData struct {
	Formulator string
	Commission string
	Delegation string
	Height     uint32
	Timestamp  uint64
}
type HyperData struct {
	Addr       string
	Name       string
	Formulator string
	Locked     string
	Unstaking  string
	Delegation string
	Commission string
	Height     uint32
	Timestamp  uint64
}

type FormulatorData struct {
	Name           string
	Addr           string
	CreateBlock    uint32
	FormulatorType string
	StakingAmount  string
	GenerateBlocks uint32
}

type BlockHeaderData struct {
	ChainID       uint8
	Version       uint16
	Height        uint32
	PrevHash      string
	LevelRootHash string
	ContextHash   string
	Timestamp     uint64
	Generator     string
	ConsensusData string
}

type TransactionDetailData struct {
	err            string
	Type           string
	Result         string
	BlockHash      string
	BlockTimestamp int64
	TxHash         string
	TxTimeStamp    int64
}

func NewBlockHeaderData(h types.Header) *BlockHeaderData {
	return &BlockHeaderData{
		ChainID:       h.ChainID,
		Version:       h.Version,
		Height:        h.Height,
		PrevHash:      h.PrevHash.String(),
		LevelRootHash: h.LevelRootHash.String(),
		ContextHash:   h.ContextHash.String(),
		Timestamp:     h.Timestamp,
		Generator:     h.Generator.String(),
		ConsensusData: string(h.ConsensusData),
	}
}
func NewHyperReward(Height uint32, Timestamp uint64) *HyperReward {
	return &HyperReward{
		Formulator: amount.NewCoinAmount(0, 0),
		Commission: amount.NewCoinAmount(0, 0),
		Delegation: amount.NewCoinAmount(0, 0),
		Height:     Height,
		Timestamp:  Timestamp,
	}
}
