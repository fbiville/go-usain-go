package packstream

import (
	"encoding/binary"
	"fmt"
	"github.com/fbiville/go-usain-go/pkg/internal/slices"
	"math"
)

var Endianness = binary.BigEndian

type mapperFunc func([]byte) (Value, int, error)

func UnpackValue(bytes []byte) (Value, int, error) {
	if len(bytes) == 0 {
		return nil, 0, fmt.Errorf("data to unpack must be at 1 byte long")
	}
	return mapper(bytes[0])(bytes)
}

func mapper(marker byte) mapperFunc {
	switch {
	case marker <= 0x7F || 0xC8 <= marker && marker <= 0xCB || 0xF0 <= marker:
		return unpackInteger
	case 0x80 <= marker && marker <= 0x8F || 0xD0 <= marker && marker <= 0xD2:
		return unpackString
	case 0x90 <= marker && marker <= 0x9F: // FIXME support large lists
		return unpackList
	case 0xA0 <= marker && marker <= 0xAF: // FIXME support large dictionaries
		return unpackDictionary
	case 0xB0 <= marker && marker <= 0xBF: // FIXME support large structures
		return unpackStructure
	case 0xC0 == marker:
		return unpackNil
	// FIXME support booleans
	// FIXME support float
	default:
		return unsupportedMarkerFunc()
	}
}

func unsupportedMarkerFunc() mapperFunc {
	return func(bytes []byte) (Value, int, error) {
		return nil, 1, fmt.Errorf("unsupported marker %X", bytes[0])
	}
}

func unpackStructure(bytes []byte) (Value, int, error) {
	fieldCount := bytes[0] - 0xB0
	result := Structure{TagByte: bytes[1]}
	readBytes := 2
	if fieldCount == 0 {
		return &result, readBytes, nil
	}
	bytes = bytes[readBytes:]
	for {
		if len(bytes) == 0 {
			break
		}
		value, n, err := UnpackValue(bytes)
		if err != nil {
			return nil, readBytes, err
		}
		bytes = bytes[n:]
		readBytes += n
		result.Fields = append(result.Fields, value)
	}
	return &result, readBytes, nil
}

func unpackDictionary(bytes []byte) (Value, int, error) {
	entryCount := bytes[0] - 0xA0
	readByteCount := 1
	payload := bytes[1:]
	entries := make(map[string][]Value, entryCount)
	for i := byte(0); i < entryCount; i++ {
		rawKey, n, err := unpackString(payload)
		if err != nil {
			return nil, readByteCount, err
		}
		readByteCount += n
		payload = payload[n:]
		value, n, err := UnpackValue(payload)
		if err != nil {
			return nil, readByteCount, err
		}
		readByteCount += n
		payload = payload[n:]
		entries[string(*(rawKey.(*String)))] = []Value{value} // FIXME: should not overwrite previous value
	}
	result := Dictionary(entries)
	return &result, readByteCount, nil
}

func unpackNil(bytes []byte) (Value, int, error) {
	value := bytes[0]
	if value != 0xC0 {
		return nil, 1, fmt.Errorf("invalid value %X for nil", value)
	}
	return NilInstance(), 1, nil
}

func unpackInteger(bytes []byte) (Value, int, error) {
	marker := bytes[0]
	switch marker {
	case 0xC8:
		return Integer(tinyInt(bytes[1:2])), 2, nil
	case 0xC9:
		return Integer(intSixteen(bytes[1:3])), 3, nil
	case 0xCA:
		return Integer(intThirtyTwo(bytes[1:5])), 5, nil
	case 0xCB:
		return Integer(intSixtyFour(bytes[1:9])), 9, nil
	default:
		return Integer(tinyInt(bytes[0:1])), 1, nil
	}
}

func unpackString(payload []byte) (Value, int, error) {
	offset, size := readStringSize(payload)
	if size == 0 {
		result := String("")
		return &result, offset, nil
	}
	end := offset + int(size) // FIXME
	result := String(payload[offset :end])
	return &result, end, nil
}

func readStringSize(payload []byte) (int, uint32) {
	marker := payload[0]
	if marker <= 0x8F {
		return 1, uint32(marker - 0x80)
	}
	offset := 1 + int(math.Pow(2, float64(marker-0XD0)))
	rawSize := slices.PadLeft(payload[1:offset], 0, 4)
	size := Endianness.Uint32(rawSize)
	return offset, size
}

func unpackList(bytes []byte) (Value, int, error) {
	var result List
	readByteCount := 1
	marker := bytes[0]
	count := marker - 0x90
	bytes = bytes[1:]
	for i := byte(0); i < count; i++ {
		value, n, err := UnpackValue(bytes)
		readByteCount += n
		if err != nil {
			return nil, readByteCount, fmt.Errorf("could not read list entry number %d: %w", i+1, err)
		}
		result = append(result, value)
		bytes = bytes[n:]
	}
	return &result, readByteCount, nil
}

func intSixtyFour(bytes []byte) int64 {
	return int64(Endianness.Uint64(bytes))
}

func intThirtyTwo(bytes []byte) int64 {
	return int64(int32(Endianness.Uint32(bytes)))
}

func intSixteen(bytes []byte) int64 {
	return int64(int16(Endianness.Uint16(bytes)))
}

func tinyInt(bytes []byte) int64 {
	return int64(int8(bytes[0]))
}
