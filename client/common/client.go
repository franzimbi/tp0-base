package common

import (
	"time"
	"net"
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
	protocol  *Protocol
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
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
	c.protocol = NewProtocol(conn)
	return nil
}
func (c *Client) SendBet(nombre string, apellido string, documento uint32, nacimiento string, numero uint32) {
	c.createClientSocket()

	err := c.protocol.sendBet(nombre, apellido, documento, nacimiento, numero)
	if err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | error: %v",
			err,
		)
		return
	}

	log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
		documento,
		numero,
	)
}

func (c *Client) Close() {
	c.protocol.Close()
}
