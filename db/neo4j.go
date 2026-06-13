package db

import (
	"context"
	"sync"

	"github.com/maslotwi/graph-auth/helpers/environment"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	neo4jOnce   sync.Once
	neo4jDriver neo4j.DriverWithContext
	neo4jErr    error
)

// Neo4j returns the shared Neo4j driver. The driver is safe for concurrent use;
// open a new session per operation or request.
func Neo4j() (neo4j.DriverWithContext, error) {
	neo4jOnce.Do(func() {
		neo4jDriver, neo4jErr = neo4j.NewDriverWithContext(
			environment.Neo4jURI,
			neo4j.BasicAuth(environment.Neo4jUser, environment.Neo4jPassword, ""),
		)
		if neo4jErr != nil {
			return
		}
		neo4jErr = neo4jDriver.VerifyConnectivity(context.Background())
	})
	return neo4jDriver, neo4jErr
}
