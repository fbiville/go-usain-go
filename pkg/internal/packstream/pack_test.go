package packstream_test

import (
	"fmt"
	"github.com/fbiville/go-usain-go/pkg/internal/packstream"
	"github.com/fbiville/go-usain-go/pkg/internal/slices"
	. "github.com/onsi/gomega"
	"math"
	"testing"
)

func TestPackNil(t *testing.T) {
	RegisterTestingT(t)
	nilValue := &packstream.Nil{}
	Expect(nilValue.Pack()).To(Equal([]byte{0xC0}))
}

func TestPackInteger(t *testing.T) {
	RegisterTestingT(t)

	testCases := []struct {
		category string
		result   string
		input    int
	}{
		{category: "tiny_int", result: "F0", input: -16},
		{category: "tiny_int", result: "7F", input: math.MaxInt8},
		{category: "int_8", result: "C8_EF", input: -17},
		{category: "int_8", result: "C8_80", input: math.MinInt8},
		{category: "int_16", result: "C9_80_00", input: math.MinInt16},
		{category: "int_16", result: "C9_FF_7F", input: math.MinInt8 - 1},
		{category: "int_16", result: "C9_00_80", input: math.MaxInt8 + 1},
		{category: "int_16", result: "C9_7F_FF", input: math.MaxInt16},
		{category: "int_32", result: "CA_80_00_00_00", input: math.MinInt32},
		{category: "int_32", result: "CA_FF_FF_7F_FF", input: math.MinInt16 - 1},
		{category: "int_32", result: "CA_00_00_80_00", input: math.MaxInt16 + 1},
		{category: "int_32", result: "CA_7F_FF_FF_FF", input: math.MaxInt32},
		{category: "int_64", result: "CB_80_00_00_00_00_00_00_00", input: math.MinInt64},
		{category: "int_64", result: "CB_FF_FF_FF_FF_7F_FF_FF_FF", input: math.MinInt32 - 1},
		{category: "int_64", result: "CB_00_00_00_00_80_00_00_00", input: math.MaxInt32 + 1},
		{category: "int_64", result: "CB_7F_FF_FF_FF_FF_FF_FF_FF", input: math.MaxInt64},
	}

	for _, testCase := range testCases {
		input := testCase.input
		result := testCase.result
		t.Run(fmt.Sprintf("%q %d should be packed", testCase.category, input), func(t *testing.T) {
			integer := packstream.Integer(input)
			Expect(integer.Pack()).To(Equal(decodeHexa(sanitize(result))))
		})
	}
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

func TestPackList(t *testing.T) {
	RegisterTestingT(t)

	value1 := packstream.String("A")
	value2 := packstream.Integer(1)

	testCases := []struct {
		input  []packstream.Value
		result []byte
	}{
		{[]packstream.Value{}, []byte{0x90}},
		{[]packstream.Value{&value1}, []byte{0x91, 0x81, 0x41}},
		{[]packstream.Value{&value2, &value1}, []byte{0x92, 0x01, 0x81, 0x41}},
	}

	for _, testCase := range testCases {
		input := testCase.input
		result := testCase.result
		t.Run(fmt.Sprintf("%q string should be packed", input), func(t *testing.T) {
			list := packstream.List(input)
			Expect(list.Pack()).To(Equal(result))
		})
	}
}

func TestPackDictionary(t *testing.T) {
	RegisterTestingT(t)

	value := packstream.String("eins")
	testCases := []struct {
		input  *packstream.Dictionary
		result []byte
	}{
		{dictionary(), []byte{0xA0}},
		{dictionary("one", &value), []byte{0xA1, 0x83, 0x6F, 0x6E, 0x65, 0x84, 0x65, 0x69, 0x6E, 0x73}},
		{nilValueDictionary("one"), []byte{0xA1, 0x83, 0x6F, 0x6E, 0x65, 0xC0}},
	}

	for _, testCase := range testCases {
		input := testCase.input
		expectedResult := testCase.result
		t.Run(fmt.Sprintf("%q dictionary should be packed", input), func(t *testing.T) {
			result := input.Pack()

			Expect(result).To(Equal(expectedResult))
		})
	}
}

func TestPackStructure(t *testing.T) {
	RegisterTestingT(t)

	value1 := packstream.String("A")
	value2 := packstream.String("B")
	testCases := []struct {
		tagByte byte
		fields  []packstream.Value
		result  []byte
	}{
		{0x02, []packstream.Value{}, []byte{0xB0, 0x02}},
		{0x01, []packstream.Value{dictionary("A", &value1, "B", &value2)}, []byte{0xB1, 0x01, 0xA2, 0x81, 0x41, 0x81, 0x41, 0x81, 0x42, 0x81, 0x42}},
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

func dictionary(keyValuePairs ...interface{}) *packstream.Dictionary {
	entries := make(map[string][]packstream.Value, len(keyValuePairs)/2)
	for i := 0; i < len(keyValuePairs)-1; i += 2 {
		key := keyValuePairs[i].(string)
		value := keyValuePairs[i+1].(packstream.Value)
		entries[key] = []packstream.Value{value}
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
