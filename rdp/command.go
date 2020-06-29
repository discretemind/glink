package rdp

import (
	"fmt"
	"github.com/discretemind/glink/utils/crypto"
	"reflect"
)

/*
	Public Unencrypted ProtectedCommands
*/

//type SignedMessage struct {
//
//	Payload   []byte
//}

type SignedMessage struct {
	Data      []byte
	Signature []byte
}

//Command from clients
type connectCmd struct {
	Cluster crypto.Certificate //
	Peer    PeerKey            //32 byte
	Version Version
}

//Command from manager
type acceptCmd struct {
	Key          crypto.PublicKey
	ClusterIndex uint16
}

/*
	Protected Encrypted ProtectedCommands over channels
*/
var ProtectedCommands *commandRegistry

type ProtectedCommand struct {
	Command uint16
	Payload []byte
}

type startCmd struct {
	Config []byte
	Start  bool
}

type stopCmd struct {
	Stop bool
}

type metricsCmd struct {
	CpuUsage                   uint32
	MemTotal, MemUsed, MemFree uint32
}

type recordSyncCmd struct {
	Index uint64
}

type recordSyncResponseCmd struct {
	Index uint64
	OK    bool
}

type commandRegistry struct {
	byType map[reflect.Type]uint16
	byID   map[uint16]reflect.Type
}

func newRegistry() *commandRegistry {
	return &commandRegistry{
		byType: make(map[reflect.Type]uint16),
		byID:   make(map[uint16]reflect.Type),
	}
}

func (r commandRegistry) register(id uint16, value interface{}) {
	t := reflect.TypeOf(value)
	fmt.Println("register", id, value)
	r.byType[t] = id
	r.byID[id] = t
}

func (r commandRegistry) GetCommand(id uint16) (res reflect.Value, ok bool) {
	t, ok := r.byID[id]
	if !ok {
		return reflect.Value{}, false
	}
	return reflect.New(t), true
}

func (r commandRegistry) GetIdByType(t reflect.Type) (res uint16, ok bool) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	res, ok = r.byType[t]
	return
}

func (r commandRegistry) GetId(cmd interface{}) (res uint16, ok bool) {
	t := reflect.TypeOf(cmd)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	res, ok = r.byType[t]
	return
}

func init() {
	ProtectedCommands = newRegistry()

	ProtectedCommands.register(1, metricsCmd{})
	ProtectedCommands.register(2, startCmd{})
	ProtectedCommands.register(3, stopCmd{})
}
