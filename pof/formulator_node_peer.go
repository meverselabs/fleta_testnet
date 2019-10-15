package pof

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/txpool"
	"github.com/fletaio/fleta/service/p2p"
	"github.com/fletaio/fleta/service/p2p/peer"
)

// OnConnected is called after a new  peer is connected
func (fr *FormulatorNode) OnConnected(p peer.Peer) {
	fr.statusLock.Lock()
	fr.statusMap[p.ID()] = &p2p.Status{}
	fr.statusLock.Unlock()
}

// OnDisconnected is called when the  peer is disconnected
func (fr *FormulatorNode) OnDisconnected(p peer.Peer) {
	fr.statusLock.Lock()
	delete(fr.statusMap, p.ID())
	fr.statusLock.Unlock()
	fr.requestNodeTimer.RemovesByValue(p.ID())
	go fr.tryRequestBlocks()
}

// OnRecv called when message received
func (fr *FormulatorNode) OnRecv(p peer.Peer, bs []byte) error {
	item := &p2p.RecvMessageItem{
		PeerID: p.ID(),
		Packet: bs,
	}
	t := p2p.PacketMessageType(bs)
	switch t {
	case p2p.RequestMessageType:
		fr.recvQueues[0].Push(item)
	case p2p.StatusMessageType:
		fr.recvQueues[0].Push(item)
	case p2p.BlockMessageType:
		fr.recvQueues[0].Push(item)
	case p2p.TransactionMessageType:
		fr.recvQueues[1].Push(item)
	case p2p.PeerListMessageType:
		fr.recvQueues[2].Push(item)
	case p2p.RequestPeerListMessageType:
		fr.recvQueues[2].Push(item)
	}
	return nil
}

func (fr *FormulatorNode) handlePeerMessage(ID string, m interface{}) error {
	var SenderPublicHash common.PublicHash
	copy(SenderPublicHash[:], []byte(ID))

	switch msg := m.(type) {
	case *p2p.RequestMessage:
		fr.statusLock.Lock()
		status, has := fr.statusMap[ID]
		fr.statusLock.Unlock()
		if has {
			if msg.Height < status.Height {
				if msg.Height+uint32(msg.Count) <= status.Height {
					return nil
				}
				msg.Height = status.Height
			}
		}

		if msg.Count == 0 {
			msg.Count = 1
		}
		if msg.Count > 10 {
			msg.Count = 10
		}
		Height := fr.cs.cn.Provider().Height()
		if msg.Height > Height {
			return nil
		}
		bs, err := p2p.BlockPacketWithCache(msg, fr.cs.cn.Provider(), fr.batchCache, fr.singleCache)
		if err != nil {
			return err
		}
		fr.sendMessage(0, SenderPublicHash, bs)
	case *p2p.StatusMessage:
		fr.statusLock.Lock()
		if status, has := fr.statusMap[ID]; has {
			if status.Height < msg.Height {
				status.Height = msg.Height
			}
		}
		fr.statusLock.Unlock()

		Height := fr.cs.cn.Provider().Height()
		if Height < msg.Height {
			enableCount := 0
			for i := Height + 1; i <= Height+10 && i <= msg.Height; i++ {
				if !fr.requestTimer.Exist(i) {
					if !fr.requestNodeTimer.Exist(i) {
						enableCount++
					}
				}
			}
			if Height%10 == 0 && enableCount == 10 {
				fr.sendRequestBlockToNode(SenderPublicHash, Height+1, 1)
			} else {
				for i := Height + 1; i <= Height+10 && i <= msg.Height; i++ {
					if !fr.requestTimer.Exist(i) {
						if !fr.requestNodeTimer.Exist(i) {
							fr.sendRequestBlockToNode(SenderPublicHash, i, 1)
						}
					}
				}
			}
		} else {
			h, err := fr.cs.cn.Provider().Hash(msg.Height)
			if err != nil {
				return err
			}
			if h != msg.LastHash {
				//TODO : critical error signal
				rlog.Println(chain.ErrFoundForkedBlock, ID, h.String(), msg.LastHash.String(), msg.Height)
				fr.nm.RemovePeer(ID)
			}
		}
	case *p2p.BlockMessage:
		for _, b := range msg.Blocks {
			if err := fr.addBlock(b); err != nil {
				if err == chain.ErrFoundForkedBlock {
					fr.nm.RemovePeer(ID)
				}
				return err
			}
		}

		if len(msg.Blocks) > 0 {
			fr.statusLock.Lock()
			if status, has := fr.statusMap[ID]; has {
				lastHeight := msg.Blocks[len(msg.Blocks)-1].Header.Height
				if status.Height < lastHeight {
					status.Height = lastHeight
				}
			}
			fr.statusLock.Unlock()
		}
	case *p2p.TransactionMessage:
		if fr.txWaitQ.Size() > 200000 {
			return txpool.ErrTransactionPoolOverflowed
		}
		TxHash := chain.HashTransactionByType(fr.cs.cn.Provider().ChainID(), msg.TxType, msg.Tx)
		fr.txWaitQ.Push(TxHash, &p2p.TxMsgItem{
			TxHash:  TxHash,
			Message: msg,
			PeerID:  ID,
		})
		return nil
	case *p2p.PeerListMessage:
		fr.nm.AddPeerList(msg.Ips, msg.Hashs)
		return nil
	case *p2p.RequestPeerListMessage:
		fr.nm.SendPeerList(ID)
		return nil
	default:
		panic(p2p.ErrUnknownMessage) //TEMP
		return p2p.ErrUnknownMessage
	}
	return nil
}

func (fr *FormulatorNode) tryRequestBlocks() {
	fr.requestLock.Lock()
	defer fr.requestLock.Unlock()

	Height := fr.cs.cn.Provider().Height()
	for q := uint32(0); q < 10; q++ {
		BaseHeight := Height + q*10

		var LimitHeight uint32
		var selectedPubHash string
		fr.statusLock.Lock()
		for pubhash, status := range fr.statusMap {
			if BaseHeight+10 <= status.Height {
				selectedPubHash = pubhash
				LimitHeight = status.Height
				break
			}
		}
		if len(selectedPubHash) == 0 {
			for pubhash, status := range fr.statusMap {
				if BaseHeight <= status.Height {
					selectedPubHash = pubhash
					LimitHeight = status.Height
					break
				}
			}
		}
		fr.statusLock.Unlock()

		if len(selectedPubHash) == 0 {
			break
		}
		enableCount := 0
		for i := BaseHeight + 1; i <= BaseHeight+10 && i <= LimitHeight; i++ {
			if !fr.requestTimer.Exist(i) {
				if !fr.requestNodeTimer.Exist(i) {
					enableCount++
				}
			}
		}

		var TargetPublicHash common.PublicHash
		copy(TargetPublicHash[:], []byte(selectedPubHash))
		if enableCount == 10 {
			fr.sendRequestBlockToNode(TargetPublicHash, BaseHeight+1, 10)
		} else if enableCount > 0 {
			for i := BaseHeight + 1; i <= BaseHeight+10 && i <= LimitHeight; i++ {
				if !fr.requestTimer.Exist(i) {
					if !fr.requestNodeTimer.Exist(i) {
						fr.sendRequestBlockToNode(TargetPublicHash, i, 1)
					}
				}
			}
		}
	}
}
