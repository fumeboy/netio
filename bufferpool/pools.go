package SmartBuffer

import (
	"sync"
)

var pools = pools_typ{}

type pools_typ struct {
	pools []*pool
	lock  sync.Mutex
}

func init() {
	l := int(highestLV)+1
	pools.pools = make([]*pool, l)
	for i := 0; i < l; i++ {
		pools.pools[i] = new_pool(level(i))
	}
}
