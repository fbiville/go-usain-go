package neo4j_test

import (
	"context"
	"fmt"
	"github.com/fbiville/go-usain-go/pkg/internal/packstream"
	"github.com/fbiville/go-usain-go/pkg/neo4j"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
)

const username = "neo4j"
const password = "v3rys3cur3"

func TestDriver(t *testing.T) {
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
	session, err := neo4j.NewDriver(address, username, password)
	defer func() {
		Expect(session.Close()).To(Succeed())
	}()

	t.Run("open driver", func(t *testing.T) {
		Expect(err).NotTo(HaveOccurred(), "driver should connect")
	})

	t.Run("run simple query", func(t *testing.T) {
		result, err := session.Run("RETURN 42", neo4j.ReadAccessMode)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(&packstream.List{packstream.Integer(42)}))
	})
}

func startContainer(ctx context.Context, username, password string) (testcontainers.Container, error) {
	request := testcontainers.ContainerRequest{
		Image:        "neo4j:4.2",
		ExposedPorts: []string{"7687/tcp"},
		Env: map[string]string{
			"NEO4J_AUTH":                  fmt.Sprintf("%s/%s", username, password),
			"NEO4J_dbms_logs_debug_level": "DEBUG",
		},
		WaitingFor:  wait.ForLog("Bolt enabled"),
		ReaperImage: "testcontainers/ryuk",
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})
}
