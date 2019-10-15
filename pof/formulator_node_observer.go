package pof

import (
	"bytes"
	"sync"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/p2p"
	"github.com/fletaio/fleta/service/p2p/peer"
)

// OnObserverConnected is called after a new observer peer is connected
func (fr *FormulatorNode) OnObserverConnected(p peer.Peer) {
	fr.statusLock.Lock()
	fr.obStatusMap[p.ID()] = &p2p.Status{}
	fr.statusLock.Unlock()

	cp := fr.cs.cn.Provider()
	height, lastHash, _ := cp.LastStatus()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	p.SendPacket(p2p.MessageToPacket(nm))
}

// OnObserverDisconnected is called when the observer peer is disconnected
func (fr *FormulatorNode) OnObserverDisconnected(p peer.Peer) {
	fr.statusLock.Lock()
	delete(fr.obStatusMap, p.ID())
	fr.statusLock.Unlock()
	fr.requestTimer.RemovesByValue(p.ID())
	go fr.tryRequestNext()
}

func (fr *FormulatorNode) onObserverRecv(p peer.Peer, bs []byte) error {
	m, err := p2p.PacketToMessage(bs)
	if err != nil {
		return err
	}
	if err := fr.handleObserverMessage(p, m, 0); err != nil {
		//rlog.Println(err)
		return nil
	}
	return nil
}

func (fr *FormulatorNode) handleObserverMessage(p peer.Peer, m interface{}, RetryCount int) error {
	cp := fr.cs.cn.Provider()

	switch msg := m.(type) {
	case *BlockReqMessage:
		rlog.Println("Formulator", fr.Config.Formulator.String(), "BlockReqMessage", msg.TargetHeight)

		fr.Lock()
		defer fr.Unlock()

		Height := cp.Height()
		if msg.TargetHeight <= fr.lastGenHeight && fr.lastGenTime+int64(30*time.Second) > time.Now().UnixNano() {
			return nil
		}
		if fr.lastReqMessage != nil {
			if msg.TargetHeight <= fr.lastReqMessage.TargetHeight {
				return nil
			}
		}
		if msg.TargetHeight <= Height {
			return nil
		}
		if msg.TargetHeight > Height+1 {
			if RetryCount >= 10 {
				return nil
			}

			if RetryCount == 0 {
				Count := uint8(msg.TargetHeight - Height - 1)
				if Count > 10 {
					Count = 10
				}

				sm := &p2p.RequestMessage{
					Height: Height + 1,
					Count:  Count,
				}
				p.SendPacket(p2p.MessageToPacket(sm))
			}
			go func() {
				time.Sleep(50 * time.Millisecond)
				fr.handleObserverMessage(p, m, RetryCount+1)
			}()
			return nil
		}

		Top, err := fr.cs.rt.TopRank(int(msg.TimeoutCount))
		if err != nil {
			return err
		}
		if msg.Formulator != Top.Address {
			return ErrInvalidRequest
		}
		if msg.Formulator != fr.Config.Formulator {
			return ErrInvalidRequest
		}
		if msg.FormulatorPublicHash != common.NewPublicHash(fr.key.PublicKey()) {
			return ErrInvalidRequest
		}
		if msg.PrevHash != cp.LastHash() {
			return ErrInvalidRequest
		}
		if msg.TargetHeight != Height+1 {
			return ErrInvalidRequest
		}
		fr.lastReqMessage = msg

		var wg sync.WaitGroup
		wg.Add(1)
		go func(req *BlockReqMessage) error {
			wg.Done()

			fr.genLock.Lock()
			defer fr.genLock.Unlock()

			fr.Lock()
			defer fr.Unlock()

			return fr.genBlock(p, req)
		}(msg)
		wg.Wait()
		return nil
	case *BlockObSignMessage:
		rlog.Println("Formulator", fr.Config.Formulator.String(), "BlockObSignMessage", msg.TargetHeight)

		fr.Lock()
		defer fr.Unlock()

		TargetHeight := fr.cs.cn.Provider().Height() + 1
		if msg.TargetHeight < TargetHeight {
			return nil
		}
		if msg.TargetHeight >= fr.lastReqMessage.TargetHeight+10 {
			return ErrInvalidRequest
		}
		fr.lastObSignMessageMap[msg.TargetHeight] = msg

		for len(fr.lastGenMessages) > 0 {
			GenMessage := fr.lastGenMessages[0]
			if GenMessage.Block.Header.Height < TargetHeight {
				if len(fr.lastGenMessages) > 1 {
					fr.lastGenMessages = fr.lastGenMessages[1:]
					fr.lastContextes = fr.lastContextes[1:]
				} else {
					fr.lastGenMessages = []*BlockGenMessage{}
					fr.lastContextes = []*types.Context{}
				}
				continue
			}
			if GenMessage.Block.Header.Height > TargetHeight {
				break
			}
			sm, has := fr.lastObSignMessageMap[GenMessage.Block.Header.Height]
			if !has {
				break
			}
			if GenMessage.Block.Header.Height == sm.TargetHeight {
				ctx := fr.lastContextes[0]

				if sm.BlockSign.HeaderHash != encoding.Hash(GenMessage.Block.Header) {
					return ErrInvalidRequest
				}

				b := &types.Block{
					Header:                GenMessage.Block.Header,
					TransactionTypes:      GenMessage.Block.TransactionTypes,
					Transactions:          GenMessage.Block.Transactions,
					TransactionSignatures: GenMessage.Block.TransactionSignatures,
					TransactionResults:    GenMessage.Block.TransactionResults,
					Signatures:            append([]common.Signature{GenMessage.GeneratorSignature}, sm.ObserverSignatures...),
				}
				if err := fr.cs.ct.ConnectBlockWithContext(b, ctx); err != nil {
					return err
				}
				fr.broadcastStatus()
				fr.cleanPool(b)
				rlog.Println("Formulator", fr.Config.Formulator.String(), "BlockConnected", b.Header.Generator.String(), b.Header.Height, len(b.Transactions))

				fr.statusLock.Lock()
				if status, has := fr.obStatusMap[p.ID()]; has {
					if status.Height < GenMessage.Block.Header.Height {
						status.Height = GenMessage.Block.Header.Height
					}
				}
				fr.statusLock.Unlock()

				if len(fr.lastGenMessages) > 1 {
					fr.lastGenMessages = fr.lastGenMessages[1:]
					fr.lastContextes = fr.lastContextes[1:]
				} else {
					fr.lastGenMessages = []*BlockGenMessage{}
					fr.lastContextes = []*types.Context{}
				}
			}
		}
		return nil
	case *p2p.BlockMessage:
		for _, b := range msg.Blocks {
			if err := fr.addBlock(b); err != nil {
				if err == chain.ErrFoundForkedBlock {
					panic(err)
				}
				return err
			}
		}

		if len(msg.Blocks) > 0 {
			fr.statusLock.Lock()
			if status, has := fr.obStatusMap[p.ID()]; has {
				lastHeight := msg.Blocks[len(msg.Blocks)-1].Header.Height
				if status.Height < lastHeight {
					status.Height = lastHeight
				}
			}
			fr.statusLock.Unlock()

			fr.tryRequestNext()
		}
		return nil
	case *p2p.StatusMessage:
		fr.statusLock.Lock()
		if status, has := fr.obStatusMap[p.ID()]; has {
			if status.Height < msg.Height {
				status.Height = msg.Height
			}
		}
		fr.statusLock.Unlock()

		TargetHeight := cp.Height() + 1
		for TargetHeight <= msg.Height {
			if !fr.requestTimer.Exist(TargetHeight) {
				if fr.blockQ.Find(uint64(TargetHeight)) == nil {
					sm := &p2p.RequestMessage{
						Height: TargetHeight,
					}
					p.SendPacket(p2p.MessageToPacket(sm))
					fr.requestTimer.Add(TargetHeight, 2*time.Second, p.ID())
				}
			}
			TargetHeight++
		}
		return nil
	case *p2p.TransactionMessage:
		TxHash := chain.HashTransactionByType(fr.cs.cn.Provider().ChainID(), msg.TxType, msg.Tx)
		fr.txWaitQ.Push(TxHash, &p2p.TxMsgItem{
			TxHash:  TxHash,
			Message: msg,
		})
		return nil
	default:
		panic(p2p.ErrUnknownMessage) //TEMP
		return p2p.ErrUnknownMessage
	}
}

func (fr *FormulatorNode) tryRequestNext() {
	fr.requestLock.Lock()
	defer fr.requestLock.Unlock()

	TargetHeight := fr.cs.cn.Provider().Height() + 1
	if !fr.requestTimer.Exist(TargetHeight) {
		if fr.blockQ.Find(uint64(TargetHeight)) == nil {
			fr.statusLock.Lock()
			var TargetPubHash string
			for pubhash, status := range fr.obStatusMap {
				if TargetHeight <= status.Height {
					TargetPubHash = pubhash
					break
				}
			}
			fr.statusLock.Unlock()

			if len(TargetPubHash) > 0 {
				fr.sendRequestBlockTo(TargetPubHash, TargetHeight, 1)
			}
		}
	}
}

func (fr *FormulatorNode) genBlock(p peer.Peer, msg *BlockReqMessage) error {
	cp := fr.cs.cn.Provider()

	fr.lastGenMessages = []*BlockGenMessage{}
	fr.lastObSignMessageMap = map[uint32]*BlockObSignMessage{}
	fr.lastContextes = []*types.Context{}

	start := time.Now().UnixNano()
	Now := uint64(time.Now().UnixNano())
	StartBlockTime := Now
	bNoDelay := false

	RemainBlocks := fr.cs.maxBlocksPerFormulator
	if msg.TimeoutCount == 0 {
		RemainBlocks = fr.cs.maxBlocksPerFormulator - fr.cs.blocksBySameFormulator
	}

	LastTimestamp := cp.LastTimestamp()
	if StartBlockTime < LastTimestamp {
		StartBlockTime = LastTimestamp + uint64(time.Millisecond)
	} else if StartBlockTime > LastTimestamp+uint64(RemainBlocks)*uint64(500*time.Millisecond) {
		bNoDelay = true
	}

	var lastHeader *types.Header
	ctx := fr.cs.ct.NewContext()
	for i := uint32(0); i < RemainBlocks; i++ {
		var TimeoutCount uint32
		if i == 0 {
			TimeoutCount = msg.TimeoutCount
		} else {
			ctx = ctx.NextContext(encoding.Hash(lastHeader), lastHeader.Timestamp)
		}

		Timestamp := StartBlockTime
		if bNoDelay || Timestamp > Now+uint64(3*time.Second) {
			Timestamp += uint64(i) * uint64(time.Millisecond)
		} else {
			Timestamp += uint64(i) * uint64(500*time.Millisecond)
		}
		if Timestamp <= ctx.LastTimestamp() {
			Timestamp = ctx.LastTimestamp() + 1
		}

		var buffer bytes.Buffer
		enc := encoding.NewEncoder(&buffer)
		if err := enc.EncodeUint32(TimeoutCount); err != nil {
			return err
		}
		bc := chain.NewBlockCreator(fr.cs.cn, ctx, msg.Formulator, buffer.Bytes())
		if err := bc.Init(); err != nil {
			return err
		}

		timer := time.NewTimer(200 * time.Millisecond)

		rlog.Println("Formulator", fr.Config.Formulator.String(), "BlockGenBegin", msg.TargetHeight)

		fr.txpool.Lock() // Prevent delaying from TxPool.Push
		Count := 0
	TxLoop:
		for {
			select {
			case <-timer.C:
				break TxLoop
			default:
				sn := ctx.Snapshot()
				item := fr.txpool.UnsafePop(ctx)
				ctx.Revert(sn)
				if item == nil {
					break TxLoop
				}
				if err := bc.UnsafeAddTx(fr.Config.Formulator, item.TxType, item.TxHash, item.Transaction, item.Signatures, item.Signers); err != nil {
					rlog.Println(err)
					continue
				}
				Count++
				if Count > fr.Config.MaxTransactionsPerBlock {
					break TxLoop
				}
			}
		}
		fr.txpool.Unlock() // Prevent delaying from TxPool.Push

		b, err := bc.Finalize(Timestamp)
		if err != nil {
			return err
		}

		sm := &BlockGenMessage{
			Block: b,
		}
		lastHeader = &b.Header

		if sig, err := fr.key.Sign(encoding.Hash(b.Header)); err != nil {
			return err
		} else {
			sm.GeneratorSignature = sig
		}
		p.SendPacket(p2p.MessageToPacket(sm))

		rlog.Println("Formulator", fr.Config.Formulator.String(), "BlockGenMessage", sm.Block.Header.Height, len(sm.Block.Transactions))

		fr.lastGenMessages = append(fr.lastGenMessages, sm)
		fr.lastContextes = append(fr.lastContextes, ctx)
		fr.lastGenHeight = ctx.TargetHeight()
		fr.lastGenTime = time.Now().UnixNano()

		ExpectedTime := time.Duration(i+1) * 500 * time.Millisecond
		if i >= 7 {
			ExpectedTime = 3500*time.Millisecond + time.Duration(i-7+1)*200*time.Millisecond
		}
		PastTime := time.Duration(time.Now().UnixNano() - start)
		if !bNoDelay && ExpectedTime > PastTime {
			fr.Unlock()
			time.Sleep(ExpectedTime - PastTime)
			fr.Lock()
		}
	}
	return nil
}
