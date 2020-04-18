package SmartBuffer

import (
	"time"
)

var watcher = &watcher_typ{}

type watcher_typ struct {
	ticker *time.Ticker
	users  []*User
}

func (this *watcher_typ) run() {
	defer this.ticker.Stop()
	var (
		i  int
		l  int
		n  int
		p  *pool
		ll = len(pools.pools)
	)
	var pool_flag = make([]int, ll)
	var rest_sum int
	for {
		select {
		case <-this.ticker.C:
			// watcher 定时检查 User 是否需要 resize
			// 以及 pool 是否需要释放多余资源
			// 如果该周期内发生过 User 主动的 resize，则跳过它的这次 resize 检查
			// 同时跳过 User.nowLevel 对应的池的检查
			for i, l = 0, len(this.users); i < l; i++ {
				user := this.users[i]
				if user.busy > 0 {
					pool_flag[user.nowLevel] = 1 // 跳过 User.nowLevel 对应的池的检查
					user.busy = 0
					continue
				}
				user.resize_lock <- 1
				user.resize()
			}
			for i = 0; i < ll; i++ {
				rest_sum = 0
				if pool_flag[i] == 1 {
					continue
				}
				p = pools.pools[i]
				for i := 0; i < p.blocks_end; i++ {
					rest_sum += p.blocks[i].rest
				}
				n = (p.blocks_cap_sum - rest_sum) / p.block_cap
				if n > 1 {
					p.release(n)
				}
			}
		}
	}
}

// func init() {
// 	watcher.ticker = time.NewTicker(time.Minute * 1)
// 	watcher.users = []*User{}
// 	go watcher.run()
// }
