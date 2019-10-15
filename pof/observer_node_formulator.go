package pof

import (
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/service/p2p"
	"github.com/fletaio/fleta/service/p2p/peer"
)

// OnFormulatorConnected is called after a new formulator peer is connected
func (ob *ObserverNode) OnFormulatorConnected(p peer.Peer) {
	ob.statusLock.Lock()
	ob.statusMap[p.ID()] = &p2p.Status{}
	ob.statusLock.Unlock()

	cp := ob.cs.cn.Provider()
	height, lastHash, _ := cp.LastStatus()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	p.SendPacket(p2p.MessageToPacket(nm))
}

// OnFormulatorDisconnected is called when the formulator peer is disconnected
func (ob *ObserverNode) OnFormulatorDisconnected(p peer.Peer) {
	ob.statusLock.Lock()
	delete(ob.statusMap, p.ID())
	ob.statusLock.Unlock()
}

func (ob *ObserverNode) onFormulatorRecv(p peer.Peer, bs []byte) error {
	m, err := p2p.PacketToMessage(bs)
	if err != nil {
		return err
	}
	raw := bs

	cp := ob.cs.cn.Provider()
	switch msg := m.(type) {
	case *BlockGenMessage:
		ob.messageQueue.Push(&messageItem{
			Message: msg,
			Raw:     raw,
		})
	case *p2p.RequestMessage:
		ob.statusLock.Lock()
		status, has := ob.statusMap[p.ID()]
		ob.statusLock.Unlock()
		if has {
			if msg.Height < status.Height {
				if msg.Height+uint32(msg.Count) <= status.Height {
					return nil
				}
				msg.Height = status.Height
			}
		}

		enable := false
		hasCount := 0
		ob.statusLock.Lock()
		for _, status := range ob.statusMap {
			if status.Height >= msg.Height {
				hasCount++
				if hasCount >= 3 {
					break
				}
			}
		}
		ob.statusLock.Unlock()

		// TODO : it is top leader, only allow top
		// TODO : it is next leader, only allow next
		// TODO : it is not leader, accept 3rd-5th
		if hasCount < 3 {
			enable = true
		} else {
			ob.Lock()
			ranks, err := ob.cs.rt.RanksInMap(ob.adjustFormulatorMap(), 5)
			ob.Unlock()
			if err != nil {
				return err
			}
			rankMap := map[string]bool{}
			for _, r := range ranks {
				rankMap[string(r.Address[:])] = true
			}
			enable = rankMap[p.ID()]
		}
		if enable {
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
			p.SendPacket(p2p.MessageToPacket(sm))

			if len(list) > 0 {
				LastHeight := list[len(list)-1].Header.Height
				ob.statusLock.Lock()
				if status, has := ob.statusMap[p.ID()]; has {
					if status.Height < LastHeight {
						status.Height = LastHeight
					}
				}
				ob.statusLock.Unlock()
			}
		}
	case *p2p.StatusMessage:
		ob.statusLock.Lock()
		if status, has := ob.statusMap[p.ID()]; has {
			if status.Height < msg.Height {
				status.Height = msg.Height
			}
		}
		ob.statusLock.Unlock()

		Height := cp.Height()
		if Height >= msg.Height {
			h, err := cp.Hash(msg.Height)
			if err != nil {
				return err
			}
			if h != msg.LastHash {
				//TODO : critical error signal
				rlog.Println(chain.ErrFoundForkedBlock, p.Name(), h.String(), msg.LastHash.String(), msg.Height)
				ob.fs.RemovePeer(p.ID())
			}
		}
	default:
		panic(p2p.ErrUnknownMessage) //TEMP
		return p2p.ErrUnknownMessage
	}
	return nil
}
