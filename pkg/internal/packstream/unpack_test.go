package packstream_test

import (
	"encoding/hex"
	"fmt"
	"github.com/fbiville/go-usain-go/pkg/internal/packstream"
	. "github.com/onsi/gomega"
	"math"
	"strings"
	"testing"
)

func TestUnpackInteger(t *testing.T) {
	RegisterTestingT(t)

	testCases := []struct {
		category string
		input    string
		result   int
	}{
		{"tiny_int", "80", math.MinInt8},
		{"tiny_int", "F0", -16},
		{"tiny_int", "FF", -1},
		{"tiny_int", "00", 0},
		{"tiny_int", "2A", 42},
		{"tiny_int", "7F", math.MaxInt8},
		{"int_8", "C8_80", math.MinInt8},
		{"int_8", "C8_F0", -16},
		{"int_8", "C8_FF", -1},
		{"int_8", "C8_00", 0},
		{"int_8", "C8_2A", 42},
		{"int_8", "C8_7F", math.MaxInt8},
		{"int_16", "C9_80_00", math.MinInt16},
		{"int_16", "C9_FF_F0", -16},
		{"int_16", "C9_FF_FF", -1},
		{"int_16", "C9_00_00", 0},
		{"int_16", "C9_00_2A", 42},
		{"int_16", "C9_7F_FF", math.MaxInt16},
		{"int_32", "CA_80_00_00_00", math.MinInt32},
		{"int_32", "CA_FF_FF_FF_F0", -16},
		{"int_32", "CA_FF_FF_FF_FF", -1},
		{"int_32", "CA_00_00_00_00", 0},
		{"int_32", "CA_00_00_00_2A", 42},
		{"int_32", "CA_7F_FF_FF_FF", math.MaxInt32},
		{"int_64", "CB_80_00_00_00_00_00_00_00", math.MinInt64},
		{"int_64", "CB_FF_FF_FF_FF_FF_FF_FF_F0", -16},
		{"int_64", "CB_FF_FF_FF_FF_FF_FF_FF_FF", -1},
		{"int_64", "CB_00_00_00_00_00_00_00_00", 0},
		{"int_64", "CB_00_00_00_00_00_00_00_2A", 42},
		{"int_64", "CB_7F_FF_FF_FF_FF_FF_FF_FF", math.MaxInt64},
	}

	for _, testCase := range testCases {
		input := testCase.input
		result := testCase.result
		t.Run(fmt.Sprintf("%q %s should be unpacked", input, testCase.category), func(t *testing.T) {
			integer, err := packstream.UnpackInteger(decodeHexa(sanitize(input)))
			Expect(err).NotTo(HaveOccurred())
			Expect(integer).To(Equal(int64(result)))
		})
	}
}

func TestUnpackIntegerInvalid(t *testing.T) {
	RegisterTestingT(t)

	_, err := packstream.UnpackInteger([]byte{})

	Expect(err).To(MatchError("at least 1 byte is expected"))
}

func TestUnpackDictionary(t *testing.T) {
	RegisterTestingT(t)

	testCases := []struct {
		input  []byte
		result *packstream.Dictionary
	}{
		{[]byte{0xA0}, dictionary()},
		{[]byte{0xA1, 0x81, 0x41, 0x81, 0x41}, dictionary("A", "A")},
	}

	for _, testCase := range testCases {
		input := testCase.input
		result := testCase.result
		t.Run("byte array should be packed", func(t *testing.T) {
			dictionary, err := packstream.UnpackDictionary(input)
			Expect(err).NotTo(HaveOccurred(), "dictionary should be unpacked")
			Expect(dictionary).To(Equal(result))
		})
	}
}

func TestUnpackStructure(t *testing.T) {
	RegisterTestingT(t)

	testCases := []struct {
		input  []byte
		result packstream.Structure
	}{
		{[]byte{0xB0, 0x7E}, packstream.Structure{TagByte: 0x7E}},
		{[]byte{0xB1, 0x70, 0xA1, 0x81, byte('K'), 0x81, byte('V')}, packstream.Structure{
			TagByte: 0x70,
			Fields:  []packstream.Value{dictionary("K", "V")}}},
	}

	for i, testCase := range testCases {
		input := testCase.input
		result := testCase.result
		t.Run(fmt.Sprintf("byte array #%d should be unpacked", i+1), func(t *testing.T) {
			structure, err := packstream.UnpackStructure(input)
			Expect(err).NotTo(HaveOccurred(), "structure should be unpacked")
			Expect(*structure).To(Equal(result))
		})
	}
}

func sanitize(input string) string {
	return strings.ReplaceAll(input, "_", "")
}

func decodeHexa(s string) []byte {
	result, err := hex.DecodeString(s)
	Expect(err).NotTo(HaveOccurred(),
		"hexadecimal string %q should be decoded to byte array", s)
	return result
}
