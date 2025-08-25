package common

import (
	// "bufio"
	// "fmt"
	"encoding/binary"
	"net"
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

func fullWrite(client Client, data []byte) (int) {
	total := 0
	for total < len(data) {
		n, err := client.conn.Write(data[total:])
		if err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				client.config.ID,
				err,
			)
			return total
		}
		total += n
	}
	return total
}

func (c *Client) sendString(msg string) (int) {
	buf := make([]byte, 1)
	buf[0] = byte(len(msg))
	n := fullWrite(*c, buf)
	if n != 1 {
		return 0
	}
	n = fullWrite(*c, []byte(msg))
	return n
}

func (c *Client) sendInt(num uint32) (int) {
	buf := make([]byte, 4)
    binary.LittleEndian.PutUint32(buf, num)
	n := fullWrite(*c, buf)
	if n == 4 {
		return 0
	}else{
		return 1
	}
}
// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(nombre string, apellido string, documento uint32, nacimiento string, numero uint32) {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	
	c.createClientSocket()

	if c.sendString(nombre) != len(nombre) {
		return
	}
	if c.sendString(apellido) != len(apellido) {
		return
	}
	if c.sendInt(documento) == 0 {
		return
	}
	if c.sendString(nacimiento) != len(nacimiento) {
		return
	}
	if c.sendInt(numero) == 0 {
		return
	}

	log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
		documento,
		numero,
	)

	
	// for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
	// 	// Create the connection the server in every loop iteration. Send an
	// 	c.createClientSocket()

	// 	// TODO: Modify the send to avoid short-write
		// fmt.Fprintf(
		// 	c.conn,
		// 	"[CLIENT %v] Message N°%v\n",
		// 	c.config.ID,
		// 	msgID,
		// )
		// msg, err := bufio.NewReader(c.conn).ReadString('\n')
		// c.conn.Close()

	// 	if err != nil {
	// 		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
	// 			c.config.ID,
	// 			err,
	// 		)
	// 		return
	// 	}

		// log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
		// 	c.config.ID,
		// 	msg,
		// )

	// 	// Wait a time between sending one message and the next one
	// 	time.Sleep(c.config.LoopPeriod)

	// }
	// log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
