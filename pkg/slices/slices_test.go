package slices_test

import (
	"github.com/fbiville/go-usain-go/pkg/slices"
	"reflect"
	"testing"
	"testing/quick"
)

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
