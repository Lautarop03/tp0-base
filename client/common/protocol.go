package common

import (
	"bytes"
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

func serializeBet(c *ClientBet, clientId string) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Helper for strings
	writeString := func(s string) error {
		if err := binary.Write(buf, binary.BigEndian, uint8(len(s))); err != nil {
			return err
		}
		if _, err := buf.Write([]byte(s)); err != nil {
			return err
		}
		return nil
	}

	if err := writeString(clientId); err != nil {
		return nil, err
	}
	if err := writeString(c.Nombre); err != nil {
		return nil, err
	}
	if err := writeString(c.Apellido); err != nil {
		return nil, err
	}
	if err := writeString(c.Documento); err != nil {
		return nil, err
	}
	if err := writeString(c.Nacimiento); err != nil {
		return nil, err
	}
	if err := writeString(c.Numero); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func sendBet(conn net.Conn, c *ClientBet, clientId string) error {
	data, err := serializeBet(c, clientId)

	if err != nil {
		return err
	}

	// To avoid short-write and ensure the entire message is sent
	total := 0
	for total < len(data) {
		n, err := conn.Write(data[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return nil
}

func sendBatch(conn net.Conn, batch []ClientBet, clientID string) error {
	// First opcode, then send a message with the batch length, then start sending the bets
	if err := binary.Write(conn, binary.BigEndian, uint8(OpcodeBatchBets)); err != nil {
		return err
	}

	if err := binary.Write(conn, binary.BigEndian, uint16(len(batch))); err != nil {
		return err
	}

	for _, bet := range batch {
		if err := sendBet(conn, &bet, clientID); err != nil {
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

func receiveBatchConfirmation(conn net.Conn) (*BatchConfirmation, error) {
	header := make([]byte, 3)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}

	success := header[0] == 1
	msgLen := binary.BigEndian.Uint16(header[1:3])

	msgBytes := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgBytes); err != nil {
		return nil, err
	}

	return &BatchConfirmation{
		Success: success,
		Message: string(msgBytes),
	}, nil
}

func sizeBytes(clientBet *ClientBet) int {
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

func sendClientID(conn net.Conn, clientID string) error {
	// Opcode
	if err := binary.Write(conn, binary.BigEndian, uint8(OpcodeInitClient)); err != nil {
		return err
	}
	// Largo
	if err := binary.Write(conn, binary.BigEndian, uint8(len(clientID))); err != nil {
		return err
	}
	// Payload
	if _, err := conn.Write([]byte(clientID)); err != nil {
		return err
	}
	return nil
}

func sendBetsSubmissionCompleted(conn net.Conn) error {
	// Send a message de un byte con 0x04 indicating that all bets have been sent
	if err := binary.Write(conn, binary.BigEndian, uint8(OpcodeFinishedBets)); err != nil {
		return err
	}
	return nil
}

func requestWinners(conn net.Conn) error {
	// Send a message indicating that the client wants to consult the winners
	if err := binary.Write(conn, binary.BigEndian, uint8(OpcodeRequestWinners)); err != nil {
		return err
	}
	return nil
}

func receiveWinnersList(conn net.Conn) ([]string, error) {
	// Read number of winners (2 bytes)
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}
	numWinners := binary.BigEndian.Uint16(header)

	winners := make([]string, 0, numWinners)
	for i := 0; i < int(numWinners); i++ {
		// Read length of document (1 byte)
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return nil, err
		}
		docLen := int(lenBuf[0])

		// Read document (docLen bytes)
		docBuf := make([]byte, docLen)
		if _, err := io.ReadFull(conn, docBuf); err != nil {
			return nil, err
		}
		winners = append(winners, string(docBuf))
	}
	return winners, nil
}

func readWinners(conn net.Conn) ([]string, error) {
	opcode := make([]byte, 1)
	if _, err := io.ReadFull(conn, opcode); err != nil {
		return nil, err
	}

	switch opcode[0] {
	case OpcodeWait:
		// El servidor indica que todavia no hay ganadores
		return nil, nil
	case OpcodeWinners:
		// El servidor envía la lista de ganadores
		return receiveWinnersList(conn)
	default:
		return nil, fmt.Errorf("unexpected opcode: %v", opcode[0])
	}
}
