package common

import (
	"encoding/binary"
	"fmt"
	"net"
)

type Protocol struct {
	skt net.Conn
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
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, num)
	err := p.FullWrite(buf)
	return err
}

func (p *Protocol) SendString(msg string) error {
	buf := make([]byte, 1)
	buf[0] = byte(len(msg))
	err := p.FullWrite(buf)
	if err != nil {
		return err
	}
	return p.FullWrite([]byte(msg))
}

func (p *Protocol) Close() {
	if p.skt != nil {
		p.skt.Close()
	}
}

func (p *Protocol) sendBet(nombre string, apellido string, documento uint32, nacimiento string, numero uint32) error {

	err := p.SendString(nombre)
	if err != nil {
		return err
	}
	err = p.SendString(apellido)
	if err != nil {
		return err
	}
	err = p.SendInt(documento)
	if err != nil {
		return err
	}
	err = p.SendString(nacimiento)
	if err != nil {
		return err
	}
	err = p.SendInt(numero)
	if err != nil {
		return err
	}
	return nil
}

func (p *Protocol) RecvAck() error {
	buf := make([]byte, 1)
	_, err := p.skt.Read(buf)
	if err != nil {
		return err
	}
	if buf[0] != byte(1) {
		return fmt.Errorf("ack no recibido")
	}
	return nil
}
