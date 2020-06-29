package rdp

import "crypto"

type connectPacket struct {
	ClusterId crypto.PublicKey
	JobId     crypto.PublicKey
}

type RecordInfo struct {
	Count uint64
	Bytes uint64
}

type jobTelemetry struct {
	ClusterId crypto.PublicKey
	NodeID    uint64
	Count     uint64
	Bytes     uint64
}

type message struct {
	ClusterId crypto.PublicKey
	Topic     string `json:"topic"`
	Payload   []byte `json:"payload"`
}
