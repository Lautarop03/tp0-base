package common

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
)

func serializar_msg(c *ClientBet, clientId string) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Helper for strings
	writeString := func(s string) error {
		if err := binary.Write(buf, binary.BigEndian, uint16(len(s))); err != nil {
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

func sendMsg(conn net.Conn, c *ClientBet, clientId string) error {
	data, err := serializar_msg(c, clientId)

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

func receiveConfirmationMsg(conn net.Conn) (*ClientBet, error) {
	firstName, err := readString(conn)
	if err != nil {
		return nil, err
	}

	lastName, err := readString(conn)
	if err != nil {
		return nil, err
	}

	document, err := readString(conn)
	if err != nil {
		return nil, err
	}

	birthdate, err := readString(conn)
	if err != nil {
		return nil, err
	}

	number, err := readString(conn)
	if err != nil {
		return nil, err
	}

	bet := &ClientBet{
		Nombre:     firstName,
		Apellido:   lastName,
		Documento:  document,
		Nacimiento: birthdate,
		Numero:     number,
	}

	return bet, nil
}
