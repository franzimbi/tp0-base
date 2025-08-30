package common

import (
	"encoding/binary"
	"net"
)

const CODE_OF_CONFIRMATION byte = 0
const CODE_OF_BEGIN_SENDING byte = 0
const CODE_OF_END byte = 1
const ONE_BYTE = 1
const INT_BYTES = 4

type Protocol struct {
	skt net.Conn
}

const MaxDataSizekB = 8

const MaxDataSize = (MaxDataSizekB * 1024) - 4 // le resto 4 bytes para el size del chunk

type Bet struct {
	nombre     string
	apellido   string
	documento  uint32
	nacimiento string
	numero     uint32
}

func NewProtocol(skt net.Conn) *Protocol {
	return &Protocol{
		skt: skt,
	}
}

func (p *Protocol) FullWrite(data []byte) error {
	// la forma de no tener un short write
	total := 0
	for total < len(data) {
		n, err := p.skt.Write(data[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return nil
}

func (p *Protocol) SendInt(num uint32) error {
	buf := make([]byte, INT_BYTES)
	binary.LittleEndian.PutUint32(buf, num)
	err := p.FullWrite(buf)
	return err
}

func (p *Protocol) Close() {
	if p.skt != nil {
		p.skt.Close()
	}
}

func (p *Protocol) ReceivedCodeOfConfirmation() (bool, error) {
	buf := make([]byte, ONE_BYTE)
	n, err := p.skt.Read(buf)
	if err == nil && buf[0] == CODE_OF_CONFIRMATION && n == ONE_BYTE {
		return true, err
	}
	return false, err
}

func (c *Protocol) SendCodeToStartSendingChuncks() error {
	buf := make([]byte, ONE_BYTE)
	buf[0] = CODE_OF_BEGIN_SENDING
	return c.FullWrite(buf)
}

func (c *Protocol) SendCodeToFinishSendingChuncks() error {
	buf := make([]byte, ONE_BYTE)
	buf[0] = CODE_OF_END
	return c.FullWrite(buf)
	// if err != nil {
	// 	return err
	// }
	// _, err = c.ReceivedCodeOfConfirmation() // espero ultimo ack
	// return err
}

func ui32ToLittleEndianBytes(num uint32) []byte {
	buf := make([]byte, INT_BYTES)
	binary.LittleEndian.PutUint32(buf, num)
	return buf
}

func (p *Protocol) SendAgencyID(agencyID uint32) error {
	err := p.SendInt(agencyID)
	return err
}

func stringToBytes(s string) []byte {
	var buf = make([]byte, 0)
	buf = append(buf, byte(len(s)))
	buf = append(buf, []byte(s)...)
	return buf
}

func (p *Protocol) betToBytes(bet Bet) []byte {
	var buf = make([]byte, 0)
	buf = append(buf, stringToBytes(bet.nombre)...)
	buf = append(buf, stringToBytes(bet.apellido)...)
	buf = append(buf, ui32ToLittleEndianBytes(bet.documento)...)
	buf = append(buf, stringToBytes(bet.nacimiento)...)
	buf = append(buf, ui32ToLittleEndianBytes(bet.numero)...)
	return buf
}

func (p *Protocol) SendBetsOnChunks(bets []Bet) (int, error) {
	p.SendCodeToStartSendingChuncks()
	chunk := make([]byte, 0)
	bets_counter := 0
	for _, v := range bets {
		data := p.betToBytes(v)
		if len(chunk)+len(data) > MaxDataSize {
			break
		}
		chunk = append(chunk, data...)
		bets_counter += 1
	}
	sizeFinalChunck := ui32ToLittleEndianBytes(uint32(bets_counter))
	err := p.FullWrite(append(sizeFinalChunck, chunk...))
	return bets_counter, err
}
