package neo4j_test

import (
	"context"
	"fmt"
	"github.com/fbiville/go-usain-go/pkg/neo4j"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
)

const username = "neo4j"
const password = "v3rys3cur3"

func TestConnection(t *testing.T) {
	RegisterTestingT(t)

	ctx := context.Background()
	container, err := startContainer(ctx, username, password)
	Expect(err).NotTo(HaveOccurred(), "container should start")
	defer func() {
		Expect(container.Terminate(ctx)).
			NotTo(HaveOccurred(), "container should shut down")
	}()

	port, err := container.MappedPort(ctx, "7687")
	Expect(err).NotTo(HaveOccurred(), "container should return mapped port")
	address := fmt.Sprintf("bolt://localhost:%d", port.Int())

	connectionId, err := neo4j.Connect(address, username, password)

	Expect(err).NotTo(HaveOccurred(), "connection should work")
	Expect(connectionId).NotTo(BeEmpty())
}

func startContainer(ctx context.Context, username, password string) (testcontainers.Container, error) {
	request := testcontainers.ContainerRequest{
		Image:        "neo4j:4.2",
		ExposedPorts: []string{"7687/tcp"},
		Env:          map[string]string{
			"NEO4J_AUTH": fmt.Sprintf("%s/%s", username, password),
			"NEO4J_dbms_logs_debug_level": "DEBUG",
		},
		WaitingFor:   wait.ForLog("Bolt enabled"),
		ReaperImage:  "testcontainers/ryuk",
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})
}
