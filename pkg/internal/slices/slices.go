package slices

import "bytes"

//FIXME: works only on small byte arrays to avoid panic
func PadLeft(data []byte, symbol, targetSize byte) []byte {
	padding := int(targetSize) - len(data)
	if padding <= 0 {
		return data
	}
	result := make([]byte, targetSize)
	copy(result[0:padding], bytes.Repeat([]byte{symbol}, padding))
	copy(result[padding:], data)
	return result
}

func PrependByte(head byte, tail []byte) []byte {
	return append([]byte{head}, tail...)
}
