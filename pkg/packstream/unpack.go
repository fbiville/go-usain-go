package packstream

import (
	"encoding/binary"
	"fmt"
)

var Endianness = binary.BigEndian

func UnpackInteger(bytes []byte) (int64, error) {
	if len(bytes) == 0 {
		return 0, fmt.Errorf("at least 1 byte is expected")
	}
	if len(bytes) > 1 {
		marker := bytes[0]
		payload := bytes[1:]
		switch marker {
		case 0xC8:
			return tinyInt(payload), nil
		case 0xC9:
			return intSixteen(payload), nil
		case 0xCA:
			return intThirtyTwo(payload), nil
		case 0xCB:
			return intSixtyFour(payload), nil
		}

	}
	return tinyInt(bytes), nil
}

func UnpackStructure(bytes []byte) (*Structure, error) {
	fieldCount := bytes[0] - 0xB0
	result := Structure{TagByte: bytes[1]}
	if fieldCount == 0 {
		return &result, nil
	}
	dictionary, err := UnpackDictionary(bytes[2:]) // FIXME: assume single dictionary field
	if err != nil {
		return nil, err
	}
	result.Fields = []Value{dictionary} // FIXME: assumes single field and field is dictionary
	return &result, nil
}

func UnpackDictionary(bytes []byte) (*Dictionary, error) {
	entryCount := bytes[0] - 0xA0
	payload := bytes[1:]
	entries := make(map[string][]Value, entryCount)
	for i := byte(0); i < entryCount; i++ {
		var stringSize byte
		var key string
		var value String
		stringSize, payload = payload[0]-0x80, payload[1:]
		key, payload = string(payload[0:stringSize]), payload[stringSize:]
		stringSize, payload = payload[0]-0x80, payload[1:]
		value, payload = String(payload[0:stringSize]), payload[stringSize:]
		entries[key] = []Value{&value} // FIXME: assumes single value
	}
	result := Dictionary(entries)
	return &result, nil
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
