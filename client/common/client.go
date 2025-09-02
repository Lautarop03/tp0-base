package common

import (
	"encoding/csv"
	"io"
	"net"
	"os"
	"time"

	"github.com/op/go-logging"
)

const batchMaxBytesSize = 8 * 1024 // 8 kB

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID             string
	ServerAddress  string
	LoopAmount     int
	LoopPeriod     time.Duration
	BatchMaxAmount int
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

// StartClientLoop reads all betting records from the CSV file and sends them in batches to the server.
// For each batch, it waits for a confirmation before proceeding to the next batch.
// The loop finishes when all records have been sent or an error occurs.
func (c *Client) StartClientLoop() {
	// Open the CSV file
	file, err := os.Open("/agency.csv")
	if err != nil {
		log.Errorf("action: open_file | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}
	defer file.Close() // Close the file when the function returns

	reader := csv.NewReader(file)

	c.createClientSocket()
	sendClientID(c.conn, c.config.ID)
	defer c.conn.Close()

	for {
		batch := c.createBatch(reader)
		if len(batch) == 0 {
			break
		}

		// TODO: el protocolo puede ser un objeto y que tenga la info que le mandamos siempre en cada func
		err = sendBatch(c.conn, batch, c.config.ID)

		if err != nil {
			log.Errorf("action: send_batch | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		confirmation, err := receiveBatchConfirmation(c.conn)

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		result := "fail"
		if confirmation.Success {
			result = "success"
		}
		log.Infof("action: apuesta_enviada | result: %s | message: %v",
			result,
			confirmation.Message,
		)

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)
	}

	sendBetsSubmissionCompleted(c.conn)

	c.conn.Close()

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)

	for {
		c.createClientSocket()
		sendClientID(c.conn, c.config.ID)

		// TODO: Consultar la lista de ganadores del sorteo de mi agencia : c.config.ID
		requestWinners(c.conn) // envio el msg

		// si me llega un 0x06-WAIT voy a tener que cerrar la conexion y conectarme en un rato a preguntar denuevo

		ganadores, err := readWinners(c.conn)
		if err != nil { // opcode incorrecto
			log.Errorf("action: leer_ganadores | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		// TODO: no me gusta el chequeo este de ganadores=nil
		if ganadores == nil { // 0x06 WAIT
			// esperar y consultar despues, pero tengo que reconectarme
			c.conn.Close()
			time.Sleep(3 * time.Second) // TODO; esta bien usar sleep?
			continue
		} else { // 0x07 GANADORES
			log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", len(ganadores))
			return
		}
	}
}

func (c *Client) createBatch(reader *csv.Reader) []ClientBet {
	batch := make([]ClientBet, 0, c.config.BatchMaxAmount)

	batchSize := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("action: read_record_csv | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
		}

		clientBet := ClientBet{
			Nombre:     record[0],
			Apellido:   record[1],
			Documento:  record[2],
			Nacimiento: record[3],
			Numero:     record[4],
		}

		batchSize += sizeBytes(&clientBet)

		batch = append(batch, clientBet)

		if len(batch) >= c.config.BatchMaxAmount || batchSize >= batchMaxBytesSize {
			break
		}
	}
	return batch
}

func (c *Client) Stop() {
	if c.conn != nil {
		c.conn.Close()
		log.Infof("action: stop | result: success | detail: client socket was closed")
	}
}
