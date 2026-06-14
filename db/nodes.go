package db

import (
	"context"
	"errors"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var ErrNodeNotFound = errors.New("session node not found")
var ErrNotAuthorized = errors.New("target node not in caller's tree")

// NodeRecord is a session node as returned to the API layer.
type NodeRecord struct {
	ID            string
	Label         string
	IsRoot        bool
	Permissions   []string
	IsActive      bool
	PredecessorID string // empty when parent is a RootSession
}

// GetSessionNode returns the single session node for the given token.
func GetSessionNode(ctx context.Context, token string) (*NodeRecord, error) {
	driver, err := Neo4j()
	if err != nil {
		return nil, err
	}

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx, `
		MATCH (parent)-[:SPAWNED]->(s:Session {token: $token})
		RETURN s.token            AS id,
		       s.device_name      AS label,
		       coalesce(s.scopes, [])        AS permissions,
		       coalesce(s.is_active, true)   AS isActive,
		       parent:RootSession            AS isRoot,
		       CASE WHEN parent:Session THEN parent.token ELSE null END AS predecessorId
		LIMIT 1
	`, map[string]any{"token": token})
	if err != nil {
		return nil, err
	}

	if !result.Next(ctx) {
		return nil, ErrNodeNotFound
	}

	return recordToNode(result.Record()), nil
}

// GetSessionTree returns all session nodes in the account tree that the given token belongs to.
func GetSessionTree(ctx context.Context, token string) ([]NodeRecord, error) {
	driver, err := Neo4j()
	if err != nil {
		return nil, err
	}

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx, `
		MATCH (root:RootSession)-[:SPAWNED*1..]->(me:Session {token: $token})
		WITH root
		MATCH (root)-[:SPAWNED*1..]->(n:Session)
		OPTIONAL MATCH (p)-[:SPAWNED]->(n)
		RETURN n.token            AS id,
		       n.device_name      AS label,
		       coalesce(n.scopes, [])        AS permissions,
		       coalesce(n.is_active, true)   AS isActive,
		       p:RootSession                 AS isRoot,
		       CASE WHEN p:Session THEN p.token ELSE null END AS predecessorId
		ORDER BY n.created_at
	`, map[string]any{"token": token})
	if err != nil {
		return nil, err
	}

	var nodes []NodeRecord
	for result.Next(ctx) {
		nodes = append(nodes, *recordToNode(result.Record()))
	}
	return nodes, result.Err()
}

// InvalidateNodeInTree invalidates the target session and all its descendants,
// but only if the target belongs to the same account tree as the caller.
func InvalidateNodeInTree(ctx context.Context, callerToken, targetToken string) error {
	driver, err := Neo4j()
	if err != nil {
		return err
	}

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// Verify target is a descendant of caller (not just in the same tree).
	result, err := session.Run(ctx, `
		MATCH (caller:Session {token: $callerToken})-[:SPAWNED*1..]->(target:Session {token: $targetToken})
		RETURN count(target) > 0 AS authorized
	`, map[string]any{
		"callerToken": callerToken,
		"targetToken": targetToken,
	})
	if err != nil {
		return err
	}
	if !result.Next(ctx) {
		return ErrNotAuthorized
	}
	authorized, _ := result.Record().Get("authorized")
	if ok, _ := authorized.(bool); !ok {
		return ErrNotAuthorized
	}
	_ = session.Close(ctx)

	return InvalidateSessionTree(ctx, targetToken)
}

func recordToNode(r *neo4j.Record) *NodeRecord {
	str := func(key string) string {
		v, _ := r.Get(key)
		s, _ := v.(string)
		return s
	}
	boolVal := func(key string) bool {
		v, _ := r.Get(key)
		b, _ := v.(bool)
		return b
	}

	return &NodeRecord{
		ID:            str("id"),
		Label:         str("label"),
		IsRoot:        boolVal("isRoot"),
		Permissions:   parseScopeSlice(func() any { v, _ := r.Get("permissions"); return v }()),
		IsActive:      boolVal("isActive"),
		PredecessorID: str("predecessorId"),
	}
}
