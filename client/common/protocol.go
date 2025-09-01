package common

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
)

const stringLenOverhead = 1 // Overhead for string length prefix

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
	// First, send a message with the batch length, then start sending the bets
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

// Helper to read a string with a uint16 big-endian prefix
func readString(conn net.Conn) (string, error) {
	lenBytes := make([]byte, 2)
	if _, err := io.ReadFull(conn, lenBytes); err != nil {
		return "", err
	}
	length := binary.BigEndian.Uint16(lenBytes)

	strBytes := make([]byte, length)
	if _, err := io.ReadFull(conn, strBytes); err != nil {
		return "", err
	}
	return string(strBytes), nil
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
