package rdp

import (
	"github.com/discretemind/glink/utils/crypto"
	"net"
)

type PeerKey [64]byte

func NewPeerKey(id crypto.Certificate, pub crypto.PublicKey) (res PeerKey) {
	copy(res[:32], id[:])
	copy(res[32:], pub[:])
	return
}

func (pk PeerKey) ID() (res crypto.Certificate) {
	copy(res[:], pk[:32])
	return
}

func (pk PeerKey) Public() (res crypto.PublicKey) {
	copy(res[:], pk[32:])
	return
}

type Peer struct {
	pk       PeerKey
	addr     *net.UDPAddr
	accepted bool
}

func newPeer(addr *net.UDPAddr, pk PeerKey) (res *Peer) {
	res = &Peer{
		addr: addr,
		pk:   pk,
	}
	return
}

func (p *Peer) IsAccepted() bool {
	return p.accepted
}

func (p *Peer) UpdateMetrics(m *metricsCmd) error {
	return nil
}
