package common

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"
	"os"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

type ClientBet struct {
	Nombre     string
	Apellido   string
	Documento  string
	Nacimiento string
	Numero     string
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()

		// TENGO QUE CARGAR LOS DATOS DE ENV
		clientBet := ClientBet{
			Nombre:     os.Getenv("NOMBRE"),
			Apellido:   os.Getenv("APELLIDO"),
			Documento:  os.Getenv("DOCUMENTO"),
			Nacimiento: os.Getenv("NACIMIENTO"),
			Numero:     os.Getenv("NUMERO"),
		}

		// SERIALIZAR EL MSG CON EL PROTOCOLO
		msg, err := serializar_msg(&clientBet)

		if err != nil {
			log.Errorf("action: serialize_msg | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		// ENVIAR EL MSG (ESTO Y LO DE ARRIBA SE ENCARGA EL PROTOCOLO, YO COMO CLIENTE NO SE NADA)
		err = sendMsg(c.conn, msg)
		if err != nil {
			log.Errorf("action: send_msg | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}
		log.Infof("action: send_msg | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)

		// TODO: RECIBIR LA CONFIRMACION (TODO A TRAVES DEL PROTOCOLO CON LA SER/DES SERIALIZACION ETC)
		recvMsg, err := bufio.NewReader(c.conn).ReadString('\n') // TODO:Recibir con el protocolo
		c.conn.Close()

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			recvMsg,
		)

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func serializar_msg(c *ClientBet) ([]byte, error) {
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

// Para evitar el short-write y asegurar el envio de todo el mensaje
func sendMsg(conn net.Conn, data []byte) error {
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

func (c *Client) Stop() {
	if c.conn != nil {
		c.conn.Close()
		log.Infof("action: stop | result: success | detail: client socket was closed")
	}
}
