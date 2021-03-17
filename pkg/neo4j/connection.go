package neo4j

import (
	"github.com/fbiville/go-usain-go/pkg/internal/bolt"
	"github.com/fbiville/go-usain-go/pkg/internal/closers"
	"github.com/fbiville/go-usain-go/pkg/internal/packstream"
)

func Connect(host, username, password string) (connectionId string, err error) {
	connector, err := bolt.NewConnector(host)
	if err != nil {
		return "", nil
	}
	defer func() {
		err = closers.SafeClose(connector, err)
	}()
	err = connector.ShakeHands(bolt.NewVersion(4, 2)) // 🤜🤛
	if err != nil {
		return "", err
	}
	err = connector.SayHello(username, password) // 👋
	if err != nil {
		return "", err
	}
	success, err := connector.ReadSuccess() // 🎉
	if err != nil {
		return "", err
	}
	return getConnectionId(success), nil // 🆔
}

func getConnectionId(success *packstream.Structure) string {
	entries := success.Fields[0].(*packstream.Dictionary)
	connectionIdValue := (*entries)["connection_id"][0]
	return string(*connectionIdValue.(*packstream.String))
}
