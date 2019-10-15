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
