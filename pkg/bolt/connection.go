package bolt

import (
	"bytes"
	"encoding/binary"
	fmt "fmt"
	"github.com/fbiville/go-usain-go/pkg/closers"
	"github.com/fbiville/go-usain-go/pkg/packstream"
	"net"
	"net/url"
)

func Connect(host, username, password string) (connectionId string, err error) {
	address := schemeless(host)
	connection, err := net.Dial("tcp", address)
	if err != nil {
		return "", err
	}
	defer func() {
		err = closers.SafeClose(connection, err)
	}()
	err = shakeHands(connection, []byte{0x0, 0x0, 0x2, 0x4})
	if err != nil {
		return "", err
	}
	helloMessage, err := Chunk(hello(username, password).Pack())
	if err != nil {
		return "", err
	}
	_, err = connection.Write(helloMessage)
	if err != nil {
		return "", err
	}
	responseChunk := make([]byte, 2)
	err = binary.Read(connection, packstream.Endianness, responseChunk)
	if err != nil {
		return "", err
	}
	if bytes.Equal(responseChunk, []byte{0x0, 0x0}) {
		return "", fmt.Errorf("could not read response chunk")
	}
	responseSize := packstream.Endianness.Uint16(responseChunk)
	response := make([]byte, responseSize)
	err = binary.Read(connection, packstream.Endianness, response)
	if err != nil {
		return "", err
	}
	structure, err := packstream.UnpackStructure(response)
	if err != nil {
		return "", err
	}
	if structure.Name() != "SUCCESS" {
		return "", fmt.Errorf("expected SUCCESS but got %v\n", structure)
	}
	entries := structure.Fields[0].(*packstream.Dictionary) // FIXME: clean this up
	connectionIdValue := (*entries)["connection_id"][0]
	connectionId = string(*connectionIdValue.(*packstream.String))
	return connectionId, nil
}

func shakeHands(connection net.Conn, supportedVersion []byte) error {
	_, err := connection.Write([]byte{0x60, 0x60, 0xB0, 0x17})
	if err != nil {
		return fmt.Errorf("could not send handshake preamble %w", err)
	}
	_, err = connection.Write(append(
		supportedVersion,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
	))
	if err != nil {
		return fmt.Errorf("could not send handshake supported versions %w", err)
	}
	response := []byte{0x0, 0x0, 0x0, 0x0}
	err = binary.Read(connection, packstream.Endianness, response)
	if err != nil {
		return fmt.Errorf("could not receive handshake response %w", err)
	}
	responseLength := len(response)
	if responseLength != 4 {
		return fmt.Errorf("expected handshake response to be 4 byte long, got %d", responseLength)
	}
	if !bytes.Equal(response, supportedVersion) {
		return fmt.Errorf("expected \"%d.%d\", but got \"%d.%d\" version",
			supportedVersion[2], supportedVersion[3],
			response[2], response[3])
	}
	return nil
}

func schemeless(host string) string {
	uri, _ := url.Parse(host)
	port := uri.Port()
	if port == "" {
		port = "7687"
	}
	return fmt.Sprintf("%s:%s", uri.Hostname(), port)
}

func hello(username, password string) *packstream.Structure {
	agent := packstream.String("Go-usain/0.0.1")
	scheme := packstream.String("basic")
	principal := packstream.String(username)
	credentials := packstream.String(password)
	return &packstream.Structure{
		TagByte: 0x01,
		Fields: []packstream.Value{
			&packstream.Dictionary{
				"user_agent":  []packstream.Value{&agent},
				"scheme":      []packstream.Value{&scheme},
				"principal":   []packstream.Value{&principal},
				"credentials": []packstream.Value{&credentials},
			},
		},
	}
}
