package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	// ErrParentNotFound is returned when the parent session is missing or inactive.
	ErrParentNotFound = errors.New("parent session not found or inactive")
	// ErrParentNotFertile is returned when the parent session lacks the fertile scope.
	ErrParentNotFertile = errors.New("parent session is not fertile")
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

// SessionHasScope reports whether an active session node includes the given scope.
func SessionHasScope(ctx context.Context, token, scope string) (bool, error) {
	driver, err := Neo4j()
	if err != nil {
		return false, err
	}

	neo4jSession := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer neo4jSession.Close(ctx)

	result, err := neo4jSession.Run(ctx, `
		MATCH (s:Session {token: $token})
		WHERE coalesce(s.is_active, true) = true
		RETURN $scope IN coalesce(s.scopes, []) AS hasScope
	`, map[string]any{
		"token": token,
		"scope": scope,
	})
	if err != nil {
		return false, err
	}

	if result.Next(ctx) {
		val, _ := result.Record().Get("hasScope")
		if b, ok := val.(bool); ok {
			return b, nil
		}
	}

	return false, nil
}

// ActiveSessionEmail returns the root account email for an active session token.
// The second return value is true when the session exists and is active.
func ActiveSessionEmail(ctx context.Context, token string) (string, bool, error) {
	driver, err := Neo4j()
	if err != nil {
		return "", false, err
	}

	neo4jSession := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer neo4jSession.Close(ctx)

	result, err := neo4jSession.Run(ctx, `
		MATCH (root:RootSession)-[:SPAWNED*1..]->(s:Session {token: $token})
		WHERE coalesce(s.is_active, true) = true
		RETURN root.email AS email
		LIMIT 1
	`, map[string]any{"token": token})
	if err != nil {
		return "", false, err
	}

	if result.Next(ctx) {
		email, _ := result.Record().Get("email")
		if emailStr, ok := email.(string); ok && emailStr != "" {
			return emailStr, true, nil
		}
	}

	return "", false, nil
}

// CreateChildSession creates a Session node spawned from an active, fertile parent Session.
func CreateChildSession(ctx context.Context, parentToken string, child Session) error {
	driver, err := Neo4j()
	if err != nil {
		return err
	}

	neo4jSession := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer neo4jSession.Close(ctx)

	_, err = neo4jSession.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		check, err := tx.Run(ctx, `
			MATCH (parent:Session {token: $parentToken})
			RETURN parent IS NOT NULL AS found,
			       coalesce(parent.is_active, true) AS active,
			       "fertile" IN coalesce(parent.scopes, []) AS fertile
		`, map[string]any{"parentToken": parentToken})
		if err != nil {
			return nil, err
		}

		if !check.Next(ctx) {
			return nil, ErrParentNotFound
		}

		record := check.Record()
		found, _ := record.Get("found")
		active, _ := record.Get("active")
		fertile, _ := record.Get("fertile")

		if foundBool, ok := found.(bool); !ok || !foundBool {
			return nil, ErrParentNotFound
		}
		if activeBool, ok := active.(bool); !ok || !activeBool {
			return nil, ErrParentNotFound
		}
		if fertileBool, ok := fertile.(bool); !ok || !fertileBool {
			return nil, ErrParentNotFertile
		}

		result, err := tx.Run(ctx, `
			MATCH (parent:Session {token: $parentToken})
			WHERE coalesce(parent.is_active, true) = true
			  AND "fertile" IN coalesce(parent.scopes, [])
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

		return nil, ErrParentNotFound
	})

	return err
}
