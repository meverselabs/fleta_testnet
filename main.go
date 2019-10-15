package main // import "github.com/fletaio/fleta"

import (
	"encoding/binary"
	"log"
	"time"
)

func main() {
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, 35948995)

	start := time.Now()
	for i := 0; i < 1000000; i++ {
		binary.BigEndian.PutUint64(bs, 1)
	}
	log.Println(time.Now().Sub(start))
}
