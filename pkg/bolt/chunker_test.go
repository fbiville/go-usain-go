package bolt_test

import (
	"bytes"
	"fmt"
	"github.com/fbiville/go-usain-go/pkg/bolt"
	. "github.com/onsi/gomega"
	"math"
	"testing"
	"testing/quick"
)

func TestMessageChunking(t *testing.T) {
	RegisterTestingT(t)

	chunkInvariants := func(message []byte) bool {
		chunk, err := bolt.Chunk(message)
		messageLength := len(message)
		if messageLength > math.MaxUint16 {
			return err != nil &&
				err.Error() == fmt.Sprintf("message size (%d) does not fit 2 bytes", messageLength)
		}
		if len(chunk) != messageLength+4 {
			return false
		}
		if !bytes.HasPrefix(chunk, []byte{0x0, byte(messageLength)}) {
			return false
		}
		if !bytes.HasSuffix(chunk, []byte{0x0, 0x0}) {
			return false
		}
		return bytes.Contains(chunk, message)
	}
	if err := quick.Check(chunkInvariants, nil); err != nil {
		t.Error(err)
	}
}
