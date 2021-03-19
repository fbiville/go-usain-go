package main

import (
	"flag"
	"fmt"
	"github.com/fbiville/go-usain-go/pkg/neo4j"
)

func main() {
	uri := flag.String("uri", "bolt://localhost", "Neo4j URI (e.g.: bolt://localhost)")
	username := flag.String("username", "neo4j", "Neo4j username (e.g.: neo4j")
	password := flag.String("password", "", "Neo4j password (e.g.: s3cr3t")

	flag.Parse()

	runExample(*uri, *username, *password)
}

func runExample(uri string, username string, password string) {
	driver, err := neo4j.NewDriver(uri, username, password)
	panicOnError(err)
	defer func() {
		panicOnError(driver.Close())
	}()
	result, err := driver.Run("RETURN 42", neo4j.ReadAccessMode)
	panicOnError(err)
	fmt.Printf("%v\n", result)
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
