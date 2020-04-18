package SmartBuffer

import (
	"sync"
)

type block struct {
	port *Buffer
	lock sync.Mutex
	cap  int
	dead bool
	rest int
}
