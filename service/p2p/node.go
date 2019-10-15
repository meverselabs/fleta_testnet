package p2p

import (
	"bytes"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/bluele/gcache"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/txpool"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/p2p/peer"
)

// Node receives a block by the consensus
type Node struct {
	sync.Mutex
	key          key.Key
	ms           *NodeMesh
	cn           *chain.Chain
	statusLock   sync.Mutex
	myPublicHash common.PublicHash
	requestTimer *RequestTimer
	requestLock  sync.RWMutex
	blockQ       *queue.SortedQueue
	statusMap    map[string]*Status
	txpool       *txpool.TransactionPool
	txQ          *queue.ExpireQueue
	txWaitQ      *queue.LinkedQueue
	recvQueues   []*queue.Queue
	sendQueues   []*queue.Queue
	singleCache  gcache.Cache
	batchCache   gcache.Cache
	isRunning    bool
	closeLock    sync.RWMutex
	isClose      bool
}

// NewNode returns a Node
func NewNode(key key.Key, SeedNodeMap map[common.PublicHash]string, cn *chain.Chain, peerStorePath string) *Node {
	nd := &Node{
		key:          key,
		cn:           cn,
		myPublicHash: common.NewPublicHash(key.PublicKey()),
		blockQ:       queue.NewSortedQueue(),
		statusMap:    map[string]*Status{},
		txpool:       txpool.NewTransactionPool(),
		txQ:          queue.NewExpireQueue(),
		txWaitQ:      queue.NewLinkedQueue(),
		recvQueues: []*queue.Queue{
			queue.NewQueue(), //block
			queue.NewQueue(), //tx
			queue.NewQueue(), //peer
		},
		sendQueues: []*queue.Queue{
			queue.NewQueue(), //block
			queue.NewQueue(), //tx
			queue.NewQueue(), //peer
		},
		singleCache: gcache.New(500).LRU().Build(),
		batchCache:  gcache.New(500).LRU().Build(),
	}
	nd.ms = NewNodeMesh(cn.Provider().ChainID(), key, SeedNodeMap, nd, peerStorePath)
	nd.requestTimer = NewRequestTimer(nd)
	nd.txQ.AddGroup(60 * time.Second)
	nd.txQ.AddGroup(600 * time.Second)
	nd.txQ.AddGroup(3600 * time.Second)
	nd.txQ.AddHandler(nd)
	rlog.SetRLogAddress("nd:" + nd.myPublicHash.String())
	return nd
}

// Init initializes node
func (nd *Node) Init() error {
	fc := encoding.Factory("message")
	fc.Register(PingMessageType, &PingMessage{})
	fc.Register(StatusMessageType, &StatusMessage{})
	fc.Register(RequestMessageType, &RequestMessage{})
	fc.Register(BlockMessageType, &BlockMessage{})
	fc.Register(TransactionMessageType, &TransactionMessage{})
	fc.Register(PeerListMessageType, &PeerListMessage{})
	fc.Register(RequestPeerListMessageType, &RequestPeerListMessage{})
	return nil
}

// Close terminates the node
func (nd *Node) Close() {
	nd.closeLock.Lock()
	defer nd.closeLock.Unlock()

	nd.Lock()
	defer nd.Unlock()

	nd.isClose = true
	nd.cn.Close()
}

// OnItemExpired is called when the item is expired
func (nd *Node) OnItemExpired(Interval time.Duration, Key string, Item interface{}, IsLast bool) {
	msg := Item.(*TransactionMessage)
	nd.limitCastMessage(1, msg)
	if IsLast {
		var TxHash hash.Hash256
		copy(TxHash[:], []byte(Key))
		nd.txpool.Remove(TxHash, msg.Tx)
	}
}

// Run starts the node
func (nd *Node) Run(BindAddress string) {
	nd.Lock()
	if nd.isRunning {
		nd.Unlock()
		return
	}
	nd.isRunning = true
	nd.Unlock()

	go nd.ms.Run(BindAddress)
	go nd.requestTimer.Run()

	WorkerCount := runtime.NumCPU() - 1
	if WorkerCount < 1 {
		WorkerCount = 1
	}
	for i := 0; i < WorkerCount; i++ {
		go func() {
			for !nd.isClose {
				Count := 0
				for !nd.isClose {
					v := nd.txWaitQ.Pop()
					if v == nil {
						break
					}
					item := v.(*TxMsgItem)
					if err := nd.addTx(item.TxHash, item.Message.TxType, item.Message.Tx, item.Message.Sigs); err != nil {
						if err != ErrInvalidUTXO && err != txpool.ErrExistTransaction && err != txpool.ErrTooFarSeq && err != txpool.ErrPastSeq {
							rlog.Println("TransactionError", chain.HashTransactionByType(nd.cn.Provider().ChainID(), item.Message.TxType, item.Message.Tx).String(), err.Error())
							if len(item.PeerID) > 0 {
								nd.ms.RemovePeer(item.PeerID)
							}
						}
					}
					rlog.Println("TransactionAppended", chain.HashTransactionByType(nd.cn.Provider().ChainID(), item.Message.TxType, item.Message.Tx).String())

					if len(item.PeerID) > 0 {
						var SenderPublicHash common.PublicHash
						copy(SenderPublicHash[:], []byte(item.PeerID))
						nd.exceptLimitCastMessage(1, SenderPublicHash, item.Message)
					} else {
						nd.limitCastMessage(1, item.Message)
					}

					Count++
					if Count > 500 {
						break
					}
				}
				time.Sleep(50 * time.Millisecond)
			}
		}()
	}

	go func() {
		for !nd.isClose {
			hasMessage := false
			for !nd.isClose {
				for _, q := range nd.recvQueues {
					v := q.Pop()
					if v == nil {
						continue
					}
					hasMessage = true
					item := v.(*RecvMessageItem)
					m, err := PacketToMessage(item.Packet)
					if err != nil {
						log.Println("PacketToMessage", err)
						nd.ms.RemovePeer(item.PeerID)
						break
					}
					if err := nd.handlePeerMessage(item.PeerID, m); err != nil {
						log.Println("handlePeerMessage", err)
						nd.ms.RemovePeer(item.PeerID)
						break
					}
				}
				if !hasMessage {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	go func() {
		for !nd.isClose {
			hasMessage := false
			for !nd.isClose {
				for _, q := range nd.sendQueues {
					v := q.Pop()
					if v == nil {
						continue
					}
					hasMessage = true
					item := v.(*SendMessageItem)
					if len(item.Packet) > 0 {
						nd.ms.SendTo(item.Target, item.Packet)
					} else {
						bs := MessageToPacket(item.Message)

						var EmptyHash common.PublicHash
						if bytes.Equal(item.Target[:], EmptyHash[:]) {
							if item.Limit > 0 {
								nd.ms.ExceptCastLimit("", bs, item.Limit)
							} else {
								nd.ms.BroadcastMessage(bs)
							}
						} else {
							if item.Limit > 0 {
								nd.ms.ExceptCastLimit(string(item.Target[:]), bs, item.Limit)
							} else {
								nd.ms.SendTo(item.Target, bs)
							}
						}
					}
					break
				}
				if !hasMessage {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	blockTimer := time.NewTimer(time.Millisecond)
	blockRequestTimer := time.NewTimer(time.Millisecond)
	for !nd.isClose {
		select {
		case <-blockTimer.C:
			nd.Lock()
			hasItem := false
			TargetHeight := uint64(nd.cn.Provider().Height() + 1)
			Count := 0
			item := nd.blockQ.PopUntil(TargetHeight)
			for item != nil {
				b := item.(*types.Block)
				if err := nd.cn.ConnectBlock(b); err != nil {
					rlog.Println(err)
					panic(err)
					break
				}
				rlog.Println("Node", nd.myPublicHash.String(), nd.cn.Provider().Height(), "BlockConnected", b.Header.Generator.String(), b.Header.Height)
				TargetHeight++
				Count++
				if Count > 10 {
					break
				}
				item = nd.blockQ.PopUntil(TargetHeight)
				hasItem = true
			}
			nd.Unlock()

			if hasItem {
				nd.broadcastStatus()
				nd.tryRequestBlocks()
			}

			blockTimer.Reset(50 * time.Millisecond)
		case <-blockRequestTimer.C:
			nd.tryRequestBlocks()
			blockRequestTimer.Reset(500 * time.Millisecond)
		}
	}
}

// OnTimerExpired called when rquest expired
func (nd *Node) OnTimerExpired(height uint32, value string) {
	nd.tryRequestBlocks()
}

// OnConnected called when peer connected
func (nd *Node) OnConnected(p peer.Peer) {
	nd.statusLock.Lock()
	nd.statusMap[p.ID()] = &Status{}
	nd.statusLock.Unlock()

	var SenderPublicHash common.PublicHash
	copy(SenderPublicHash[:], []byte(p.ID()))
	nd.sendStatusTo(SenderPublicHash)
}

// OnDisconnected called when peer disconnected
func (nd *Node) OnDisconnected(p peer.Peer) {
	nd.statusLock.Lock()
	delete(nd.statusMap, p.ID())
	nd.statusLock.Unlock()

	nd.requestTimer.RemovesByValue(p.ID())
	go nd.tryRequestBlocks()
}

// OnRecv called when message received
func (nd *Node) OnRecv(p peer.Peer, bs []byte) error {
	item := &RecvMessageItem{
		PeerID: p.ID(),
		Packet: bs,
	}
	t := PacketMessageType(bs)
	switch t {
	case RequestMessageType:
		nd.recvQueues[0].Push(item)
	case StatusMessageType:
		nd.recvQueues[0].Push(item)
	case BlockMessageType:
		nd.recvQueues[0].Push(item)
	case TransactionMessageType:
		nd.recvQueues[1].Push(item)
	case PeerListMessageType:
		nd.recvQueues[2].Push(item)
	case RequestPeerListMessageType:
		nd.recvQueues[2].Push(item)
	}
	return nil
}

func (nd *Node) handlePeerMessage(ID string, m interface{}) error {
	var SenderPublicHash common.PublicHash
	copy(SenderPublicHash[:], []byte(ID))

	switch msg := m.(type) {
	case *RequestMessage:
		nd.statusLock.Lock()
		status, has := nd.statusMap[ID]
		nd.statusLock.Unlock()
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
		Height := nd.cn.Provider().Height()
		if msg.Height > Height {
			return nil
		}
		var bs []byte
		if msg.Height%10 == 0 && msg.Count == 10 && msg.Height+uint32(msg.Count) <= Height {
			value, err := nd.batchCache.Get(msg.Height)
			if err != nil {
				list := make([]*types.Block, 0, 10)
				for i := uint32(0); i < uint32(msg.Count); i++ {
					if msg.Height+i > Height {
						break
					}
					b, err := nd.cn.Provider().Block(msg.Height + i)
					if err != nil {
						return err
					}
					list = append(list, b)
				}
				sm := &BlockMessage{
					Blocks: list,
				}
				bs = MessageToPacket(sm)
				nd.batchCache.Set(msg.Height, bs)
			} else {
				bs = value.([]byte)
			}
		} else if msg.Count == 1 {
			value, err := nd.singleCache.Get(msg.Height)
			if err != nil {
				b, err := nd.cn.Provider().Block(msg.Height)
				if err != nil {
					return err
				}
				sm := &BlockMessage{
					Blocks: []*types.Block{b},
				}
				bs = MessageToPacket(sm)
				nd.singleCache.Set(msg.Height, bs)
			} else {
				bs = value.([]byte)
			}
		} else {
			list := make([]*types.Block, 0, 10)
			for i := uint32(0); i < uint32(msg.Count); i++ {
				if msg.Height+i > Height {
					break
				}
				b, err := nd.cn.Provider().Block(msg.Height + i)
				if err != nil {
					return err
				}
				list = append(list, b)
			}
			sm := &BlockMessage{
				Blocks: list,
			}
			bs = MessageToPacket(sm)
		}
		nd.sendMessagePacket(0, SenderPublicHash, bs)
		return nil
	case *StatusMessage:
		nd.statusLock.Lock()
		if status, has := nd.statusMap[ID]; has {
			if status.Height < msg.Height {
				status.Height = msg.Height
			}
		}
		nd.statusLock.Unlock()

		Height := nd.cn.Provider().Height()
		if Height < msg.Height {
			enableCount := 0
			for i := Height + 1; i <= Height+10 && i <= msg.Height; i++ {
				if !nd.requestTimer.Exist(i) {
					enableCount++
				}
			}
			if Height%10 == 0 && enableCount == 10 {
				nd.sendRequestBlockTo(SenderPublicHash, Height+1, 10)
			} else {
				for i := Height + 1; i <= Height+10 && i <= msg.Height; i++ {
					if !nd.requestTimer.Exist(i) {
						nd.sendRequestBlockTo(SenderPublicHash, i, 1)
					}
				}
			}
		} else {
			h, err := nd.cn.Provider().Hash(msg.Height)
			if err != nil {
				return err
			}
			if h != msg.LastHash {
				//TODO : critical error signal
				rlog.Println(chain.ErrFoundForkedBlock, ID, h.String(), msg.LastHash.String(), msg.Height)
				nd.ms.RemovePeer(ID)
			}
		}
		return nil
	case *BlockMessage:
		for _, b := range msg.Blocks {
			if err := nd.addBlock(b); err != nil {
				if err == chain.ErrFoundForkedBlock {
					//TODO : critical error signal
					nd.ms.RemovePeer(ID)
				}
				return err
			}
		}

		if len(msg.Blocks) > 0 {
			nd.statusLock.Lock()
			if status, has := nd.statusMap[ID]; has {
				lastHeight := msg.Blocks[len(msg.Blocks)-1].Header.Height
				if status.Height < lastHeight {
					status.Height = lastHeight
				}
			}
			nd.statusLock.Unlock()
		}
		return nil
	case *TransactionMessage:
		if nd.txWaitQ.Size() > 200000 {
			return txpool.ErrTransactionPoolOverflowed
		}
		TxHash := chain.HashTransactionByType(nd.cn.Provider().ChainID(), msg.TxType, msg.Tx)
		nd.txWaitQ.Push(TxHash, &TxMsgItem{
			TxHash:  TxHash,
			Message: msg,
			PeerID:  ID,
		})
		return nil
	case *PeerListMessage:
		nd.ms.AddPeerList(msg.Ips, msg.Hashs)
		return nil
	case *RequestPeerListMessage:
		nd.ms.SendPeerList(ID)
		return nil
	default:
		panic(ErrUnknownMessage) //TEMP
		return ErrUnknownMessage
	}
	return nil
}

func (nd *Node) addBlock(b *types.Block) error {
	cp := nd.cn.Provider()
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
		if item := nd.blockQ.FindOrInsert(b, uint64(b.Header.Height)); item != nil {
			old := item.(*types.Block)
			if encoding.Hash(old.Header) != encoding.Hash(b.Header) {
				//TODO : critical error signal
				return chain.ErrFoundForkedBlock
			}
		}
	}
	return nil
}

// AddTx adds tx to txpool that only have valid signatures
func (nd *Node) AddTx(tx types.Transaction, sigs []common.Signature) error {
	fc := encoding.Factory("transaction")
	t, err := fc.TypeOf(tx)
	if err != nil {
		return err
	}
	TxHash := chain.HashTransactionByType(nd.cn.Provider().ChainID(), t, tx)
	nd.txWaitQ.Push(TxHash, &TxMsgItem{
		TxHash: TxHash,
		Message: &TransactionMessage{
			TxType: t,
			Tx:     tx,
			Sigs:   sigs,
		},
	})
	return nil
}

func (nd *Node) addTx(TxHash hash.Hash256, t uint16, tx types.Transaction, sigs []common.Signature) error {
	if nd.txpool.Size() > 65535 {
		return txpool.ErrTransactionPoolOverflowed
	}

	cp := nd.cn.Provider()
	if nd.txpool.IsExist(TxHash) {
		return txpool.ErrExistTransaction
	}
	if atx, is := tx.(chain.AccountTransaction); is {
		seq := cp.Seq(atx.From())
		if atx.Seq() <= seq {
			return txpool.ErrPastSeq
		} else if atx.Seq() > seq+100 {
			return txpool.ErrTooFarSeq
		}
	}
	signers := make([]common.PublicHash, 0, len(sigs))
	for _, sig := range sigs {
		pubkey, err := common.RecoverPubkey(TxHash, sig)
		if err != nil {
			return err
		}
		signers = append(signers, common.NewPublicHash(pubkey))
	}
	pid := uint8(t >> 8)
	p, err := nd.cn.Process(pid)
	if err != nil {
		return err
	}
	ctx := nd.cn.NewContext()
	ctw := types.NewContextWrapper(pid, ctx)
	if err := tx.Validate(p, ctw, signers); err != nil {
		return err
	}
	if err := nd.txpool.Push(nd.cn.Provider().ChainID(), t, TxHash, tx, sigs, signers); err != nil {
		return err
	}
	nd.txQ.Push(string(TxHash[:]), &TransactionMessage{
		TxType: t,
		Tx:     tx,
		Sigs:   sigs,
	})
	return nil
}

func (nd *Node) tryRequestBlocks() {
	nd.requestLock.Lock()
	defer nd.requestLock.Unlock()

	Height := nd.cn.Provider().Height()
	for q := uint32(0); q < 10; q++ {
		BaseHeight := Height + q*10

		var LimitHeight uint32
		var selectedPubHash string
		nd.statusLock.Lock()
		for pubhash, status := range nd.statusMap {
			if BaseHeight+10 <= status.Height {
				selectedPubHash = pubhash
				LimitHeight = status.Height
				break
			}
		}
		if len(selectedPubHash) == 0 {
			for pubhash, status := range nd.statusMap {
				if BaseHeight <= status.Height {
					selectedPubHash = pubhash
					LimitHeight = status.Height
					break
				}
			}
		}
		nd.statusLock.Unlock()

		if len(selectedPubHash) == 0 {
			break
		}
		enableCount := 0
		for i := BaseHeight + 1; i <= BaseHeight+10 && i <= LimitHeight; i++ {
			if !nd.requestTimer.Exist(i) {
				enableCount++
			}
		}

		var TargetPublicHash common.PublicHash
		copy(TargetPublicHash[:], []byte(selectedPubHash))
		if enableCount == 10 {
			nd.sendRequestBlockTo(TargetPublicHash, BaseHeight+1, 10)
		} else if enableCount > 0 {
			for i := BaseHeight + 1; i <= BaseHeight+10 && i <= LimitHeight; i++ {
				if !nd.requestTimer.Exist(i) {
					nd.sendRequestBlockTo(TargetPublicHash, i, 1)
				}
			}
		}
	}
}

func (nd *Node) cleanPool(b *types.Block) {
	for i, tx := range b.Transactions {
		t := b.TransactionTypes[i]
		TxHash := chain.HashTransactionByType(nd.cn.Provider().ChainID(), t, tx)
		nd.txpool.Remove(TxHash, tx)
		nd.txQ.Remove(string(TxHash[:]))
	}
}
