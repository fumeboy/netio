package SmartBuffer

import (
	"sync"
)

const defaultPoolAcceleration = 64

type pool struct {
	lock sync.Mutex

	blocks         []*block
	blocks_end     int // blocks 实际可用长度，最尾 block 的 index
	block_cap      int
	blocks_cap_sum int

	lv level
}

func new_pool(lv level) *pool {
	p := &pool{}
	p.lv = lv
	p.block_cap = defaultPoolAcceleration
	return p
}

// grow() 获取新的内存
func (p *pool) grow(to int) (b *Buffer) {
	p.lock.Lock()
	n := to - p.blocks_end
	if n <= 0 { // grow to 是为了避免多goroutine时的过度扩容
		b = p.get()
		if b == nil { // 如果在等候 grow 的瞬间新增的资源又被抢完
			n = 1
			goto label
		}
		p.lock.Unlock()
		return
	}
label:
	var (
		i      int
		l               = p.block_cap * n
		size            = p.lv.size()
		bs     []Buffer = make([]Buffer, l)                // 每次新加分块数
		buf    []byte   = make([]byte, p.block_cap*size*n) // 新取一块连续的内存
		head   *Buffer                                     // 指定为 链表头
		node   *Buffer                                     // 链表结点
		block_ *block
		offest int
	)
	for j := 0; j < n; j++ {
		i = 1 + j*p.block_cap
		l = (j + 1) * p.block_cap
		head = &bs[j*p.block_cap]
		node = head // 链表结点
		block_ = &block{port: head, cap: p.block_cap, rest: p.block_cap}
		for ; i < l; i++ {
			node.buf = buf[offest:offest]
			node.level = p.lv
			node.next = &bs[i]
			node.Cap = size
			node.block = block_
			node = node.next
			offest += size
		}
		// 最后一块
		node.buf = buf[offest:offest]
		offest += size
		node.level = p.lv
		node.Cap = size
		node.block = block_
		node.next = nil
		p.blocks = append(p.blocks, block_)
	}
	b = block_.port
	block_.port = b.next
	p.blocks_end += n
	p.blocks_cap_sum += p.block_cap * n
	p.lock.Unlock()
	return
}

func (p *pool) release(num int) {
	p.lock.Lock()
	for i := 0; i < num; i++ {
		if p.blocks_end > 0 {
			p.blocks_end--
			b := p.blocks[p.blocks_end]
			b.lock.Lock()
			b.dead = true
			b.port = nil
			b.lock.Unlock()
		}
	}
	p.lock.Unlock()
}

//

func (p *pool) get() (b *Buffer) {
	for i := 0; i < p.blocks_end; i++ {
		block := p.blocks[i]
		if block.port != nil {
			block.lock.Lock()
			if block.port != nil {
				b = block.port
				block.port = b.next
				b.block.rest--
				block.lock.Unlock()
				return
			}
			block.lock.Unlock()
		}
	}
	return nil
}