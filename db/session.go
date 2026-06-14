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
	// ErrChildScopesNotSubset is returned when child scopes exceed parent scopes.
	ErrChildScopesNotSubset = errors.New("child scopes are not a subset of parent scopes")
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

// ActiveSessionScopes returns the scopes of an active session token.
// The second return value is true when the session exists and is active.
func ActiveSessionScopes(ctx context.Context, token string) ([]string, bool, error) {
	driver, err := Neo4j()
	if err != nil {
		return nil, false, err
	}

	neo4jSession := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer neo4jSession.Close(ctx)

	result, err := neo4jSession.Run(ctx, `
		MATCH (s:Session {token: $token})
		WHERE coalesce(s.is_active, true) = true
		RETURN coalesce(s.scopes, []) AS scopes
	`, map[string]any{"token": token})
	if err != nil {
		return nil, false, err
	}

	if result.Next(ctx) {
		scopes, _ := result.Record().Get("scopes")
		return parseScopeSlice(scopes), true, nil
	}

	return nil, false, nil
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
			       "fertile" IN coalesce(parent.scopes, []) AS fertile,
			       coalesce(parent.scopes, []) AS parentScopes
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
		parentScopesRaw, _ := record.Get("parentScopes")

		if foundBool, ok := found.(bool); !ok || !foundBool {
			return nil, ErrParentNotFound
		}
		if activeBool, ok := active.(bool); !ok || !activeBool {
			return nil, ErrParentNotFound
		}
		if fertileBool, ok := fertile.(bool); !ok || !fertileBool {
			return nil, ErrParentNotFertile
		}
		if !scopesSubset(child.Scopes, parseScopeSlice(parentScopesRaw)) {
			return nil, ErrChildScopesNotSubset
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

// InvalidateSessionTree marks a session and every session it has spawned (at any depth)
// as inactive. Used when a device logs out — no descendant can remain active.
func InvalidateSessionTree(ctx context.Context, token string) error {
	driver, err := Neo4j()
	if err != nil {
		return err
	}

	neo4jSession := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer neo4jSession.Close(ctx)

	_, err = neo4jSession.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `
			MATCH (s:Session {token: $token})-[:SPAWNED*0..]->(d:Session)
			SET d.is_active = false
		`, map[string]any{"token": token})
		return nil, err
	})
	return err
}

func parseScopeSlice(val any) []string {
	switch v := val.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func scopesSubset(child, parent []string) bool {
	parentSet := make(map[string]struct{}, len(parent))
	for _, scope := range parent {
		parentSet[scope] = struct{}{}
	}
	for _, scope := range child {
		if _, ok := parentSet[scope]; !ok {
			return false
		}
	}
	return true
}
