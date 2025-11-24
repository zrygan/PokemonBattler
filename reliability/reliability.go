// Package reliability implements a reliability layer for UDP communication.
// It provides sequence number tracking, acknowledgements, and retransmission.
package reliability

import (
	"net"
	"sync"
	"time"

	"github.com/zrygan/pokemonbattler/messages"
)

const (
	DefaultTimeout    = 500 * time.Millisecond // 500ms timeout
	DefaultMaxRetries = 3                      // Maximum number of retries
)

// ReliableConnection wraps a UDP connection with reliability features.
type ReliableConnection struct {
	conn           *net.UDPConn
	sequenceNumber int
	expectedSeqNum int
	pendingAcks    map[int]*PendingMessage
	mu             sync.Mutex
	timeout        time.Duration
	maxRetries     int
}

// PendingMessage represents a message waiting for acknowledgement.
type PendingMessage struct {
	Message     messages.Message
	Destination *net.UDPAddr
	RetriesLeft int
	LastSent    time.Time
	AckReceived bool
}

// NewReliableConnection creates a new reliable connection wrapper.
func NewReliableConnection(conn *net.UDPConn) *ReliableConnection {
	return &ReliableConnection{
		conn:           conn,
		sequenceNumber: 1,
		expectedSeqNum: 1,
		pendingAcks:    make(map[int]*PendingMessage),
		timeout:        DefaultTimeout,
		maxRetries:     DefaultMaxRetries,
	}
}

// GetNextSequenceNumber returns the next sequence number and increments the counter.
func (rc *ReliableConnection) GetNextSequenceNumber() int {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	seqNum := rc.sequenceNumber
	rc.sequenceNumber++
	return seqNum
}

// SendReliable sends a message reliably with automatic retransmission.
// Returns the sequence number used for this message.
func (rc *ReliableConnection) SendReliable(msg messages.Message, dest *net.UDPAddr) (int, error) {
	// Get sequence number from message params if present, otherwise generate new one
	seqNum := rc.GetNextSequenceNumber()

	// Store for potential retransmission
	rc.mu.Lock()
	rc.pendingAcks[seqNum] = &PendingMessage{
		Message:     msg,
		Destination: dest,
		RetriesLeft: rc.maxRetries,
		LastSent:    time.Now(),
		AckReceived: false,
	}
	rc.mu.Unlock()

	// Send the message
	data := msg.SerializeMessage()
	_, err := rc.conn.WriteToUDP(data, dest)

	return seqNum, err
}

// ReceiveAck processes an ACK message.
func (rc *ReliableConnection) ReceiveAck(ackNumber int) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if pending, ok := rc.pendingAcks[ackNumber]; ok {
		pending.AckReceived = true
		delete(rc.pendingAcks, ackNumber)
	}
}

// SendAck sends an acknowledgement for a received message.
func (rc *ReliableConnection) SendAck(seqNum int, dest *net.UDPAddr) error {
	ackMsg := messages.MakeAck(seqNum)
	data := ackMsg.SerializeMessage()
	_, err := rc.conn.WriteToUDP(data, dest)
	return err
}

// CheckRetransmissions checks for unacknowledged messages and retransmits if needed.
// Should be called periodically (e.g., in a goroutine).
func (rc *ReliableConnection) CheckRetransmissions() []int {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	var failedSeqNums []int
	now := time.Now()

	for seqNum, pending := range rc.pendingAcks {
		if pending.AckReceived {
			continue
		}

		// Check if timeout has elapsed
		if now.Sub(pending.LastSent) > rc.timeout {
			if pending.RetriesLeft > 0 {
				// Retransmit
				data := pending.Message.SerializeMessage()
				rc.conn.WriteToUDP(data, pending.Destination)
				pending.LastSent = now
				pending.RetriesLeft--
			} else {
				// Max retries exceeded
				failedSeqNums = append(failedSeqNums, seqNum)
				delete(rc.pendingAcks, seqNum)
			}
		}
	}

	return failedSeqNums
}

// HasPendingMessages returns true if there are unacknowledged messages.
func (rc *ReliableConnection) HasPendingMessages() bool {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return len(rc.pendingAcks) > 0
}

// ClearPending removes a message from the pending list without waiting for ACK.
func (rc *ReliableConnection) ClearPending(seqNum int) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	delete(rc.pendingAcks, seqNum)
}
