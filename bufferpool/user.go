package SmartBuffer

import (
	"sync"
)

type User struct {
	maxLevel level
	nowLevel level

	stat         []int // 存储每个 level 的使用量
	stat_lock    sync.Mutex
	ranking      []level // 除了 nowLevel，还有哪些 level 经常被使用（ buffer 升级时的选项
	ranking_lock sync.Mutex

	resize_lock chan int8

	busy int // 一个检查周期里，进行主动 resize 的次数
}

func NewUser(max level) *User {
	u := &User{
		maxLevel:    max,
		nowLevel:    LV_1024, // 默认的👨初始分配大小
		stat:        make([]int, max+1),
		resize_lock: make(chan int8, 1),
		ranking:     make([]level, 0),
	}
	pools.pools[u.nowLevel].grow(1)
	highest_check(max)
	return u
}

// 约定：由该 User 『给出(get)』 的 Buffer 必须由该 User 『回收(put)』
func (this *User) Get() (b *Buffer) {
	var p *pool
	for i, l := int(this.nowLevel), int(this.maxLevel); i < l; i++ {
		p = pools.pools[i]
		if b := p.get(); b != nil {
			b.user = this
			return b
		}
	}
	// 进行到这里说明没能找到空余的buffer
	// 并且可以这样认为，当前瞬间的并发请求数 >= len(p.blocks)
	select {
	case this.resize_lock <- 1:
		this.busy++
		old_blocks_num := pools.pools[this.nowLevel].blocks_end // 旧时所需的量
		this.resize()                                           // 忙时调整
		p = pools.pools[this.nowLevel]
		b = p.grow(old_blocks_num + old_blocks_num/2)
	default:
		p = pools.pools[this.nowLevel]
		b = p.grow(p.blocks_end + p.blocks_end/2)
	}
	b.user = this
	return b
}

func (this *User) stat_hook(b *Buffer) {
	blv := b.level
	this.stat_lock.Lock()
	if blv > 0 { // 如果有降级的余地
		lower := blv.lookdown(b.UsedSize())
		// 剩余的空间多于所分配的一半
		this.stat[lower]++
		goto label

	}
	this.stat[blv]++
label:
	this.stat_lock.Unlock()
}

//

func quickSort(arr []int, l, r int) {
	if l < r {
		pivot := arr[r]
		i := l - 1
		for j := l; j < r; j++ {
			if arr[j] >= pivot {
				i++
				arr[j], arr[i] = arr[i], arr[j]
			}
		}
		i++
		arr[r], arr[i] = arr[i], arr[r]
		quickSort(arr, l, i-1)
		quickSort(arr, i+1, r)
	}
}

func (this *User) resize() {
	var length = int(this.maxLevel) + 1
	this.stat_lock.Lock()
	stat := this.stat
	this.stat = make([]int, length)
	this.stat_lock.Unlock()
	temp_map := map[int]int{}
	stat_sum := 0
	for i := 0; i < length; i++ { // 在 map 中存放 stat 的 lv（stat的index） 和对应的 频率
		temp_map[stat[i]] = i // 如果 lv_1 的频率和 lv_2 的频率一致，高级别的排在前面
		stat_sum += stat[i]
	}
	if stat_sum < defaultPoolAcceleration { // 临时写法，给出 resize 的下限，如果访问量过低，不会进行 resize
		<-this.resize_lock
		return
	}
	quickSort(stat, 0, length-1)
	ranking := []level{}
	highest := temp_map[stat[0]]
	this.nowLevel = level(highest)
	for i := 1; i < length; i++ {
		if lv := temp_map[stat[i]]; lv > highest {
			ranking = append(ranking, level(lv))
		}
	}
	this.ranking_lock.Lock()
	this.ranking = ranking
	this.ranking_lock.Unlock()
	<-this.resize_lock
}

func (this *User) levelup(b *Buffer) bool {
	if b.level >= this.maxLevel {
		return false
	}
	var ranking []level
	var swap_buf *Buffer
	var (
		i        int
		l        int
		p        *pool
		first    *pool
		first_be int
	)
	this.ranking_lock.Lock()
	ranking = this.ranking
	this.ranking_lock.Unlock()

	l = len(ranking)
	if l > 0 {
		first = pools.pools[ranking[0]]
		for ; i < l; i++ {
			if ranking[i] > b.level {
				p = pools.pools[i]
				if swap_buf = p.get(); swap_buf != nil {
					goto swap
				}
			}
		}
	} else {
		first = pools.pools[b.level+1]
	}
	first_be = first.blocks_end
	for i, l = int(b.level)+1, int(this.maxLevel); i < l; i++ {
		p = pools.pools[i]
		if swap_buf = p.get(); swap_buf != nil {
			goto swap
		}
	}
	swap_buf = first.grow(first_be + 1)
swap:
	oldbuf := b.buf
	newbuf := swap_buf.buf
	oldlv := b.level
	newlv := swap_buf.level
	oldblock := b.block
	newblock := swap_buf.block

	b.buf = newbuf
	b.level = newlv
	b.Cap = newlv.size()
	b.block = newblock
	swap_buf.buf = oldbuf
	swap_buf.Cap = oldlv.size()
	swap_buf.level = oldlv
	swap_buf.block = oldblock

	copy(newbuf, oldbuf)
	swap_buf.GoHome()
	return true
}
