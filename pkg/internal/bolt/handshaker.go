package bolt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fbiville/go-usain-go/pkg/internal/packstream"
	"net"
)

type Handshaker struct {
	connection net.Conn
}

func (h *Handshaker) shakeHands(supportedVersion []byte) error {
	_, err := h.connection.Write([]byte{0x60, 0x60, 0xB0, 0x17})
	if err != nil {
		return fmt.Errorf("could not send handshake preamble %w", err)
	}
	_, err = h.connection.Write(append(
		supportedVersion,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
	))
	if err != nil {
		return fmt.Errorf("could not send handshake supported versions %w", err)
	}
	response := []byte{0x0, 0x0, 0x0, 0x0}
	err = binary.Read(h.connection, packstream.Endianness, response)
	if err != nil {
		return fmt.Errorf("could not receive handshake response %w", err)
	}
	responseLength := len(response)
	if responseLength != 4 {
		return fmt.Errorf("expected handshake response to be 4 byte long, got %d", responseLength)
	}
	if !bytes.Equal(response, supportedVersion) {
		return fmt.Errorf("expected \"%d.%d\", but got \"%d.%d\" version",
			supportedVersion[2], supportedVersion[3],
			response[2], response[3])
	}
	return nil
}
