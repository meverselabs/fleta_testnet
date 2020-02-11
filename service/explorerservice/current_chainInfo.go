package explorerservice

import (
	"io"

	"github.com/fletaio/fleta_testnet/common/binutil"
)

type currentChainInfo struct {
	Foumulators         int    `json:"foumulators"`
	Blocks              uint32 `json:"blocks"`
	Transactions        int    `json:"transactions"`
	currentTransactions int
}

// WriteTo is a serialization function
func (c *currentChainInfo) WriteTo(w io.Writer) (int64, error) {
	var wrote int64

	if n, err := w.Write(binutil.LittleEndian.Uint32ToBytes(uint32(c.Foumulators))); err != nil {
		return wrote, err
	} else {
		wrote += int64(n)
	}
	if n, err := w.Write(binutil.LittleEndian.Uint32ToBytes(uint32(c.Blocks))); err != nil {
		return wrote, err
	} else {
		wrote += int64(n)
	}
	if n, err := w.Write(binutil.LittleEndian.Uint32ToBytes(uint32(c.Transactions))); err != nil {
		return wrote, err
	} else {
		wrote += int64(n)
	}
	if n, err := w.Write(binutil.LittleEndian.Uint32ToBytes(uint32(c.currentTransactions))); err != nil {
		return wrote, err
	} else {
		wrote += int64(n)
	}

	return wrote, nil
}

// ReadFrom is a deserialization function
func (c *currentChainInfo) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	bs := make([]byte, 4)
	if n, err := r.Read(bs); err != nil {
		return read, err
	} else {
		read += int64(n)
		c.Foumulators = int(binutil.LittleEndian.Uint32(bs))
	}
	if n, err := r.Read(bs); err != nil {
		return read, err
	} else {
		read += int64(n)
		c.Blocks = binutil.LittleEndian.Uint32(bs)
	}
	if n, err := r.Read(bs); err != nil {
		return read, err
	} else {
		read += int64(n)
		c.Transactions = int(binutil.LittleEndian.Uint32(bs))
	}
	if n, err := r.Read(bs); err != nil {
		return read, err
	} else {
		read += int64(n)
		c.currentTransactions = int(binutil.LittleEndian.Uint32(bs))
	}

	return read, nil
}
