package bolt

import (
	"fmt"
	"github.com/fbiville/go-usain-go/pkg/internal/packstream"
	"net"
	"net/url"
	"time"
)

const userAgent = "Go-usain/0.0.1"

type Connector struct {
	chunker    *Chunker
	handshaker *Handshaker
	connection net.Conn
}

func (c *Connector) Close() error {
	return c.connection.Close()
}

func NewConnector(host string) (*Connector, error) {
	address := schemeless(host)
	connection, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &Connector{
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

func (c *Connector) ShakeHands(version *serverVersion) error {
	return c.handshaker.shakeHands(version.toByteArray())
}

func (c *Connector) SendHello(username, password string) error {
	hello := newHelloMessage(username, password)
	return c.chunker.WriteChunked(hello.Pack())
}

func (c *Connector) ReceiveSuccess() (*packstream.Structure, error) {
	response, err := c.chunker.ReadUnchunked()
	if err != nil {
		return nil, err
	}
	value, _, err := packstream.UnpackValue(response)
	if err != nil {
		return nil, err
	}
	structure, casted := value.(*packstream.Structure)
	if !casted {
		return nil, fmt.Errorf("expected structure but got %v\n", value)
	}
	if structure.Name() != "SUCCESS" {
		return nil, fmt.Errorf("expected SUCCESS but got %v\n", structure)
	}
	return structure, nil
}
func (c *Connector) SendRun(query string, accessMode string) error {
	run := newRunMessage(query, accessMode)
	pull := newPullMessage(1000)
	return c.chunker.WriteChunked(run.Pack(), pull.Pack())
}

func (c *Connector) ReceiveRecord() (*packstream.List, error) {
	response, err := c.chunker.ReadUnchunked()
	if err != nil {
		return nil, err
	}
	value, _, err := packstream.UnpackValue(response)
	if err != nil {
		return nil, err
	}
	structure, casted := value.(*packstream.Structure)
	if !casted {
		return nil, fmt.Errorf("expected structure but got %v\n", value)
	}
	if structure.Name() != "RECORD" {
		return nil, fmt.Errorf("expected RECORD, got %v response", structure)
	}
	return structure.Fields[0].(*packstream.List), nil
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

const transactionTimeout = time.Second * 30

func newRunMessage(query string, accessMode string) *packstream.Structure {
	queryValue := packstream.String(query)
	accessModeValue := packstream.String(accessMode)
	return &packstream.Structure{
		TagByte: 0x10,
		Fields: []packstream.Value{
			&queryValue,
			&packstream.Dictionary{},
			&packstream.Dictionary{
				"bookmarks":   []packstream.Value{&packstream.List{}},
				"tx_timeout":  []packstream.Value{packstream.Integer(transactionTimeout.Milliseconds())},
				"tx_metadata": []packstream.Value{&packstream.Dictionary{}},
				"mode":        []packstream.Value{&accessModeValue},
			},
		},
	}
}

func newPullMessage(i int) *packstream.Structure {
	return &packstream.Structure{
		TagByte: 0x3F,
		Fields: []packstream.Value{
			&packstream.Dictionary{
				"n": []packstream.Value{packstream.Integer(i)},
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
