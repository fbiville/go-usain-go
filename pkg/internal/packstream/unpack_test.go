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

func TestUnpackNil(t *testing.T) {
	RegisterTestingT(t)

	value, n, err := packstream.UnpackValue([]byte{0xC0})

	Expect(err).NotTo(HaveOccurred())
	Expect(n).To(Equal(1), "should read 1 byte")
	Expect(value).To(Equal(packstream.NilInstance()))
}

func TestUnpackInteger(t *testing.T) {
	RegisterTestingT(t)

	testCases := []struct {
		category string
		input    string
		result   int
	}{
		{"tiny_int", "F0", -16},
		{"tiny_int", "FF", -1},
		{"tiny_int", "00", 0},
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
			inputPayload := decodeHexa(sanitize(input))

			integer, n, err := packstream.UnpackValue(inputPayload)

			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(len(inputPayload)), "should read %d bytes", len(inputPayload))
			Expect(integer).To(Equal(packstream.Integer(result)))
		})
	}
}

func TestUnpackInvalidValue(t *testing.T) {
	RegisterTestingT(t)

	_, _, err := packstream.UnpackValue([]byte{})

	Expect(err).To(MatchError("data to unpack must be at 1 byte long"))
}

func TestUnpackUnknownMarker(t *testing.T) { // FIXME: should be gone once all types and value ranges are supported
	RegisterTestingT(t)

	_, _, err := packstream.UnpackValue([]byte{0xDF})

	Expect(err).To(MatchError("unsupported marker DF"))
}

func TestUnpackString(t *testing.T) {
	RegisterTestingT(t)

	testCases := []struct {
		input  string
		result packstream.String
	}{
		{"80", ""},
		{"81_41", "A"},
		{"D0_01_41", "A"},
		{"D1_00_01_41", "A"},
		{"D2_00_00_00_01_41", "A"},
	}

	for _, testCase := range testCases {
		input := testCase.input
		result := testCase.result

		t.Run(fmt.Sprintf("%s should be unpacked", input), func(t *testing.T) {
			inputPayload := decodeHexa(sanitize(input))

			str, n, err := packstream.UnpackValue(inputPayload)

			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(len(inputPayload)), "should read %d bytes", len(inputPayload))
			Expect(*str.(*packstream.String)).To(Equal(result))
		})
	}
}

func TestUnpackList(t *testing.T) {
	RegisterTestingT(t)

	value := packstream.String("A")
	testCases := []struct {
		input  string
		result packstream.List
	}{
		{"90", nil},
		{"92_81_41_C0", packstream.List([]packstream.Value{&value, packstream.NilInstance()})},
	}

	for _, testCase := range testCases {
		input := testCase.input
		result := testCase.result

		t.Run(fmt.Sprintf("%s should be unpacked", input), func(t *testing.T) {
			inputPayload := decodeHexa(sanitize(input))

			list, n, err := packstream.UnpackValue(inputPayload)

			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(len(inputPayload)), "should read %d bytes", len(inputPayload))
			Expect(*list.(*packstream.List)).To(Equal(result))
		})
	}
}

func TestUnpackDictionary(t *testing.T) {
	RegisterTestingT(t)

	stringValue := packstream.String("A")
	testCases := []struct {
		input  []byte
		result *packstream.Dictionary
	}{
		{[]byte{0xA0}, dictionary()},
		{[]byte{0xA1, 0x81, byte('A'), 0x81, byte('A')}, dictionary("A", &stringValue)},
		{[]byte{0xA1, 0x81, byte('A'), 0xCA, 0, 0, 0, 0x2A}, dictionary("A", packstream.Integer(42))},
	}

	for i, testCase := range testCases {
		input := testCase.input
		result := testCase.result
		t.Run(fmt.Sprintf("byte array #%d should be packed", i+1), func(t *testing.T) {
			dictionary, n, err := packstream.UnpackValue(input)

			Expect(err).NotTo(HaveOccurred(), "dictionary should be unpacked")
			Expect(n).To(Equal(len(input)), "should read %d bytes", len(input))
			Expect(dictionary).To(Equal(result))
		})
	}
}

func TestUnpackStructure(t *testing.T) {
	RegisterTestingT(t)

	value := packstream.String("V")
	testCases := []struct {
		input  []byte
		result *packstream.Structure
	}{
		{[]byte{0xB0, 0x7E}, &packstream.Structure{TagByte: 0x7E}},
		{[]byte{0xB1, 0x70, 0xA1, 0x81, byte('K'), 0x81, byte('V')}, &packstream.Structure{
			TagByte: 0x70,
			Fields:  []packstream.Value{dictionary("K", &value)}}},
		{[]byte{0xB2, 0x70, 0xA1, 0x81, byte('K'), 0x81, byte('V'), 0x7F}, &packstream.Structure{
			TagByte: 0x70,
			Fields:  []packstream.Value{
				dictionary("K", &value),
				packstream.Integer(127),
			}}},
	}

	for i, testCase := range testCases {
		input := testCase.input
		result := testCase.result
		t.Run(fmt.Sprintf("byte array #%d should be unpacked", i+1), func(t *testing.T) {
			structure, n, err := packstream.UnpackValue(input)

			Expect(err).NotTo(HaveOccurred(), "structure should be unpacked")
			Expect(n).To(Equal(len(input)), "should read %d bytes", len(input))
			Expect(structure).To(Equal(result))
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
