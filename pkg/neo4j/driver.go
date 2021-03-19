package neo4j

import (
	"github.com/fbiville/go-usain-go/pkg/internal/bolt"
	"github.com/fbiville/go-usain-go/pkg/internal/packstream"
)

type Driver struct {
	connector    *bolt.Connector
}

func (d *Driver) Close() error {
	return d.connector.Close()
}

type AccessMode byte

const (
	ReadAccessMode AccessMode = iota
	WriteAccessMode
)

func (a AccessMode) String() string {
	return [...]string{"r", "w"}[a]
}

func NewDriver(host, username, password string) (connectionId *Driver, err error) {
	connector, err := bolt.NewConnector(host)
	if err != nil {
		return nil, err
	}
	err = connector.ShakeHands(bolt.NewVersion(4, 2))
	if err != nil {
		return nil, err
	}
	err = connector.SendHello(username, password)
	if err != nil {
		return nil, err
	}
	_, err = connector.ReceiveSuccess()
	if err != nil {
		return nil, err
	}
	return &Driver{
		connector:    connector,
	}, nil
}

func (d *Driver) Run(query string, accessMode AccessMode) (*packstream.List, error) {
	connector := d.connector
	err := connector.SendRun(query, accessMode.String())
	if err != nil {
		return nil, err
	}
	_, err = connector.ReceiveSuccess()
	if err != nil {
		return nil, err
	}
	record, err := connector.ReceiveRecord()
	if err != nil {
		return nil, err
	}
	_, err = connector.ReceiveSuccess()
	if err != nil {
		return nil, err
	}
	return record, nil
}
