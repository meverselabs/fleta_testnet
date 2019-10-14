package p2p

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/fletaio/fleta/common/util"
	"github.com/fletaio/fleta/core/types"
)

// TCPPeer manages send and recv of the connection
type TCPPeer struct {
	conn          net.Conn
	id            string
	name          string
	isClose       bool
	connectedTime int64
	pingCount     uint64
	pingType      uint16
}

// NewTCPPeer returns a TCPPeer
func NewTCPPeer(conn net.Conn, ID string, Name string, connectedTime int64) *TCPPeer {
	if len(Name) == 0 {
		Name = ID
	}
	p := &TCPPeer{
		conn:          conn,
		id:            ID,
		name:          Name,
		connectedTime: connectedTime,
		pingType:      types.DefineHashedType("p2p.PingMessage"),
	}

	go func() {
		defer p.Close()

		pingCountLimit := uint64(3)
		pingTicker := time.NewTicker(10 * time.Second)
		for !p.isClose {
			select {
			case <-pingTicker.C:
				if err := p.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
					return
				}
				_, err := p.conn.Write(util.Uint16ToBytes(p.pingType))
				if err != nil {
					return
				}
				if atomic.AddUint64(&p.pingCount, 1) > pingCountLimit {
					return
				}
			}
		}
	}()
	return p
}

// ID returns the id of the peer
func (p *TCPPeer) ID() string {
	return p.id
}

// Name returns the name of the peer
func (p *TCPPeer) Name() string {
	return p.name
}

// Close closes TCPPeer
func (p *TCPPeer) Close() {
	p.isClose = true
	p.conn.Close()
}

// ReadPacket returns a packet data
func (p *TCPPeer) ReadPacket() (uint16, bool, []byte, error) {
	r := NewCopyReader(p.conn)

	var t uint16
	for {
		if v, _, err := ReadUint16(r); err != nil {
			return 0, false, nil, err
		} else {
			atomic.StoreUint64(&p.pingCount, 0)
			if v == p.pingType {
				r.Reset()
				continue
			} else {
				t = v
				break
			}
		}
	}

	if cp, _, err := ReadUint8(r); err != nil {
		return 0, false, nil, err
	} else if Len, _, err := ReadUint32(r); err != nil {
		return 0, false, nil, err
	} else if Len == 0 {
		return 0, false, nil, ErrUnknownMessage
	} else {
		zbs := make([]byte, Len)
		if _, err := FillBytes(r, zbs); err != nil {
			return 0, false, nil, err
		}
		return t, cp == 1, r.Bytes(), nil
	}
}

// SendPacket sends packet to the WebsocketPeer
func (p *TCPPeer) SendPacket(bs []byte) {
	if err := p.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return
	}
	_, err := p.conn.Write(bs)
	if err != nil {
		return
	}
}

// ConnectedTime returns peer connected time
func (p *TCPPeer) ConnectedTime() int64 {
	return p.connectedTime
}
