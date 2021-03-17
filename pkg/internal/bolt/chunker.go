package bolt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fbiville/go-usain-go/pkg/internal/packstream"
	"math"
	"net"
)

type Chunker struct {
	Connection net.Conn
}

// FIXME: assumes message is writable to a single chunk
func (c *Chunker) WriteChunked(rawMessage []byte) error {
	message := chunk(rawMessage)
	_, err := c.Connection.Write(message)
	return err
}

// FIXME: assumes message is readable from a single chunk
func (c *Chunker) ReadUnchunked() ([]byte, error) {
	chunk := make([]byte, 2)
	err := binary.Read(c.Connection, packstream.Endianness, chunk)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(chunk, []byte{0, 0}) {
		return nil, fmt.Errorf("could not read response chunk")
	}
	responseSize := packstream.Endianness.Uint16(chunk)
	response := make([]byte, responseSize)
	err = binary.Read(c.Connection, packstream.Endianness, response)
	return response, err
}

func chunk(rawMessage []byte) []byte {
	length := len(rawMessage)
	if length > math.MaxUint16 {
		return nil
	}
	buffer := make([]byte, 2+length+2)
	packstream.Endianness.PutUint16(buffer[:2], uint16(length))
	copy(buffer[2:len(buffer)-2], rawMessage)
	buffer[len(buffer)-2] = 0
	buffer[len(buffer)-1] = 0
	return buffer
}
