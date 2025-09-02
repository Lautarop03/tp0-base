package common

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const (
	stringLenOverhead = 1 // Overhead for string length prefix

	OpcodeInitClient     = 0x01
	OpcodeBatchBets      = 0x02
	OpcodeFinishedBets   = 0x04
	OpcodeRequestWinners = 0x05
	OpcodeWait           = 0x06
	OpcodeWinners        = 0x07
)

type Protocol struct {
	conn net.Conn
}

func (p *Protocol) sendOpcode(opcode uint8) error {
	return binary.Write(p.conn, binary.BigEndian, opcode)
}

func (p *Protocol) sendString(s string) error {
	if err := binary.Write(p.conn, binary.BigEndian, uint8(len(s))); err != nil {
		return err
	}

	// To avoid short-write and ensure the entire message is sent
	total := 0
	data := []byte(s)
	for total < len(data) {
		n, err := p.conn.Write(data[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return nil
}

func (p *Protocol) sendBet(c *ClientBet, clientId string) error {
	fields := []string{
		clientId,
		c.Nombre,
		c.Apellido,
		c.Documento,
		c.Nacimiento,
		c.Numero,
	}

	for _, field := range fields {
		if err := p.sendString(field); err != nil {
			return err
		}
	}

	return nil
}

func (p *Protocol) sendBatch(batch []ClientBet, clientID string) error {
	// First opcode, then send a message with the batch length, then start sending the bets
	if err := p.sendOpcode(OpcodeBatchBets); err != nil {
		return err
	}

	if err := binary.Write(p.conn, binary.BigEndian, uint16(len(batch))); err != nil {
		return err
	}

	for _, bet := range batch {
		if err := p.sendBet(&bet, clientID); err != nil {
			return err
		}
	}
	return nil
}

// receiveBatchConfirmation receives the batch confirmation from the server.
// Format:
//
//	1 byte: success (0 = fail, 1 = success)
//	2 bytes: message length (uint16 big-endian)
//	N bytes: message (utf-8)
type BatchConfirmation struct {
	Success bool
	Message string
}

func (p *Protocol) receiveBatchConfirmation() (*BatchConfirmation, error) {
	header := make([]byte, 3)
	if _, err := io.ReadFull(p.conn, header); err != nil {
		return nil, err
	}

	success := header[0] == 1
	msgLen := binary.BigEndian.Uint16(header[1:3])

	msgBytes := make([]byte, msgLen)
	if _, err := io.ReadFull(p.conn, msgBytes); err != nil {
		return nil, err
	}

	return &BatchConfirmation{
		Success: success,
		Message: string(msgBytes),
	}, nil
}

func (p *Protocol) sizeBytes(clientBet *ClientBet) int {
	fields := []string{
		clientBet.Nombre,
		clientBet.Apellido,
		clientBet.Documento,
		clientBet.Nacimiento,
		clientBet.Numero,
	}

	size := 0
	for _, f := range fields {
		size += stringLenOverhead + len([]byte(f))
	}
	return size
}

func (p *Protocol) sendClientID(clientID string) error {
	if err := p.sendOpcode(OpcodeInitClient); err != nil {
		return err
	}

	return p.sendString(clientID)
}

func (p *Protocol) sendBetsSubmissionCompleted() error {
	// Send a message de un byte con 0x04 indicating that all bets have been sent
	return p.sendOpcode(OpcodeFinishedBets)
}

func (p *Protocol) requestWinners() error {
	// Send a message indicating that the client wants to consult the winners
	return p.sendOpcode(OpcodeRequestWinners)
}

func (p *Protocol) receiveWinnersList() ([]string, error) {
	// Read number of winners (2 bytes)
	header := make([]byte, 2)
	if _, err := io.ReadFull(p.conn, header); err != nil {
		return nil, err
	}
	numWinners := binary.BigEndian.Uint16(header)

	winners := make([]string, 0, numWinners)
	for i := 0; i < int(numWinners); i++ {
		// Read length of document (1 byte)
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(p.conn, lenBuf); err != nil {
			return nil, err
		}
		docLen := int(lenBuf[0])

		// Read document (docLen bytes)
		docBuf := make([]byte, docLen)
		if _, err := io.ReadFull(p.conn, docBuf); err != nil {
			return nil, err
		}
		winners = append(winners, string(docBuf))
	}
	return winners, nil
}

func (p *Protocol) readWinners() ([]string, error) {
	opcode := make([]byte, 1)
	if _, err := io.ReadFull(p.conn, opcode); err != nil {
		return nil, err
	}

	switch opcode[0] {
	case OpcodeWait:
		// El servidor indica que todavia no hay ganadores
		return nil, nil
	case OpcodeWinners:
		// El servidor envía la lista de ganadores
		return p.receiveWinnersList()
	default:
		return nil, fmt.Errorf("unexpected opcode: %v", opcode[0])
	}
}
