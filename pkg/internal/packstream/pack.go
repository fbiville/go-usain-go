package packstream

import (
	"fmt"
	"github.com/fbiville/go-usain-go/pkg/internal/slices"
	"math"
	"sort"
	"strings"
	"sync"
)

type Value interface {
	Pack() []byte
	String() string
}

var once sync.Once
var singleNilInstance *Nil

func NilInstance() *Nil {
	once.Do(func() {
		singleNilInstance = &Nil{}
	})
	return singleNilInstance
}

type Nil struct{}

func (n *Nil) Pack() []byte {
	return []byte{0xC0}
}

func (n *Nil) String() string {
	return "<nil>"
}

type Integer int

func (i Integer) Pack() []byte {
	if -16 <= i && i <= math.MaxInt8 {
		return []byte{byte(i)}
	}
	if math.MinInt8 <= i && i <= -17 {
		return []byte{0xC8, byte(i)}
	}
	if math.MinInt16 <= i && i <= math.MinInt8-1 || math.MaxInt8+1 <= i && i <= math.MaxInt16 {
		result := make([]byte, 3)
		result[0] = 0xC9
		Endianness.PutUint16(result[1:], uint16(i))
		return result
	}
	if math.MinInt32 <= i && i <= math.MinInt16-1 || math.MaxInt16+1 <= i && i <= math.MaxInt32 {
		result := make([]byte, 5)
		result[0] = 0xCA
		Endianness.PutUint32(result[1:], uint32(i))
		return result
	}
	if math.MinInt64 <= i && i <= math.MinInt32-1 || math.MaxInt32+1 <= i && i <= math.MaxInt64 {
		result := make([]byte, 9)
		result[0] = 0xCB
		Endianness.PutUint64(result[1:], uint64(i))
		return result
	}
	panic("only 64-bit value range are supported")
}

func (i Integer) String() string {
	return fmt.Sprintf("%d", i)
}

type String string

func (s *String) Pack() []byte {
	str := string(*s)
	marker := 0x80 + byte(len(str))
	stringBytes := []byte(str)
	return append([]byte{marker}, stringBytes...)
}

func (s *String) String() string {
	return fmt.Sprintf("%q", string(*s))
}

type List []Value

func (l *List) Pack() []byte {
	marker := byte(0x90 + len(*l)) // FIXME: does not support max value range
	var result []byte
	for _, value := range *l {
		result = append(result, value.Pack()...)
	}
	return slices.PrependByte(marker, result)
}

func (l *List) String() string {
	builder := strings.Builder{}
	builder.WriteString("[")
	values := []Value(*l)
	for i, value := range values {
		builder.WriteString(value.String())
		if i < len(values)-1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString("]")
	return builder.String()
}

type Dictionary map[string][]Value

func (d *Dictionary) Pack() []byte {
	var result []byte
	marker := 0xA0 + d.Length()
	dictionary := d.asMap()
	keys := d.sortedKeys(dictionary)
	for _, keyName := range keys {
		key := String(keyName)
		values := dictionary[keyName]
		if values == nil {
			result = append(result, key.Pack()...)
			result = append(result, NilInstance().Pack()...)
		} else {
			for _, value := range values {
				result = append(result, key.Pack()...)
				result = append(result, value.Pack()...)
			}
		}
	}
	return slices.PrependByte(byte(marker), result)
}

func (d *Dictionary) Length() int {
	result := 0
	dictionary := d.asMap()
	for _, values := range dictionary {
		increment := len(values)
		if increment == 0 {
			increment = 1
		}
		result += increment
	}
	return result
}

func (d *Dictionary) String() string {
	str := strings.Builder{}
	str.WriteString("{")
	dictionary := d.asMap()
	keys := d.sortedKeys(dictionary)
	for _, key := range keys {
		for _, value := range dictionary[key] {
			str.WriteString(fmt.Sprintf("\t%q: %s,", key, value.String()))
		}
	}
	str.WriteString("}")
	return str.String()
}

func (d *Dictionary) asMap() map[string][]Value {
	return *d
}

func (d *Dictionary) sortedKeys(dictionary map[string][]Value) []string {
	keys := make([]string, 0, len(dictionary))
	for k := range dictionary {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

type Structure struct {
	TagByte byte
	Fields  []Value
}

func (s *Structure) Pack() []byte {
	fieldCount := len(s.Fields)
	marker := 0xB0 + fieldCount
	preamble := []byte{byte(marker), s.TagByte}
	result := preamble
	for _, field := range s.Fields {
		result = append(result, field.Pack()...)
	}
	return result
}

func (s *Structure) Name() string {
	return structureNames()[s.TagByte]
}

func (s *Structure) String() string {
	name := fmt.Sprintf("<%d>", s.TagByte)
	if friendlyName, found := structureNames()[s.TagByte]; found {
		name = friendlyName
	}
	result := strings.Builder{}
	result.WriteString(name)
	result.WriteString("(\n")
	for _, field := range s.Fields {
		result.WriteString(fmt.Sprintf("\t%s\n", field.String()))
	}
	result.WriteString(")")
	return result.String()
}

func structureNames() map[byte]string {
	return map[byte]string{
		0x4E: "NODE",
		0x52: "RELATIONSHIP",
		0x72: "UNBOUND_RELATIONSHIP",
		0x50: "PATH",
		0x44: "DATE",
		0x54: "TIME",
		0x74: "LOCALTIME",
		0x46: "DATETIME",
		0x66: "DATETIME_ZONE_ID",
		0x64: "LOCAL_DATETIME",
		0x45: "DURATION",
		0x58: "POINT_2D",
		0x59: "POINT_3D",
		0x01: "HELLO",
		0x02: "GOODBYE",
		0x0F: "RESET",
		0x10: "RUN",
		0x2F: "DISCARD",
		0x3F: "PULL",
		0x11: "BEGIN",
		0x12: "COMMIT",
		0x13: "ROLLBACK",
		0x70: "SUCCESS",
		0x7E: "IGNORED",
		0x7F: "FAILURE",
		0x71: "RECORD",
	}
}
