package pof

import (
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/service/p2p"
)

func (ob *ObserverNode) sendStatusTo(TargetPubHash common.PublicHash) error {
	if TargetPubHash == ob.myPublicHash {
		return nil
	}

	cp := ob.cs.cn.Provider()
	height, lastHash, _ := cp.LastStatus()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	ob.ms.SendTo(TargetPubHash, nm)
	return nil
}

func (ob *ObserverNode) broadcastStatus() error {
	cp := ob.cs.cn.Provider()
	height, lastHash, _ := cp.LastStatus()
	nm := &p2p.StatusMessage{
		Version:  cp.Version(),
		Height:   height,
		LastHash: lastHash,
	}
	ob.ms.BroadcastMessage(nm)
	return nil
}

func (ob *ObserverNode) sendRequestBlockTo(TargetPubHash common.PublicHash, Height uint32, Count uint8) error {
	if TargetPubHash == ob.myPublicHash {
		return nil
	}

	nm := &p2p.RequestMessage{
		Height: Height,
		Count:  Count,
	}
	ob.ms.SendTo(TargetPubHash, nm)
	for i := uint32(0); i < uint32(Count); i++ {
		ob.requestTimer.Add(Height+i, 2*time.Second, string(TargetPubHash[:]))
	}
	return nil
}
