package relayer

import (
	"sort"
	"sync"
	"time"
)

type Packet struct {
	SequenceNumber uint64
	Data           []byte
	Timeout        time.Time
}

type PacketAcknowledgment struct {
	SequenceNumber uint64
	Success        bool
}

type IBCRelay struct {
	sourceChan      chan Packet
	destChan        chan Packet
	ackChan         chan PacketAcknowledgment
	packets         map[uint64]Packet
	acknowledgments map[uint64]bool
	lock            sync.Mutex
}

func NewIBCRelay() *IBCRelay {
	return &IBCRelay{
		sourceChan:      make(chan Packet),
		destChan:        make(chan Packet),
		ackChan:         make(chan PacketAcknowledgment),
		packets:         make(map[uint64]Packet),
		acknowledgments: make(map[uint64]bool),
	}
}

func (r *IBCRelay) SendPacket(packet Packet) {
	r.sourceChan <- packet
}

func (r *IBCRelay) ReceivePacket(sequenceNumber uint64, data []byte) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.packets[sequenceNumber]; !ok {
		r.packets[sequenceNumber] = Packet{
			SequenceNumber: sequenceNumber,
			Data:           data,
			Timeout:        time.Now().Add(time.Minute),
		}
	}
}

func (r *IBCRelay) ProcessPackets() {
	for {
		select {
		case packet := <-r.sourceChan:
			r.lock.Lock()
			r.packets[packet.SequenceNumber] = packet
			r.lock.Unlock()
		case ack := <-r.ackChan:
			r.lock.Lock()
			r.acknowledgments[ack.SequenceNumber] = ack.Success
			if ack.Success {
				delete(r.packets, ack.SequenceNumber)
			}
			r.lock.Unlock()
		default:
			r.lock.Lock()
			packets := make([]Packet, 0, len(r.packets))
			for _, packet := range r.packets {
				packets = append(packets, packet)
			}
			sort.Slice(packets, func(i, j int) bool {
				return packets[i].SequenceNumber < packets[j].SequenceNumber
			})
			for _, packet := range packets {
				if time.Now().After(packet.Timeout) {
					delete(r.packets, packet.SequenceNumber)
					continue
				}
				select {
				case r.destChan <- packet:
				default:
				}
			}
			r.lock.Unlock()
			time.Sleep(time.Second)
		}
	}
}

func (r *IBCRelay) ProcessAcknowledgments() {
	for ack := range r.ackChan {
		r.lock.Lock()
		r.acknowledgments[ack.SequenceNumber] = ack.Success
		if ack.Success {
			delete(r.packets, ack.SequenceNumber)
		}
		r.lock.Unlock()
	}
}
