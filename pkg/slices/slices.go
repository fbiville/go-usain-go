package slices

func PrependByte(head byte, tail []byte) []byte {
	return append([]byte{head}, tail...)
}
