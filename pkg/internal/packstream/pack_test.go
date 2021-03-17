package packstream_test

import (
	"fmt"
	"github.com/fbiville/go-usain-go/pkg/internal/packstream"
	"github.com/fbiville/go-usain-go/pkg/internal/slices"
	. "github.com/onsi/gomega"
	"testing"
)

func TestPackNil(t *testing.T) {
	RegisterTestingT(t)
	nilValue := &packstream.Nil{}
	Expect(nilValue.Pack()).To(Equal([]byte{0xC0}))
}

func TestPackString(t *testing.T) {
	RegisterTestingT(t)

	testCases := []struct {
		input  string
		result []byte
	}{
		{"", []byte{0x80}},
		{"A", []byte{0x81, 0x41}},
		{"plsfitin15bytes", slices.PrependByte(0x8F, bytesOf("plsfitin15bytes"))},
	}

	for _, testCase := range testCases {
		input := testCase.input
		result := testCase.result
		t.Run(fmt.Sprintf("%q string should be packed", input), func(t *testing.T) {
			str := packstream.String(input)
			Expect(str.Pack()).To(Equal(result))
		})
	}
}

func TestPackDictionary(t *testing.T) {
	RegisterTestingT(t)

	testCases := []struct {
		input  *packstream.Dictionary
		result []byte
	}{
		{dictionary(), []byte{0xA0}},
		{dictionary("one", "eins"), []byte{0xA1, 0x83, 0x6F, 0x6E, 0x65, 0x84, 0x65, 0x69, 0x6E, 0x73}},
		{nilValueDictionary("one"), []byte{0xA1, 0x83, 0x6F, 0x6E, 0x65, 0xC0}},
	}

	for _, testCase := range testCases {
		input := testCase.input
		result := testCase.result
		t.Run(fmt.Sprintf("%q dictionary should be packed", input), func(t *testing.T) {
			kfeorp := input.Pack()
			Expect(kfeorp).To(Equal(result))
		})
	}
}

func TestPackStructure(t *testing.T) {
	RegisterTestingT(t)

	testCases := []struct {
		tagByte byte
		fields  []packstream.Value
		result  []byte
	}{
		{0x02, []packstream.Value{}, []byte{0xB0, 0x02}},
		{0x01, []packstream.Value{dictionary("A", "A", "B", "B")}, []byte{0xB1, 0x01, 0xA2, 0x81, 0x41, 0x81, 0x41, 0x81, 0x42, 0x81, 0x42}},
	}

	for _, testCase := range testCases {
		result := testCase.result
		dictionaryField := testCase.fields
		t.Run(fmt.Sprintf("%s structure should be packed", dictionaryField), func(t *testing.T) {
			structure := packstream.Structure{
				TagByte: testCase.tagByte,
				Fields:  dictionaryField,
			}
			Expect(structure.Pack()).To(Equal(result))
		})
	}
}

func dictionary(keyValuePairs ...string) *packstream.Dictionary {
	entries := make(map[string][]packstream.Value, len(keyValuePairs)/2)
	for i := 0; i < len(keyValuePairs)-1; i += 2 {
		key := keyValuePairs[i]
		value := packstream.String(keyValuePairs[i+1])
		entries[key] = []packstream.Value{&value}
	}
	result := packstream.Dictionary(entries)
	return &result
}

func nilValueDictionary(key string) *packstream.Dictionary {
	result := packstream.Dictionary(map[string][]packstream.Value{
		key: nil,
	})
	return &result
}

func bytesOf(s string) []byte {
	return []byte(s)
}
