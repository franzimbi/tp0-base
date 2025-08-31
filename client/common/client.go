package common

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

const (
	NAME_POSITION      = 0
	SURNAME_POSITION   = 1
	DOCUMENT_POSITION  = 2
	BIRTHDATE_POSITION = 3
	NUMBER_POSITION    = 4
	MAXCHUNKSIZE       = (8 * 1024) - 4
)

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config   ClientConfig
	protocol *Protocol
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:   config,
		protocol: nil,
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
	c.protocol = NewProtocol(conn) // esto lo dejo pq venia asi, pero si es por mi q se cree el Protocol adentro del cliente
	return nil
}

func (c *Client) SendBets(filePath string, id uint32, max int) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Warningf("action: open file | result: fail | error: %s", err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	bytesChunk := make([]byte, 0)
	betsCounter := 0
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) != 5 {
			log.Warningf("action: parse line | result: fail | line: %s", line)
			continue
		}
		doc, _ := strconv.ParseUint(parts[DOCUMENT_POSITION], 10, 32)
		num, _ := strconv.ParseUint(parts[NUMBER_POSITION], 10, 32)
		bytes := c.protocol.BetToBytes(parts[NAME_POSITION], parts[SURNAME_POSITION], uint32(doc), parts[BIRTHDATE_POSITION], uint32(num))

		bytesChunk = append(bytesChunk, bytes...)
		betsCounter++

		if len(bytesChunk)+len(bytes) >= MAXCHUNKSIZE || betsCounter >= int(max) {
			if c.createClientSocket() != nil {
				log.Error("action: create_socket | result: fail")
				return
			}
			err := c.protocol.SendAgentID(id)
			if err != nil {
				log.Errorf("action: send_agent_id | result: fail | error: %v", err)
				return
			}
			err = c.protocol.SendBytes(bytesChunk, uint32(betsCounter))
			if err != nil {
				log.Errorf("action: send_bets | result: fail | error: %v", err)
				return
			}
			if b, e := c.protocol.RecvAck(); e != nil || !b {
				log.Errorf("action: send_bets | result: fail | error: no se recibió ack del servidor")
				return
			}
			log.Infof("action: send_bets | result: success | bets_sent: %v", betsCounter)
			bytesChunk = make([]byte, 0)
			betsCounter = 0
			c.Close()
		}
	}
	if betsCounter > 0 {
		if c.createClientSocket() != nil {
			log.Error("action: create_socket | result: fail")
			return
		}
		err := c.protocol.SendAgentID(id)
		if err != nil {
			log.Errorf("action: send_agent_id | result: fail | error: %v", err)
			return
		}
		err = c.protocol.SendBytes(bytesChunk, uint32(betsCounter))
		if err != nil {
			log.Errorf("action: send_bets | result: fail | error: %v", err)
			return
		}
		if b, e := c.protocol.RecvAck(); e != nil || !b {
			log.Errorf("action: send_bets | result: fail | error: no se recibió ack del servidor")
			return
		}
		log.Infof("action: send_bets | result: success | bets_sent: %v", betsCounter)
		c.Close()
	}
}

func (c *Client) Close() {
	c.protocol.Close()
}
