package rdp

type partition struct {
	index uint64
	owner *Peer
}

func Partition()(res *partition){
	res = &partition{}
	return
}

func (p *partition) UpdateIndex(index uint64){
	p.index = index
}

func (p *partition) UpdateOwner(peer *Peer){
	//p.index = index
}
