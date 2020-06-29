package rdp

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/discretemind/glink/utils/crypto"
	"github.com/discretemind/glink/utils/encoder"
	"go.uber.org/zap"
	"net"
	"reflect"
	"sync"
)

type cluster struct {
	sync.Mutex
	id             uint16
	key            crypto.PrivateKey
	peers          map[crypto.Certificate]*Peer
	peersByAddr    map[string]*Peer
	handlers       map[uint16]handlerType
	partitionsSize uint32
	name           string
	outbox         chan *outboxMessage
	logger         *zap.Logger
}

type ICluster interface {
	ID() crypto.Certificate
	Start(peer string, config interface{}) error
	Stop(peer string) error
	Disable(peer string) (err error)
}

func Cluster(outbox chan *outboxMessage, id uint16, name string, partitions uint32, logger *zap.Logger) (res *cluster) {
	res = &cluster{
		name:           name,
		id:             id,
		outbox:         outbox,
		peers:          make(map[crypto.Certificate]*Peer),
		peersByAddr:    make(map[string]*Peer),
		key:            crypto.GeneratePrivateKey(),
		partitionsSize: partitions,
		handlers:       make(map[uint16]handlerType),
		logger:         logger.With(zap.String("cluster", name)),
	}

	res.registerHandler(res.handleMetrics)
	return
}

func (c *cluster) ID() crypto.Certificate {
	return c.key.Certificate()
}

func (c *cluster) Partitions(num uint32) {
	c.Lock()
	c.partitionsSize = num
	c.Unlock()
}

func (c *cluster) Handle(addr *net.UDPAddr, packet []byte) error {
	if p, ok := c.peersByAddr[addr.String()]; ok {

		err := c.handleMessage(p, packet)
		if err != nil {
			return err
		}

		//if !p.accepted {
		//	return fmt.Errorf("peer [%s] wasn't accepted yet", addr.String())
		//}

		return nil
	}

	return fmt.Errorf("unknown Peer %s", addr.String())
}

//func (c *cluster) AcceptPeer(id crypto.Certificate) (err error) {
//	if p, ok := c.peers[id]; ok {
//		p.accepted = true
//		return c.messageAccepted(p)
//	}
//	return fmt.Errorf("not found perr with id %x", id[:])
//}

func (c *cluster) Disable(id string) (err error) {
	//if p, ok := c.peers[id]; ok {
	//	p.accepted = true
	//	return c.messageAccepted(p)
	//}
	return fmt.Errorf("not found perr with id %x", id[:])
}

func (c *cluster) packMessage(p *Peer, value interface{}) (res *outboxMessage, err error) {

	id, ok := ProtectedCommands.GetId(value)
	if !ok {
		return nil, fmt.Errorf("command not registered %v", value)
	}

	encoded, err := c.key.Encrypt(p.pk.Public(), encoder.EncodeRaw(value))
	if err != nil {
		return
	}

	cmd := ProtectedCommand{
		Command: id,
		Payload: encoded,
	}
	c.logger.Info("Send command ", zap.Uint16("command", id))
	packet := Packet{}
	copy(packet[:], encoder.EncodeRaw(cmd))

	res = &outboxMessage{
		Addr: p.addr,
		Packet: packet,
	}

	return
}

func (c *cluster) handleMessage(p *Peer, data []byte) (err error) {

	message := &ProtectedCommand{}
	if err := encoder.DecodeRaw(data, message); err != nil {
		return err
	}
	cmd, ok := ProtectedCommands.GetCommand(message.Command)

	if !ok {
		return fmt.Errorf("command not found %d", message.Command)
	}

	//data := encoder.EncodeRaw(value)
	decoded, ok := c.key.Decrypt(p.pk.Public(), message.Payload)
	if !ok {
		return fmt.Errorf("command decrypt message %d", message.Command)
	}
	if err := encoder.DecodeRaw(decoded, cmd.Interface()); err != nil {
		return err
	}

	h, ok := c.handlers[message.Command]
	if !ok {
		return fmt.Errorf("handler not found for command %d", message.Command)
	}
	v := reflect.Value(h)
	res := v.Call([]reflect.Value{reflect.ValueOf(p), cmd})

	if len(res) > 0 && !res[0].IsNil() {
		return res[0].Interface().(error)

	}
	return nil
}

func (c *cluster) Peers() (res []*Peer) {
	for _, c := range c.peers {
		res = append(res, c)
	}
	return
}

func (c *cluster) Start(peer string, config interface{}) error {
	peerId := crypto.CertificateFromString(peer)

	cfgData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	if p, ok := c.peers[peerId]; ok {
		if msg, err := c.packMessage(p, startCmd{
			Start:  true,
			Config: cfgData,
		}); err != nil {
			return err
		} else {
			c.logger.Info("Start message", zap.Uint16("command", binary.BigEndian.Uint16(msg.Packet[:2])))
			c.outbox <- msg
		}
	}
	return nil
}

func (c *cluster) Stop(peer string) error {
	peerId := crypto.CertificateFromString(peer)
	if p, ok := c.peers[peerId]; ok {
		if msg, err := c.packMessage(p, stopCmd{
			Stop: true,
		}); err != nil {
			return err
		} else {
			c.outbox <- msg
		}
	}
	return nil
}

func (c *cluster) addPeerID(p *Peer) error {
	c.logger.Info("added new Peer", zap.String("Peer", p.addr.String()))
	c.Lock()
	c.peersByAddr[p.addr.String()] = p
	c.peers[p.pk.ID()] = p
	c.Unlock()

	return c.messageConnected(p)
}

func (c *cluster) messageConnected(p *Peer) error {
	out := &outboxMessage{
		Addr: p.addr,
	}
	cmd := acceptCmd{
		Key:          c.key.Public(),
		ClusterIndex: c.id,
	}
	data := encoder.EncodeRaw(cmd)
	msg := SignedMessage{
		Data:      data,
		Signature: c.key.Sign(data),
	}
	c.logger.Info("Send accept message ", zap.Binary("sign", msg.Signature))
	copy(out.Packet[:], encoder.EncodeRaw(msg))

	c.outbox <- out
	return nil
}

//func (c *cluster) messageAccepted(p *Peer) error {
//	msg := &outboxMessage{
//		Addr: p.addr,
//	}
//	pub := c.key.Public()
//	copy(msg.Packet[:32], pub[:])
//	copy(msg.Packet[32:], c.key.Sign(pub[:]))
//	c.outbox <- msg
//	return nil
//}

//func (c *cluster) messageStart(p *Peer, config []byte) error {
//	msg := &outboxMessage{
//		Addr: p.addr,
//	}
//	pub := c.key.Public()
//	copy(msg.Packet[:32], pub[:])
//	copy(msg.Packet[32:], c.key.Sign(pub[:]))
//	c.outbox <- msg
//	return nil
//}
