package explorer

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/common/binutil"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/process/formulator"
)

func (ex *Explorer) isInitDB() bool {
	bs, err := ex.db.Get(tagInitDB)
	if err != nil {
		return false
	}
	return len(bs) > 0
}

func (ex *Explorer) setInitDB() error {
	if err := ex.db.Set(tagInitDB, []byte{1}); err != nil {
		return err
	}
	return nil
}

func (ex *Explorer) Height() uint32 {
	bs, err := ex.db.Get(tagHeight)
	if err != nil {
		return 0
	}
	if len(bs) == 0 {
		return 0
	}
	return binutil.LittleEndian.Uint32(bs)
}

func (ex *Explorer) setHeight(height uint32) error {
	if err := ex.db.Set(tagHeight, binutil.LittleEndian.Uint32ToBytes(height)); err != nil {
		return err
	}
	return nil
}

func (ex *Explorer) AccountNames() ([]string, error) {
	pairs, err := ex.db.HGetAll(tagAccountName)
	if err != nil {
		return nil, err
	}
	list := make([]string, 0, len(pairs))
	for _, v := range pairs {
		list = append(list, string(v.Value))
	}
	return list, nil
}

func (ex *Explorer) AccountName(addr common.Address) (string, error) {
	bs, err := ex.db.HGet(tagAccountName, addr[:])
	if err != nil {
		return "", err
	}
	if len(bs) == 0 {
		return "", ErrNotExist
	}
	return string(bs), nil
}

func (ex *Explorer) setAccountName(addr common.Address, Name string) error {
	if _, err := ex.db.HSet(tagAccountName, addr[:], []byte(Name)); err != nil {
		return err
	}
	return nil
}

func (ex *Explorer) TotalSupply() *amount.Amount {
	bs, err := ex.db.Get(tagTotalSupply)
	if err != nil {
		return amount.NewCoinAmount(0, 0)
	}
	if len(bs) == 0 {
		return amount.NewCoinAmount(0, 0)
	}
	return amount.NewAmountFromBytes(bs)
}

func (ex *Explorer) setTotalSupply(am *amount.Amount) error {
	if err := ex.db.Set(tagTotalSupply, am.Bytes()); err != nil {
		return err
	}
	return nil
}

func (ex *Explorer) TransactionCount() uint64 {
	n, err := ex.db.HLen(tagTransaction)
	if err != nil {
		return 0
	}
	return uint64(n)
}

func (ex *Explorer) BlockHeightByHash(h hash.Hash256) (uint32, error) {
	bs, err := ex.db.HGet(tagBlockHashHeight, h[:])
	if err != nil {
		return 0, err
	}
	return binutil.LittleEndian.Uint32(bs), nil
}

func (ex *Explorer) setBlockHash(height uint32, h hash.Hash256) error {
	if err := ex.db.Set(toHashBlock(height), h[:]); err != nil {
		return err
	}
	if _, err := ex.db.HSet(tagBlockHashHeight, h[:], binutil.LittleEndian.Uint32ToBytes(height)); err != nil {
		return err
	}
	return nil
}

func (ex *Explorer) FormulatorCount() uint32 {
	n, err := ex.db.HLen(tagFormulatorAddress)
	if err != nil {
		return 0
	}
	return uint32(n)
}

func (ex *Explorer) IsFormulator(addr common.Address) bool {
	if bs, err := ex.db.HGet(tagFormulatorAddress, addr[:]); err != nil {
		panic(err)
	} else {
		return len(bs) > 0
	}
}

func (ex *Explorer) getFormulatorCount(t formulator.FormulatorType) uint32 {
	n, err := ex.db.HLen(toFormulatorTypeKey(t))
	if err != nil {
		return 0
	}
	return uint32(n)
}

func (ex *Explorer) addFormulator(addr common.Address, am *amount.Amount, t formulator.FormulatorType) error {
	if _, err := ex.db.HSet(tagFormulatorAddress, addr[:], []byte{uint8(t)}); err != nil {
		return err
	}
	if _, err := ex.db.HSet(toFormulatorTypeKey(t), addr[:], am.Bytes()); err != nil {
		return err
	}
	if err := ex.addFormulatorLocked(addr, am, t); err != nil {
		return err
	}
	return nil
}

func (ex *Explorer) removeFormulator(addr common.Address) error {
	t, err := ex.FormulatorType(addr)
	if err != nil {
		return err
	}
	if _, err := ex.db.HDel(tagFormulatorAddress, addr[:]); err != nil {
		return err
	}
	if _, err := ex.db.HDel(toFormulatorTypeKey(t), addr[:]); err != nil {
		return err
	}
	if err := ex.removeFormulatorLocked(addr, t); err != nil {
		return err
	}
	return nil
}

func (ex *Explorer) FormulatorType(addr common.Address) (formulator.FormulatorType, error) {
	bs, err := ex.db.HGet(tagFormulatorAddress, addr[:])
	if err != nil {
		return 0, err
	}
	if len(bs) == 0 {
		return 0, ErrNotExist
	}
	return formulator.FormulatorType(bs[0]), nil
}

func (ex *Explorer) FormulatorLocked(addr common.Address) *amount.Amount {
	bs, err := ex.db.HGet(tagFormulationLocked, addr[:])
	if err != nil {
		return amount.NewCoinAmount(0, 0)
	}
	if len(bs) == 0 {
		return amount.NewCoinAmount(0, 0)
	}
	return amount.NewAmountFromBytes(bs)
}

func (ex *Explorer) getFormulatorLockedTotal() *amount.Amount {
	bs, err := ex.db.HGet(tagFormulationLocked, []byte{0})
	if err != nil {
		return amount.NewCoinAmount(0, 0)
	}
	if len(bs) == 0 {
		return amount.NewCoinAmount(0, 0)
	}
	return amount.NewAmountFromBytes(bs)
}

func (ex *Explorer) getFormulatorLockedByType(t formulator.FormulatorType) *amount.Amount {
	bs, err := ex.db.HGet(tagFormulationLocked, []byte{0, byte(t)})
	if err != nil {
		return amount.NewCoinAmount(0, 0)
	}
	if len(bs) == 0 {
		return amount.NewCoinAmount(0, 0)
	}
	return amount.NewAmountFromBytes(bs)
}

func (ex *Explorer) addFormulatorLocked(addr common.Address, am *amount.Amount, t formulator.FormulatorType) error {
	if _, err := ex.db.HSet(tagFormulationLocked, addr[:], am.Bytes()); err != nil {
		return err
	}
	if _, err := ex.db.HSet(tagFormulationLocked, []byte{0}, ex.getFormulatorLockedTotal().Add(am).Bytes()); err != nil {
		return err
	}
	if _, err := ex.db.HSet(tagFormulationLocked, []byte{0, byte(t)}, ex.getFormulatorLockedByType(t).Add(am).Bytes()); err != nil {
		return err
	}
	return nil
}

func (ex *Explorer) removeFormulatorLocked(addr common.Address, t formulator.FormulatorType) error {
	am := ex.FormulatorLocked(addr)
	if _, err := ex.db.HDel(tagFormulationLocked, addr[:]); err != nil {
		return err
	}
	if _, err := ex.db.HSet(tagFormulationLocked, []byte{0}, ex.getFormulatorLockedTotal().Sub(am).Bytes()); err != nil {
		return err
	}
	if _, err := ex.db.HSet(tagFormulationLocked, []byte{0, byte(t)}, ex.getFormulatorLockedByType(t).Sub(am).Bytes()); err != nil {
		return err
	}
	return nil
}

func (ex *Explorer) addFormulatorBlock(bh *types.Header) error {
	bs, err := encoding.Marshal(bh)
	if err != nil {
		return err
	}
	if _, err := ex.db.RPush(toFormulatorBlockListKey(bh.Generator), bs); err != nil {
		return err
	}
	return nil
}

func (ex *Explorer) getFormulatorBlockCount(addr common.Address) uint32 {
	n, err := ex.db.LLen(toFormulatorBlockListKey(addr))
	if err != nil {
		return 0
	}
	return uint32(n)
}

func (ex *Explorer) getBlocks(from uint32, to uint32, reverse bool) ([]*HeaderData, error) {
	curr := ex.Height()
	if from > to {
		from, to = to, from
	}
	if to > curr {
		to = curr
	}
	if from > curr {
		from = curr
	}

	hds := make([]*HeaderData, to-from+1)
	for i := from; i <= to; i++ {
		h, err := ex.cn.Header(i)
		if err != nil {
			return nil, err
		}
		var index uint32
		if reverse == true {
			index = to - i
		} else {
			index = i - from
		}
		hds[index] = &HeaderData{
			Height:    h.Height,
			Hash:      h.ContextHash.String(),
			Timestamp: h.Timestamp,
			Generator: h.Generator.String(),
		}
	}

	return hds, nil
}

func (ex *Explorer) getFormulatorBlocks(addr common.Address, from int32, to int32, reverse bool) ([]*types.Header, error) {
	values, err := ex.db.LRange(toFormulatorBlockListKey(addr), from, to)
	if err != nil {
		return nil, err
	}
	list := make([]*types.Header, len(values))
	for i, bs := range values {
		item := &types.Header{}
		if err := encoding.Unmarshal(bs, &item); err != nil {
			return nil, err
		}
		if reverse {
			list[len(values)-1-i] = item
		} else {
			list[i] = item
		}
	}
	return list, nil
}

func (ex *Explorer) getFormulatorBlocksFromTail(addr common.Address, offset int32, count int32) ([]*types.Header, error) {
	s, err := ex.db.LLen(toFormulatorBlockListKey(addr))
	if err != nil {
		return nil, err
	}
	from := int32(s) - count - offset
	if from < 0 {
		from = 0
	}
	to := int32(s) - offset
	if to < 0 {
		to = 0
	}
	return ex.getFormulatorBlocks(addr, from, to, true)
}

func (ex *Explorer) getTransaction(txid string) (types.Transaction, uint8, error) {
	bs, err := ex.db.HGet(tagTransaction, []byte(txid))
	if err != nil {
		return nil, 0, err
	}
	if len(bs) == 0 {
		return nil, 0, ErrNotExist
	}
	item := &Transaction{}
	if err := encoding.Unmarshal(bs, &item); err != nil {
		return nil, 0, err
	}
	fc := encoding.Factory("transaction")
	t, err := fc.Create(item.Type)
	if err != nil {
		return nil, 0, err
	}
	if err := encoding.Unmarshal(item.Data, &t); err != nil {
		return nil, 0, err
	}
	return t.(types.Transaction), item.Result, nil
}

func (ex *Explorer) getTransactionFromTail(addr common.Address, offset int32, count int32) ([]string, []types.Transaction, []uint8, error) {
	s, err := ex.db.LLen(toTransactionListKey(addr))
	if err != nil {
		return nil, nil, nil, err
	}
	from := int32(s) - count - offset
	if from < 0 {
		from = 0
	}
	to := int32(s) - offset
	if to < 0 {
		to = 0
	}
	return ex.getTransactions(addr, from, to, true)
}

func (ex *Explorer) getTransactions(addr common.Address, from int32, to int32, reverse bool) ([]string, []types.Transaction, []uint8, error) {
	values, err := ex.db.LRange(toTransactionListKey(addr), from, to)
	if err != nil {
		return nil, nil, nil, err
	}
	txids := make([]string, len(values))
	txs := make([]types.Transaction, len(values))
	results := make([]uint8, len(values))
	for i, bs := range values {
		txid := string(bs)
		tx, result, err := ex.getTransaction(txid)
		if err != nil {
			return nil, nil, nil, err
		}
		if reverse {
			txids[len(values)-1-i] = txid
			txs[len(values)-1-i] = tx
			results[len(values)-1-i] = result
		} else {
			txids[i] = txid
			txs[i] = tx
			results[i] = result
		}
	}
	return txids, txs, results, nil
}

func (ex *Explorer) getTransactionTotalFromTail(offset int32, count int32) ([]string, []types.Transaction, []uint8, error) {
	s, err := ex.db.LLen(tagTransactionList)
	if err != nil {
		return nil, nil, nil, err
	}
	from := int32(s) - count - offset
	if from < 0 {
		from = 0
	}
	to := int32(s) - offset
	if to < 0 {
		to = 0
	}
	return ex.getTransactionTotal(from, to, true)
}

func (ex *Explorer) getTransactionTotal(from int32, to int32, reverse bool) ([]string, []types.Transaction, []uint8, error) {
	values, err := ex.db.LRange(tagTransactionList, from, to)
	if err != nil {
		return nil, nil, nil, err
	}
	txids := make([]string, len(values))
	txs := make([]types.Transaction, len(values))
	results := make([]uint8, len(values))
	for i, bs := range values {
		txid := string(bs)
		tx, result, err := ex.getTransaction(txid)
		if err != nil {
			return nil, nil, nil, err
		}
		if reverse {
			txids[len(values)-1-i] = txid
			txs[len(values)-1-i] = tx
			results[len(values)-1-i] = result
		} else {
			txids[i] = txid
			txs[i] = tx
			results[i] = result
		}
	}
	return txids, txs, results, nil
}

func (ex *Explorer) addTransaction(txid string, t uint16, at chain.AccountTransaction, res uint8) error {
	isGateway := false
	if res == 1 {
		if _, err := ex.db.RPush(tagTransactionList, []byte(txid)); err != nil {
			return err
		}
		if _, err := ex.db.HSet(toTransactionTypeKey(t), []byte(txid), []byte{1}); err != nil {
			return err
		}
	}
	if !isGateway {
		if _, err := ex.db.RPush(toTransactionListKey(at.From()), []byte(txid)); err != nil {
			return err
		}
	}
	data, err := encoding.Marshal(at)
	if err != nil {
		return err
	}
	bs, err := encoding.Marshal(&Transaction{
		Type:   t,
		Data:   data,
		Result: res,
	})
	if err != nil {
		return err
	}
	if _, err := ex.db.HSet(tagTransaction, []byte(txid), bs); err != nil {
		return err
	}
	return nil
}

func (ex *Explorer) getMainSummary() map[string]interface{} {
	m := map[string]interface{}{
		"height":   ex.Height(),
		"tx_count": ex.TransactionCount(),
	}
	cs := ex.cs.Candidates()
	m["formulator_count"] = len(cs)
	return m
}

func (ex *Explorer) getAccountInfo(address string) (m map[string]interface{}, addr common.Address, err error) {
	addr, err = common.ParseAddress(address)
	if err != nil {
		return
	}

	ctw := ex.cn.NewLoaderWrapper(1)
	acc, err := ctw.Account(addr)
	if err != nil {
		return
	}
	accType := "single"
	if a, is := acc.(*formulator.FormulatorAccount); is {
		if a.FormulatorType == formulator.HyperFormulatorType {
			accType = "validator"
		} else {
			accType = "formulator"
		}
	}
	height := addr.Height()
	if height == 0 {
		height = 1
	}
	bh, err := ex.cn.Header(height)
	if err != nil {
		return
	}
	m = map[string]interface{}{
		"name":      acc.Name(),
		"address":   address,
		"type":      accType,
		"timestamp": bh.Timestamp,
	}
	return
}

func (ex *Explorer) txInfo(b *types.Block, txIndex uint16) map[string]interface{} {
	m := map[string]interface{}{}
	t := b.TransactionTypes[int(txIndex)]
	tx := b.Transactions[int(txIndex)]
	txResult := b.TransactionResults[int(txIndex)]

	fc := encoding.Factory("transaction")

	name, err := fc.TypeName(t)
	if err != nil {
		m["err"] = "지원하지 않는 transaction 입니다."
	}
	m["Type"] = name
	if txResult == 0 {
		m["Result"] = "Fail"
	} else if txResult == 1 {
		m["Result"] = "Success"
	} else {
		m["Result"] = "Unknown"
	}

	m["Block Hash"] = encoding.Hash(b.Header).String()

	m["Block Timestamp"] = int64(b.Header.Timestamp)
	m["Tx Hash"] = chain.HashTransactionByType(ex.cn.ChainID(), t, tx).String()
	m["Tx TimeStamp"] = int64(tx.Timestamp())

	return m
}

func (ex *Explorer) txDetailMap(t types.Transaction) ([]byte, error) {
	txbs, err := t.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return txbs, nil
}
