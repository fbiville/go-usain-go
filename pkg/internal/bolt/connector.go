package bolt

import (
	"fmt"
	"github.com/fbiville/go-usain-go/pkg/internal/packstream"
	"net"
	"net/url"
)

const userAgent = "Go-usain/0.0.1"

type connector struct {
	chunker    *Chunker
	handshaker *Handshaker
	connection net.Conn
}

func (c *connector) Close() error {
	return c.connection.Close()
}

func NewConnector(host string) (*connector, error) {
	address := schemeless(host)
	connection, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &connector{
		connection: connection,
		chunker:    &Chunker{Connection: connection},
		handshaker: &Handshaker{connection: connection},
	}, nil
}

type serverVersion struct {
	major byte
	minor byte
}

func NewVersion(major, minor byte) *serverVersion {
	return &serverVersion{
		major: major,
		minor: minor,
	}
}

func (v *serverVersion) toByteArray() []byte {
	return []byte{0x0, 0x0, v.minor, v.major}
}

func (c *connector) ShakeHands(version *serverVersion) error {
	return c.handshaker.shakeHands(version.toByteArray())
}

func (c *connector) SayHello(username, password string) error {
	hello := newHelloMessage(username, password)
	return c.chunker.WriteChunked(hello.Pack())
}

func (c *connector) ReadSuccess() (*packstream.Structure, error) {
	response, err := c.chunker.ReadUnchunked()
	if err != nil {
		return nil, err
	}
	structure, err := packstream.UnpackStructure(response)
	if err != nil {
		return nil, err
	}
	if structure.Name() != "SUCCESS" {
		return nil, fmt.Errorf("expected SUCCESS but got %v\n", structure)
	}
	return structure, nil
}

func newHelloMessage(username string, password string) *packstream.Structure {
	agent := packstream.String(userAgent)
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

func schemeless(host string) string {
	uri, _ := url.Parse(host)
	port := uri.Port()
	if port == "" {
		port = "7687"
	}
	return fmt.Sprintf("%s:%s", uri.Hostname(), port)
}
