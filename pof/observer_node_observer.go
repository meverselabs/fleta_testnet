package pof

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/service/p2p"
	"github.com/fletaio/fleta/service/p2p/peer"
)

func (ob *ObserverNode) onObserverRecv(p peer.Peer, bs []byte) error {
	m, err := p2p.PacketToMessage(bs)
	if err != nil {
		return err
	}

	if msg, is := m.(*BlockGenMessage); is {
		ob.messageQueue.Push(&messageItem{
			Message: msg,
		})
	} else {
		var pubhash common.PublicHash
		copy(pubhash[:], []byte(p.ID()))
		ob.messageQueue.Push(&messageItem{
			PublicHash: pubhash,
			Message:    m,
		})
	}
	return nil
}

func (ob *ObserverNode) handleObserverMessage(SenderPublicHash common.PublicHash, m interface{}, raw []byte) error {
	cp := ob.cs.cn.Provider()

	switch msg := m.(type) {
	case *p2p.RequestMessage:
		if msg.Count == 0 {
			msg.Count = 1
		}
		if msg.Count > 10 {
			msg.Count = 10
		}
		Height := cp.Height()
		if msg.Height > Height {
			return nil
		}
		list := make([]*types.Block, 0, 10)
		for i := uint32(0); i < uint32(msg.Count); i++ {
			if msg.Height+i > Height {
				break
			}
			b, err := cp.Block(msg.Height + i)
			if err != nil {
				return err
			}
			list = append(list, b)
		}
		sm := &p2p.BlockMessage{
			Blocks: list,
		}
		if err := ob.ms.SendTo(SenderPublicHash, sm); err != nil {
			return err
		}
	case *p2p.StatusMessage:
		Height := cp.Height()
		if Height < msg.Height {
			for q := uint32(0); q < 10; q++ {
				BaseHeight := Height + q*10
				if BaseHeight > msg.Height {
					break
				}
				enableCount := 0
				for i := BaseHeight + 1; i <= BaseHeight+10 && i <= msg.Height; i++ {
					if !ob.requestTimer.Exist(i) {
						enableCount++
					}
				}
				if enableCount == 10 {
					ob.sendRequestBlockTo(SenderPublicHash, BaseHeight+1, 10)
				} else if enableCount > 0 {
					for i := BaseHeight + 1; i <= BaseHeight+10 && i <= msg.Height; i++ {
						if !ob.requestTimer.Exist(i) {
							ob.sendRequestBlockTo(SenderPublicHash, i, 1)
						}
					}
				}
			}
		} else {
			h, err := cp.Hash(msg.Height)
			if err != nil {
				return err
			}
			if h != msg.LastHash {
				//TODO : critical error signal
				rlog.Println(SenderPublicHash.String(), h.String(), msg.LastHash.String(), msg.Height)
				panic(chain.ErrFoundForkedBlock)
			}
		}
	case *p2p.BlockMessage:
		for _, b := range msg.Blocks {
			if err := ob.addBlock(b); err != nil {
				if err != nil {
					panic(chain.ErrFoundForkedBlock)
				}
				return err
			}
		}
	default:
		return p2p.ErrUnknownMessage
	}
	return nil
}
