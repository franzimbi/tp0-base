package common

import (
	"encoding/binary"
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

func (p *Protocol) FullRead(bytes int) ([]byte, error) {
	// la forma de no tener un short read
	var buf = make([]byte, bytes)
	total := 0

	for total < bytes {
		n, err := p.skt.Read(buf[total:])
		if err != nil {
			return buf[:total], err
		}
		total += n
	}
	return buf, nil
}

func stringToBytes(s string) []byte {
	var buf = make([]byte, 0)
	buf = append(buf, byte(len(s)))
	buf = append(buf, []byte(s)...)
	return buf
}

func ui32ToLittleEndianBytes(num uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, num)
	return buf
}

func (p *Protocol) Close() {
	if p.skt != nil {
		p.skt.Close()
	}
}

func (p *Protocol) BetToBytes(nombre string, apellido string, documento uint32, nacimiento string, numero uint32) []byte {
	var buf = make([]byte, 0)
	buf = append(buf, stringToBytes(nombre)...)
	buf = append(buf, stringToBytes(apellido)...)
	buf = append(buf, ui32ToLittleEndianBytes(documento)...)
	buf = append(buf, stringToBytes(nacimiento)...)
	buf = append(buf, ui32ToLittleEndianBytes(numero)...)
	return buf
}

func (p *Protocol) SendBytes(data []byte, size uint32) error {
	sizeBytes := ui32ToLittleEndianBytes(size)
	return p.FullWrite(append(sizeBytes, data...))
}

func (p *Protocol) SendAgentID(id uint32) error {
	idBytes := ui32ToLittleEndianBytes(id)
	return p.FullWrite(idBytes)
}

func (p *Protocol) RecvAck() (bool, error) {
	ack, err := p.FullRead(1)
	if err == nil && ack[0] == 1 {
		return true, err
	}
	return false, err
}
