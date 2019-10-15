package pof

import (
	"sync"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/apiserver"
	"github.com/fletaio/fleta/service/p2p"
	"github.com/fletaio/fleta/service/p2p/peer"
)

type messageItem struct {
	PublicHash common.PublicHash
	Message    interface{}
	Raw        []byte
}

// ObserverNode observes a block by the consensus
type ObserverNode struct {
	sync.Mutex
	key          key.Key
	ms           *ObserverNodeMesh
	fs           *FormulatorService
	cs           *Consensus
	ignoreMap    map[common.Address]int64
	myPublicHash common.PublicHash
	statusLock   sync.Mutex
	statusMap    map[string]*p2p.Status
	requestTimer *p2p.RequestTimer
	blockQ       *queue.SortedQueue
	messageQueue *queue.Queue
	isRunning    bool
	closeLock    sync.RWMutex
	isClose      bool
}

// NewObserverNode returns a ObserverNode
func NewObserverNode(key key.Key, NetAddressMap map[common.PublicHash]string, cs *Consensus) *ObserverNode {
	ob := &ObserverNode{
		key:          key,
		cs:           cs,
		ignoreMap:    map[common.Address]int64{},
		myPublicHash: common.NewPublicHash(key.PublicKey()),
		statusMap:    map[string]*p2p.Status{},
		blockQ:       queue.NewSortedQueue(),
		messageQueue: queue.NewQueue(),
	}
	ob.ms = NewObserverNodeMesh(key, NetAddressMap, ob)
	ob.fs = NewFormulatorService(ob)
	ob.requestTimer = p2p.NewRequestTimer(ob)

	rlog.SetRLogAddress("ob:" + ob.myPublicHash.String())
	return ob
}

// Init initializes observer
func (ob *ObserverNode) Init() error {
	fc := encoding.Factory("message")
	fc.Register(types.DefineHashedType("pof.RoundVoteMessage"), &RoundVoteMessage{})
	fc.Register(types.DefineHashedType("pof.RoundVoteAckMessage"), &RoundVoteAckMessage{})
	fc.Register(types.DefineHashedType("pof.RoundSetupMessage"), &RoundSetupMessage{})
	fc.Register(types.DefineHashedType("pof.BlockReqMessage"), &BlockReqMessage{})
	fc.Register(types.DefineHashedType("pof.BlockGenMessage"), &BlockGenMessage{})
	fc.Register(types.DefineHashedType("pof.BlockVoteMessage"), &BlockVoteMessage{})
	fc.Register(types.DefineHashedType("pof.BlockObSignMessage"), &BlockObSignMessage{})
	fc.Register(types.DefineHashedType("pof.BlockGenRequestMessage"), &BlockGenRequestMessage{})
	fc.Register(types.DefineHashedType("p2p.PingMessage"), &p2p.PingMessage{})
	fc.Register(types.DefineHashedType("p2p.StatusMessage"), &p2p.StatusMessage{})
	fc.Register(types.DefineHashedType("p2p.BlockMessage"), &p2p.BlockMessage{})
	fc.Register(types.DefineHashedType("p2p.RequestMessage"), &p2p.RequestMessage{})

	if s, err := ob.cs.cn.ServiceByName("fleta.apiserver"); err != nil {
	} else if as, is := s.(*apiserver.APIServer); !is {
	} else {
		js, err := as.JRPC("observer")
		if err != nil {
			return err
		}
		js.Set("formulatorMap", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			m := ob.fs.FormulatorMap()
			nm := map[string]bool{}
			for k, v := range m {
				nm[k.String()] = v
			}
			return nm, nil
		})
		js.Set("adjustFormulatorMap", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			m := ob.adjustFormulatorMap()
			nm := map[string]bool{}
			for k, v := range m {
				nm[k.String()] = v
			}
			return nm, nil
		})
	}
	return nil
}

// Close terminates the observer
func (ob *ObserverNode) Close() {
	ob.closeLock.Lock()
	defer ob.closeLock.Unlock()

	ob.Lock()
	defer ob.Unlock()

	ob.isClose = true
	ob.cs.cn.Close()
}

// Run starts the pof consensus on the observer
func (ob *ObserverNode) Run(BindObserver string, BindFormulator string) {
	ob.Lock()
	if ob.isRunning {
		ob.Unlock()
		return
	}
	ob.isRunning = true
	ob.Unlock()

	go ob.ms.Run(BindObserver)
	go ob.fs.Run(BindFormulator)
	go ob.requestTimer.Run()
}

// OnTimerExpired called when rquest expired
func (ob *ObserverNode) OnTimerExpired(height uint32, value string) {
	if height > ob.cs.cn.Provider().Height() {
		var TargetPublicHash common.PublicHash
		copy(TargetPublicHash[:], []byte(value))
		list := ob.ms.Peers()
		for _, p := range list {
			var pubhash common.PublicHash
			copy(pubhash[:], []byte(p.ID()))
			if pubhash != ob.myPublicHash && pubhash != TargetPublicHash {
				ob.sendRequestBlockTo(pubhash, height, 1)
				break
			}
		}
	}
}

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

func (ob *ObserverNode) addBlock(b *types.Block) error {
	cp := ob.cs.cn.Provider()
	if b.Header.Height <= cp.Height() {
		h, err := cp.Hash(b.Header.Height)
		if err != nil {
			return err
		}
		if h != encoding.Hash(b.Header) {
			//TODO : critical error signal
			return chain.ErrFoundForkedBlock
		}
	} else {
		if item := ob.blockQ.FindOrInsert(b, uint64(b.Header.Height)); item != nil {
			old := item.(*types.Block)
			if encoding.Hash(old.Header) != encoding.Hash(b.Header) {
				//TODO : critical error signal
				return chain.ErrFoundForkedBlock
			}
		}
	}
	return nil
}

func (ob *ObserverNode) adjustFormulatorMap() map[common.Address]bool {
	FormulatorMap := ob.fs.FormulatorMap()
	now := time.Now().UnixNano()
	for addr := range FormulatorMap {
		if now < ob.ignoreMap[addr] {
			delete(FormulatorMap, addr)
		}
	}
	return FormulatorMap
}
