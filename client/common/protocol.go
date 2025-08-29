package common

import (
	"bytes"
	"encoding/binary"
	"net"
)

// TODO: Manejo de errores y coms en ingles
func serializar_msg(c *ClientBet, clientId string) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Helper para strings
	writeString := func(s string) error {
		if err := binary.Write(buf, binary.BigEndian, uint16(len(s))); err != nil {
			return err
		}
		if _, err := buf.Write([]byte(s)); err != nil {
			return err
		}
		return nil
	}

	// Serializar campos
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

// TODO: Manejo de errores y coms en ingles
func sendMsg(conn net.Conn, c *ClientBet, clientId string) error {
	data, err := serializar_msg(c, clientId)

	if err != nil {
		return err
	}

	// Para evitar el short-write y asegurar el envio de todo el mensaje
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
