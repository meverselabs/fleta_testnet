package loop

import (
	"sync"
	"time"
)

type LoopManager struct {
	sync.Mutex
	isClosed bool
	isRun    bool
	items    []*loopItem
}

func NewLoopManager() *LoopManager {
	lm := &LoopManager{
		items: []*loopItem{},
	}
	return lm
}

func (lm *LoopManager) Close() {
	lm.isClosed = true
}

func (lm *LoopManager) Add(fn func(), interval time.Duration) {
	lm.Lock()
	if lm.isRun {
		lm.Unlock()
		panic("Loop cannot added after run executed")
	}
	lm.Unlock()

	lm.items = append(lm.items, &loopItem{
		fn:       fn,
		interval: interval,
		timer:    interval,
	})
}

func (lm *LoopManager) Run() {
	lm.Lock()
	lm.isRun = true
	lm.Unlock()

	prev := time.Now()
	for !lm.isClosed {
		now := time.Now()
		timedelta := now.Sub(prev)
		hasRun := false
		for _, v := range lm.items {
			v.timer -= timedelta
			if v.timer <= 0 {
				v.fn()
				v.timer += v.interval
				hasRun = true
			}
		}
		if !hasRun {
			time.Sleep(10 * time.Millisecond)
		}
		prev = now
	}
}

type loopItem struct {
	fn       func()
	interval time.Duration
	timer    time.Duration
}
