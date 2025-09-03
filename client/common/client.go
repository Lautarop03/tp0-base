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
	config   ClientConfig
	conn     net.Conn
	protocol Protocol
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
	c.protocol = Protocol{conn: conn}
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

	c.runBatchesLoop(reader)
	c.protocol.sendBetsSubmissionCompleted()

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)

	c.runWinnersLoop()

	c.conn.Close()
}

// runBatchesLoop handles the main loop for sending batches of bets to the server.
// It creates a client socket, sends the client ID, and continuously reads and sends batches
// of bets until there are no more bets to send or an error occurs.
func (c *Client) runBatchesLoop(reader *csv.Reader) {
	c.createClientSocket()
	c.protocol.sendClientID(c.config.ID)

	for {
		batch := c.createBatch(reader)
		if len(batch) == 0 {
			break
		}

		if !c.sendBatch(batch) {
			return
		}

		time.Sleep(c.config.LoopPeriod)
	}
}

// sendBatch sends a batch of bets to the server and waits for a confirmation response.
// It returns true if the batch was sent and confirmed successfully, otherwise false.
func (c *Client) sendBatch(batch []ClientBet) bool {
	err := c.protocol.sendBatch(batch, c.config.ID)
	if err != nil {
		log.Errorf("action: send_batch | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return false
	}

	confirmation, err := c.protocol.receiveBatchConfirmation()
	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return false
	}

	result := "fail"
	if confirmation.Success {
		result = "success"
	}
	log.Infof("action: apuesta_enviada | result: %s | message: %v",
		result,
		confirmation.Message,
	)
	return true
}

func (c *Client) runWinnersLoop() {
	for {
		c.createClientSocket()
		c.protocol.sendClientID(c.config.ID)

		c.protocol.requestWinners()

		ganadores, err := c.protocol.readWinners()
		if err != nil {
			log.Errorf("action: leer_ganadores | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		if ganadores == nil { // 0x06 WAIT
			c.conn.Close()
			time.Sleep(3 * time.Second) // TODO; esta bien usar sleep?
			continue
		}

		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", len(ganadores))

		return
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

		batchSize += c.protocol.sizeBytes(&clientBet)

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
