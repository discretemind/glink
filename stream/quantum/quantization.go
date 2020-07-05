package quantum

import (
	"github.com/discretemind/glink/utils/crypto"
	"sync"
)

type Quantum struct {
	Index uint32
	Space  uint32 //Quantum space
}

type quantumPool struct {
	sync.RWMutex
	size     uint32
	free     []*Quantum
	assigned map[crypto.Certificate]*Quantum
}

func Pool(size uint32) (res *quantumPool) {
	res = &quantumPool{
		size: size,
	}
	return res
}

func (p *quantumPool) releaseQuantum(q *Quantum) {

}

func (p *quantumPool) issueQuantum(index uint32, size uint32) *Quantum {
	return &Quantum{
		Index: index,
		Space:  size,
	}
}

func (p *quantumPool) Resize(size uint32) {
	if p.size == size {
		return
	}
	p.Lock()
	for _, q := range p.free {
		p.releaseQuantum(q)
	}

	var i uint32
	for i < size {
		q := p.issueQuantum(i, size)
		p.free = append(p.free, q)
		i++
	}

	p.Unlock()
}
