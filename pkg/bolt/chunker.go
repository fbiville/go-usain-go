package bolt

import (
	"fmt"
	"github.com/fbiville/go-usain-go/pkg/packstream"
	"math"
)

func Chunk(message []byte) ([]byte, error) {
	length := len(message)
	if length > math.MaxUint16 {
		return nil, fmt.Errorf("message size (%d) does not fit 2 bytes", length)
	}
	buffer := make([]byte, 2+length+2)
	packstream.Endianness.PutUint16(buffer[:2], uint16(length))
	copy(buffer[2:len(buffer)-2], message)
	buffer[len(buffer)-2] = 0
	buffer[len(buffer)-1] = 0
	return buffer, nil
}
