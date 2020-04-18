package SmartBuffer

import (
	"sync"
)

type level int

var highestLV level = LV_1024x16
var highestLVLock = sync.Mutex{}

const (
	LV_128 level = iota
	LV_256
	LV_512
	LV_1024
	LV_1024x2
	LV_1024x4
	LV_1024x8
	LV_1024x16
)

func (this level) size() int {
	return 128 << uint(this)
}

func (this level) lookdown(size int) level {
	lv_size := this.size()
	lv := this
	for {
		if lv == 0 {
			return 0
		}
		lv_size >>= 1
		if lv_size <= size {
			return lv
		}
		lv--
	}
}

func highest_check(lv level) {
	highestLVLock.Lock()
	if lv > highestLV {
		for i := highestLV + 1; i <= lv; i++ {
			pools.pools = append(pools.pools, new_pool(i))
		}
		highestLV = lv
	}
	highestLVLock.Unlock()
}
