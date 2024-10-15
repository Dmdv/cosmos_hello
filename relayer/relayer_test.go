package relayer

import (
	"testing"
	"time"
)

func TestIBCRelay(t *testing.T) {
	relay := NewIBCRelay()
	go relay.ProcessPackets()
	go relay.ProcessAcknowledgments()

	// Test sending packets
	relay.SendPacket(Packet{SequenceNumber: 1, Data: []byte("Packet 1")})
	relay.SendPacket(Packet{SequenceNumber: 2, Data: []byte("Packet 2")})
	relay.SendPacket(Packet{SequenceNumber: 3, Data: []byte("Packet 3")})

	// Simulate receiving packets with different block times
	time.Sleep(time.Second)
	relay.ReceivePacket(1, []byte("Packet 1"))
	time.Sleep(2 * time.Second)
	relay.ReceivePacket(2, []byte("Packet 2"))
	time.Sleep(500 * time.Millisecond)
	relay.ReceivePacket(3, []byte("Packet 3"))

	// Simulate packet acknowledgments
	relay.ackChan <- PacketAcknowledgment{SequenceNumber: 1, Success: true}
	relay.ackChan <- PacketAcknowledgment{SequenceNumber: 2, Success: true}
	relay.ackChan <- PacketAcknowledgment{SequenceNumber: 3, Success: false}

	// Wait for packets to be processed
	time.Sleep(2 * time.Second)

	// Check if packets are correctly acknowledged
	if !relay.acknowledgments[1] {
		t.Errorf("Packet 1 should be acknowledged")
	}
	if !relay.acknowledgments[2] {
		t.Errorf("Packet 2 should be acknowledged")
	}
	if relay.acknowledgments[3] {
		t.Errorf("Packet 3 should not be acknowledged")
	}

	// Check if packets are deleted after acknowledgment
	if _, ok := relay.packets[1]; ok {
		t.Errorf("Packet 1 should be deleted after acknowledgment")
	}
	if _, ok := relay.packets[2]; ok {
		t.Errorf("Packet 2 should be deleted after acknowledgment")
	}
	if _, ok := relay.packets[3]; !ok {
		t.Errorf("Packet 3 should not be deleted due to failed acknowledgment")
	}

	// Simulate packet timeout
	time.Sleep(time.Minute)
	if _, ok := relay.packets[3]; ok {
		t.Errorf("Packet 3 should be deleted due to timeout")
	}
}
