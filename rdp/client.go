package rdp

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/discretemind/glink/utils/crypto"
	"github.com/discretemind/glink/utils/encoder"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
	"math"
	"net"
	"reflect"
	"time"
)

type client struct {
	key          crypto.PrivateKey
	version      Version
	logger       *zap.Logger
	outbox       chan *Packet
	clusterIndex uint16
	clusterId    crypto.Certificate
	clusterKey   crypto.PublicKey
	conn         *net.UDPConn
	handlers     map[uint16]handlerType
}

func Client(version Version, logger *zap.Logger) (res *client) {
	res = &client{
		key:      crypto.GeneratePrivateKey(),
		version:  version,
		outbox:   make(chan *Packet, 100),
		handlers: make(map[uint16]handlerType),
	}
	res.logger = logger.With(zap.String("id", res.key.Certificate().String()))
	res.registerHandler(res.startHandler)
	return
}

func (c *client) runOutbox(parent context.Context) {
	ctx, _ := context.WithCancel(parent)
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Close client")
			return
		case p := <-c.outbox:
			if p == nil {
				c.logger.Info("Close client channel")
				return
			}
			if _, err := c.conn.Write(p[:]); err != nil {
				c.logger.Error("can't write to master", zap.Error(err))
			}
		}
	}
}

func (c *client) ID() crypto.Certificate {
	return c.key.Certificate()
}

func (c *client) Connect(ctx context.Context, master string, id string) (err error) {
	clusterIdData, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return err
	}

	remote, err := net.ResolveUDPAddr("udp", master)
	if err != nil {
		return err
	}

	copy(c.clusterId[:], clusterIdData)

	//ctx, _ := context.WithCancel(context.Background())
	//go c.doConnect(remote, clusterId)
	//<-ctx.Done()

	c.logger.Info("dial to ", zap.String("to", remote.String()))
	c.conn, err = net.DialUDP("udp", nil, remote)
	if err != nil {
		return err
	}

	pub := c.key.Public()
	pk := NewPeerKey(c.key.Certificate(), pub)
	packet := Packet{01, 01}

	data := encoder.EncodeRaw(connectCmd{
		Cluster: c.clusterId,
		Version: c.version,
		Peer:    pk,
	})
	signedData := encoder.EncodeRaw(SignedMessage{
		Data:      data,
		Signature: c.key.Sign(data[:]),
	})

	copy(packet[2:], signedData[:])

	childCtx, cancel := context.WithCancel(ctx)
	go func() {
		err = c.runConnectionReader(childCtx, c.conn, c.clusterId, pub)
		cancel()
	}()

	go c.runOutbox(ctx)

	_, err = c.conn.Write(packet[:])
	if err != nil {
		cancel()
	}

	<-ctx.Done()

	return err
}

func (c *client) runConnectionReader(ctx context.Context, conn *net.UDPConn, clusterId crypto.Certificate, pubKey [32]byte) (err error) {
	err = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return
	}

	packet := Packet{}
	if _, err = conn.Read(packet[:]); err != nil {
		return err
	}

	msg := SignedMessage{}
	if err := encoder.DecodeRaw(packet[:], &msg); err != nil {
		return err
	}

	if !clusterId.Verify(msg.Data, msg.Signature) {
		return errors.New("invalid signature")
	}

	cmd := acceptCmd{}
	if err := encoder.DecodeRaw(msg.Data, &cmd); err != nil {
		return err
	}

	c.clusterKey = cmd.Key
	c.clusterIndex = cmd.ClusterIndex

	c.logger.Info("Connection Accepted", zap.Uint16("Cluster", cmd.ClusterIndex))
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return err
	}
	for {
		go c.runHealthCheck(ctx, 5*time.Second)
		packet := Packet{}
		if _, err := conn.Read(packet[:]); err != nil {
			return err
		}
		if err = c.handlePacket(packet[:]); err != nil {
			c.logger.Error("can't handle error", zap.Error(err))
		}

	}
}
func (c *client) handlePacket(data []byte) error {
	msg := ProtectedCommand{}
	if err := encoder.DecodeRaw(data, &msg); err != nil {
		return err
	}
	cmd, ok := ProtectedCommands.GetCommand(msg.Command)
	if !ok {
		return fmt.Errorf("command not found %d", msg.Command)
	}
	c.logger.Info("Client command ", zap.Uint16("id", msg.Command))
	h, ok := c.handlers[msg.Command]
	if !ok {
		return fmt.Errorf("can't execute. handler not found %d", msg.Command)
	}

	decoded, ok := c.key.Decrypt(c.clusterKey, msg.Payload[:])
	if !ok {
		return errors.New("can't decrypt message")
	}

	if err := encoder.DecodeRaw(decoded, cmd.Interface()); err != nil {
		return err
	}

	v := reflect.Value(h)
	res := v.Call([]reflect.Value{cmd})
	if len(res) > 0 && !res[0].IsNil() {
		return res[0].Interface().(error)

	}
	return nil
}

func (c *client) runHealthCheck(parent context.Context, period time.Duration) {
	ctx, _ := context.WithCancel(parent)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(period):
			if err := c.publishMetrics(); err != nil {
				c.logger.Error("can't publish metrics ", zap.Error(err))
			}
		}
	}
}

const mBytes = 1024 * 1024

func (c *client) publishMetrics() error {
	v, _ := mem.VirtualMemory()
	cp, _ := cpu.Percent(1*time.Second, false)
	var cpuUsage uint32 = 0
	if len(cp) > 0 {
		cpuUsage = uint32(math.Round(cp[0] * 100))
	}

	cmd := metricsCmd{
		CpuUsage: cpuUsage,
		MemTotal: uint32(v.Total / mBytes),
		MemUsed:  uint32(v.Used / mBytes),
		MemFree:  uint32(v.Free / mBytes),
	}
	packet, err := c.packMessage(cmd)
	if err != nil {
		return err
	}
	c.outbox <- packet
	return nil
}

func (c *client) packMessage(value interface{}) (res *Packet, err error) {
	packet := Packet{}
	binary.BigEndian.PutUint16(packet[:2], c.clusterIndex)

	commandId, ok := ProtectedCommands.GetId(value)
	if !ok {
		return nil, fmt.Errorf("command not registered %v", reflect.TypeOf(value).String())
	}

	encoded, err := c.key.Encrypt(c.clusterKey, encoder.EncodeRaw(value))
	if err != nil {
		return
	}

	cmd := &ProtectedCommand{
		Command: commandId,
		Payload: encoded,
	}

	copy(packet[2:], encoder.EncodeRaw(cmd))
	return &packet, nil
}
