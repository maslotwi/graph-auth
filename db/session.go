package db

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// RootSession represents the canonical account anchor keyed by email.
type RootSession struct {
	Email     string
	CreatedAt time.Time
}

// Session represents a device-bound session node in the provenance graph.
type Session struct {
	Token      string
	DeviceName string
	Scopes     []string
	IsActive   bool
	CreatedAt  time.Time
}

// CreateSessionFromRoot ensures a RootSession exists for the email and creates a
// child Session node linked via SPAWNED.
func CreateSessionFromRoot(ctx context.Context, email string, session Session) error {
	driver, err := Neo4j()
	if err != nil {
		return err
	}

	neo4jSession := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer neo4jSession.Close(ctx)

	_, err = neo4jSession.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MERGE (root:RootSession {email: $email})
			ON CREATE SET root.created_at = datetime()
			CREATE (child:Session {
				token: $token,
				device_name: $deviceName,
				scopes: $scopes,
				is_active: $isActive,
				created_at: datetime()
			})
			CREATE (root)-[:SPAWNED {created_at: datetime()}]->(child)
			RETURN child.token AS token
		`, map[string]any{
			"email":      email,
			"token":      session.Token,
			"deviceName": session.DeviceName,
			"scopes":     session.Scopes,
			"isActive":   session.IsActive,
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to create session from root")
	})

	return err
}

// CreateChildSession creates a Session node spawned from an active parent Session.
func CreateChildSession(ctx context.Context, parentToken string, child Session) error {
	driver, err := Neo4j()
	if err != nil {
		return err
	}

	neo4jSession := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer neo4jSession.Close(ctx)

	_, err = neo4jSession.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (parent:Session {token: $parentToken})
			WHERE coalesce(parent.is_active, true) = true
			CREATE (child:Session {
				token: $token,
				device_name: $deviceName,
				scopes: $scopes,
				is_active: $isActive,
				created_at: datetime()
			})
			CREATE (parent)-[:SPAWNED {created_at: datetime()}]->(child)
			RETURN child.token AS token
		`, map[string]any{
			"parentToken": parentToken,
			"token":       child.Token,
			"deviceName":  child.DeviceName,
			"scopes":      child.Scopes,
			"isActive":    child.IsActive,
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return nil, nil
		}

		return nil, fmt.Errorf("parent session not found or inactive")
	})

	return err
}
