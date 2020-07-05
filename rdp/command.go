package rdp

import (
	"fmt"
	"github.com/discretemind/glink/stream/quantum"
	"github.com/discretemind/glink/utils/crypto"
	"reflect"
)

type Packet [300]byte

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
type ConnectCmd struct {
	Cluster crypto.Certificate //
	Peer    PeerKey            //32 byte
	Version Version
}

//Command from manager
type AcceptCmd struct {
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

type StartCmd struct {
	Config []byte
	Start  bool
}

type StopCmd struct {
	Stop bool
}

//from jobs
type MetricsCmd struct {
	CpuUsage                   uint32
	MemTotal, MemUsed, MemFree uint32
}

/*		 Quantum Sync			 */

//master =>
type syncStatusCmd struct {
	ID    string
	Space uint32 //Current Quantum space
}

//job =>
type syncStatusResponseCmd struct {
	ID      string
	Quantum map[uint32]uint32 //running quantum by space. Sometime it's not possible to close quantum space immediately because of Time Windows. When window will be closed - quantum will be released
	Status  uint8             //Current job status
}

//job => master
type releasingQuantumCmd struct {
	ID    string
	Space uint32 //Quantum space to release
}

//master => job. Accepted
type releasingQuantumResponseCmd struct {
	ID string
	OK bool
}

//master =>
type assignQuantumCmd struct {
	ID      string
	Quantum []quantum.Quantum
}

//job =>
type assignQuantumResponseCmd struct {
	ID string
	OK bool
}

//18 bytes
type DataSetStat struct {
	ID      uint16
	Records uint64
	Bytes   uint64
}

type recordSyncCmd struct {
	DataSet [16]DataSetStat //Max 16 data set stats at a time 288bytes
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

	ProtectedCommands.register(1, MetricsCmd{})
	ProtectedCommands.register(2, StartCmd{})
	ProtectedCommands.register(3, StopCmd{})
}
