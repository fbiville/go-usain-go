package slices_test

import (
	bytes "bytes"
	"github.com/fbiville/go-usain-go/pkg/internal/slices"
	"reflect"
	"testing"
	"testing/quick"
)

func TestPadLeft(t *testing.T) {
	padLeftInvariants := func(data []byte, symbol byte, targetSize byte) bool {
		result := slices.PadLeft(data, symbol, targetSize)
		if len(result) != max(len(data), int(targetSize)) {
			return false
		}
		paddingSize := max(0, int(targetSize)-len(data))
		if !bytes.HasPrefix(result, bytes.Repeat([]byte{symbol}, paddingSize)) {
			return false
		}
		return bytes.HasSuffix(result, data)
	}
	if err := quick.Check(padLeftInvariants, nil); err != nil {
		t.Error(err)
	}
}

func TestPrependByte(t *testing.T) {
	prependInvariants := func(head byte, tail []byte) bool {
		result := slices.PrependByte(head, tail)
		if len(result) != len(tail)+1 {
			return false
		}
		if result[0] != head {
			return false
		}
		return reflect.DeepEqual(result[1:], tail)
	}
	if err := quick.Check(prependInvariants, nil); err != nil {
		t.Error(err)
	}
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
