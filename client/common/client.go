package common

import (
	"bufio"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/op/go-logging"
)

const TOTAL_FIELDS_BET = 5
const BASE_TEN = 10
const BITS_32 = 32
const NAME_POSITION = 0
const SURNAME_POSITION = 1
const DOCUMENT_POSITION = 2
const BIRTHDATE_POSITION = 3
const NUMBER_POSITION = 4

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

func parseBet(line string) (Bet, error) {
	content := strings.Split(line, ",")
	if len(content) != TOTAL_FIELDS_BET {
		log.Warningf("action: parse_bet | result: invalid_format | line: %s", line)
		return Bet{}, errors.New("invalid format")
	}
	documento, err := strconv.ParseUint(content[DOCUMENT_POSITION], BASE_TEN, BITS_32)
	if err != nil {
		log.Warningf("action: parse_bet | result: invalid_document | line: %s", line)
		return Bet{}, errors.New("invalid format")
	}
	numero, err := strconv.ParseUint(content[NUMBER_POSITION], BASE_TEN, BITS_32)
	if err != nil {
		log.Warningf("action: parse_bet | result: invalid_number | line: %s", line)
		return Bet{}, errors.New("invalid format")
	}
	bet := Bet{
		nombre:     content[NAME_POSITION],
		apellido:   content[SURNAME_POSITION],
		documento:  uint32(documento),
		nacimiento: content[BIRTHDATE_POSITION],
		numero:     uint32(numero),
	}
	return bet, nil
}

func (c *Client) sendChucksAndReceiveConfirmation(bets []Bet) (int, error) {
	sent, err := c.protocol.SendBetsOnChunks(bets)
	if err != nil {
		log.Errorf("action: send_chunck | result: fail | err: %s", err)
		return 0, err
	}
	ok, _ := c.protocol.ReceivedCodeOfConfirmation()
	if !ok {
		log.Infof("action: answer_of_chunck | result: fail")
	}
	return sent, nil
}

func (c *Client) SendBets(filePath string, agencyID uint32, maxBatchAmount int) {
	c.createClientSocket()
	// defer c.Close()

	err := c.protocol.SendAgencyID(agencyID)
	if err != nil {
		log.Warningf("action: send_agency_id | result: fail | agency_id: %d, error: %s", agencyID, err)
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		log.Warningf("action: open file | result: fail | error: %s", err)
		return
	}
	defer f.Close()

	var bets []Bet

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		bet, err := parseBet(line)
		if err != nil {
			continue
		}
		bets = append(bets, bet)

		if len(bets) >= maxBatchAmount {
			sent, err := c.sendChucksAndReceiveConfirmation(bets)
			if err != nil {
				log.Error("error al enviar chunk")
				return
			}
			if sent < len(bets) {
				bets = bets[sent:]
			} else {
				bets = make([]Bet, 0)
			}
		}
	}
	if len(bets) > 0 {
		_, err := c.sendChucksAndReceiveConfirmation(bets)
		if err != nil {
			log.Error("error al enviar los ultimos bets")
			return
		}
	}
	err = c.protocol.SendCodeToFinishSendingChuncks()
	if err != nil {
		log.Errorf("action: send_end_code | result: error | err: %s", err)
	}
	log.Info("termino el client sendBets")
	// time.Sleep(c.config.LoopPeriod)
	// c.Close()
}

func (c *Client) Close() {
	c.protocol.Close()
}
