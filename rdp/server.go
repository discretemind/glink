package rdp

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/discretemind/glink/utils/crypto"
	"github.com/discretemind/glink/utils/encoder"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"
)

type Packet [300]byte

type connMessage struct {
	Addr    *net.UDPAddr
	Payload []byte
}

type outboxMessage struct {
	Addr   *net.UDPAddr
	Packet Packet
}

type server struct {
	sync.Mutex
	name     string
	logger   *zap.Logger
	conn     *net.UDPConn
	clusters map[uint16]*cluster
	connPipe chan *connMessage
	outbox   chan *outboxMessage
	closing  bool
}

func Server(name string, logger *zap.Logger) (res *server) {
	res = &server{
		name: name,
		//key:      key,
		connPipe: make(chan *connMessage, 100),
		outbox:   make(chan *outboxMessage, 100),
		clusters: make(map[uint16]*cluster),
		logger:   logger,
	}
	return
}

func (s *server) CreateCluster(id uint16, name string, size uint32) (res *cluster) {
	s.Lock()
	res = Cluster(s.outbox, id, name, size, s.logger)
	s.clusters[id] = res
	s.Unlock()
	return
}

func (s *server) Listen(ctx context.Context, port int) error {
	s.logger.Info("Start listener", zap.Int("Port", port))
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	go s.runConnectionPipe(ctx)
	go s.runOutbox(ctx)

	go func() {
		for !s.closing {
			s.Lock()
			s.conn, err = net.ListenUDP("udp", addr)
			s.Unlock()
			if err != nil {
				s.logger.Error("Cant run listener", zap.String("address", addr.String()), zap.Error(err))
				time.Sleep(time.Second)
			}
			if err := s.listenerReader(s.conn); err != nil {
				s.logger.Error("Reader error", zap.String("address", addr.String()), zap.Error(err))
			}
		}
	}()
	<-ctx.Done()
	s.closing = true
	//if err := conn.Close(); err != nil {
	//	s.logger.Error("Can't close connection", zap.Error(err))
	//}
	close(s.connPipe)
	return nil
}

var (
	connectPrefix = string([]byte{1, 1})
)

func (s *server) listenerReader(conn *net.UDPConn) (err error) {

	for !s.closing {
		packet := [260]byte{}
		if _, addr, err := conn.ReadFromUDP(packet[:]); err != nil {
			return err
		} else {
			//s.logger.Info("incoming packet ", zap.String("from", addr.String()), zap.Binary("prefix", packet[:2]))
			switch string(packet[:2]) {
			case connectPrefix:
				s.logger.Info("Connection message")
				s.connPipe <- &connMessage{
					Addr:    addr,
					Payload: packet[2:],
				}
			default:
				clusterIndex := binary.BigEndian.Uint16(packet[:2])
				if cluster, ok := s.clusters[clusterIndex]; ok {
					if err := cluster.Handle(addr, packet[2:]); err != nil {
						s.logger.Info("Can't handle cluster message", zap.Error(err))
					}

				} else {
					s.logger.Warn("Cluster not found", zap.Uint16("cluster", clusterIndex))
				}
			}
		}
	}
	s.logger.Info("Exit reader")
	return nil
}

func (s *server) runOutbox(parent context.Context) {
	ctx, _ := context.WithCancel(parent)
	for {
		select {
		case <-ctx.Done():
			return
		case out := <-s.outbox:
			if !s.closing {
				if _, err := s.conn.WriteToUDP(out.Packet[:], out.Addr); err != nil {
					s.logger.Error("Can't send message to peer", zap.String("peer", out.Addr.String()))
				}
			}
		}
	}
}

func (s *server) getCluster(cert crypto.Certificate) (res *cluster, err error) {
	s.Lock()
	defer s.Unlock()
	for _, cl := range s.clusters {
		if cl.key.Certificate() == cert {
			return cl, nil
		}
	}
	return nil, fmt.Errorf("Cluster not found %s\n", cert.String())

}
func (s *server) runConnectionPipe(parent context.Context) {
	ctx, _ := context.WithCancel(parent)
	for {
		select {
		case <-ctx.Done():
			return
		case conn := <-s.connPipe:
			s.logger.Info("Got new connection", zap.String("Peer", conn.Addr.String()))

			//s.logger.Info("cluster message", zap.Binary("data", conn.Payload[:]))
			msg := SignedMessage{}
			if err := encoder.DecodeRaw(conn.Payload[:], &msg); err != nil {
				s.logger.Error("Invalid message format", zap.Error(err))
				continue
			}

			cmd := connectCmd{}
			if err := encoder.DecodeRaw(msg.Data, &cmd); err != nil {
				s.logger.Error("Invalid message format", zap.Error(err))
				continue
			}

			cluster, err := s.getCluster(cmd.Cluster)
			if err != nil {
				s.logger.Error("Cluster not found", zap.Error(err))
				continue
			}

			if cmd.Peer.ID().Verify(msg.Data, msg.Signature) {
				if err := cluster.addPeerID(newPeer(conn.Addr, cmd.Peer)); err != nil {
					s.logger.Error("cant add peer", zap.Error(err))
				}
			} else {
				s.logger.Error("Can't verify signature")
			}

		}
	}
}
