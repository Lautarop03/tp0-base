package common

import (
	"encoding/csv"
	"io"
	"net"
	"os"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID                string
	ServerAddress     string
	LoopAmount        int
	LoopPeriod        time.Duration
	BatchMaxAmount    int
	BatchMaxBytesSize int
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

		batch := c.createBatch()

		err := sendBatch(c.conn, batch, c.config.ID)

		if err != nil {
			log.Errorf("action: send_batch | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		confirmationBet, err := receiveConfirmationMsg(c.conn) // cambia el msg de exito para todo el batch ahora
		c.conn.Close()

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
			confirmationBet.Documento,
			confirmationBet.Numero,
		)

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) createBatch() []ClientBet {
	// Open the CSV file
	file, err := os.Open("/agency.csv") // El file lo necesito abrir desde afuera asi mantengo el orden desde donde estoy con el reader
	if err != nil {
		log.Errorf("action: open_file | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return nil
	}
	defer file.Close() // Close the file when the function returns

	reader := csv.NewReader(file)

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

		if len(batch) >= c.config.BatchMaxAmount || batchSize >= c.config.BatchMaxBytesSize {
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
