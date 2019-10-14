package p2p

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/fletaio/fleta/common/queue"
	"github.com/fletaio/fleta/common/util"
	"github.com/gorilla/websocket"
)

// WebsocketPeer manages send and recv of the connection
type WebsocketPeer struct {
	sync.Mutex
	conn          *websocket.Conn
	id            string
	name          string
	guessHeight   uint32
	writeQueue    *queue.Queue
	isClose       bool
	connectedTime int64
	pingCount     uint64
}

// NewWebsocketPeer returns a WebsocketPeer
func NewWebsocketPeer(conn *websocket.Conn, ID string, Name string, connectedTime int64) *WebsocketPeer {
	if len(Name) == 0 {
		Name = ID
	}
	p := &WebsocketPeer{
		conn:          conn,
		id:            ID,
		name:          Name,
		writeQueue:    queue.NewQueue(),
		connectedTime: connectedTime,
	}
	conn.EnableWriteCompression(false)
	conn.SetPingHandler(func(appData string) error {
		atomic.StoreUint64(&p.pingCount, 0)
		return nil
	})

	go func() {
		defer p.Close()

		pingCountLimit := uint64(3)
		pingTicker := time.NewTicker(10 * time.Second)
		for !p.isClose {
			select {
			case <-pingTicker.C:
				if err := p.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
					return
				}
				if atomic.AddUint64(&p.pingCount, 1) > pingCountLimit {
					return
				}
			default:
				v := p.writeQueue.Pop()
				if v == nil {
					time.Sleep(50 * time.Millisecond)
					continue
				}
				if err := p.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
					return
				}
				if err := p.conn.WriteMessage(websocket.BinaryMessage, v.([]byte)); err != nil {
					return
				}
			}
		}
	}()
	return p
}

// ID returns the id of the peer
func (p *WebsocketPeer) ID() string {
	return p.id
}

// Name returns the name of the peer
func (p *WebsocketPeer) Name() string {
	return p.name
}

// Close closes WebsocketPeer
func (p *WebsocketPeer) Close() {
	p.isClose = true
	p.conn.Close()
}

// ReadPacket returns a packet data
func (p *WebsocketPeer) ReadPacket() (uint16, bool, []byte, error) {
	_, rb, err := p.conn.ReadMessage()
	if err != nil {
		return 0, false, nil, err
	}
	if len(rb) < 7 {
		return 0, false, nil, ErrInvalidLength
	}

	t := util.BytesToUint16(rb)
	cps := rb[2:3]
	Len := util.BytesToUint32(rb[3:])
	if Len == 0 {
		return 0, false, nil, ErrUnknownMessage
	} else if len(rb) != 7+int(Len) {
		return 0, false, nil, ErrInvalidLength
	} else {
		return t, cps[0] == 1, rb, nil
	}
}

// SendPacket sends packet to the WebsocketPeer
func (p *WebsocketPeer) SendPacket(bs []byte) {
	p.writeQueue.Push(bs)
}

// UpdateGuessHeight updates the guess height of the WebsocketPeer
func (p *WebsocketPeer) UpdateGuessHeight(height uint32) {
	p.Lock()
	defer p.Unlock()

	p.guessHeight = height
}

// GuessHeight updates the guess height of the WebsocketPeer
func (p *WebsocketPeer) GuessHeight() uint32 {
	return p.guessHeight
}

// ConnectedTime returns peer connected time
func (p *WebsocketPeer) ConnectedTime() int64 {
	return p.connectedTime
}
