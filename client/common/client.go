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

const TOTAL_FIELDS_BET = 5

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
func (c *Client) SendBets(filePath string, agencyID uint32, maxBatchAmount int) {
	c.createClientSocket()

	err := c.protocol.SendAgencyID(agencyID)
	if err != nil {
		log.Criticalf("action: send_agency_id | result: fail | agency_id: %d, error: %s", agencyID, err)
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		log.Criticalf("action: open file | result: fail | error: %s", err)
		return
	}
	defer f.Close()

	var bets []Bet

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		content := strings.Split(line, ",")
		if len(content) != TOTAL_FIELDS_BET {
			log.Warningf("action: parse_bet | result: invalid_format | line: %s", line)
			continue
		}
		documento, err := strconv.ParseUint(content[2], 10, 32)
		if err != nil {
			log.Warningf("action: parse_bet | result: invalid_document | line: %s", line)
			continue
		}
		numero, err := strconv.ParseUint(content[4], 10, 32)
		if err != nil {
			log.Warningf("action: parse_bet | result: invalid_number | line: %s", line)
			continue
		}
		bet := Bet{
			nombre:     content[0],
			apellido:   content[1],
			documento:  uint32(documento),
			nacimiento: content[3],
			numero:     uint32(numero),
		}
		bets = append(bets, bet)

		if len(bets) >= maxBatchAmount {
			sent, err := c.protocol.SendBetsOnChunks(bets)
			if err != nil {
				log.Infof("action: send_chunck | result: fail | err: %s", err)
				return
			}
			if sent < len(bets) {
				bets = bets[sent:]
			} else {
				bets = make([]Bet, 0)
			}
			ok, _ := c.protocol.ReceivedCodeOfConfirmation()
			if !ok {
				log.Infof("action: answer_of_chunck | result: fail")
			}
		}

	}
	if len(bets) > 0 {
		_, err := c.protocol.SendBetsOnChunks(bets)
		if err != nil {
			log.Errorf("action: send_chunck | result: fail | err: %s", err)
		}

		ok, _ := c.protocol.ReceivedCodeOfConfirmation()
		if !ok {
			log.Infof("action: answer_of_chunck | result: fail")
		}
	}
	err = c.protocol.SendCodeToFinishSendingChuncks()
	if err != nil {
		log.Errorf("action: send_end_code | result: error | err: %s", err)
	}
}

func (c *Client) Close() {
	c.protocol.Close()
}
