package db

import (
	"context"
	"crypto/subtle"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Client represents an OAuth client owned by a RootSession.
type Client struct {
	ClientID     string
	ClientSecret string
	Name         string
	RedirectURIs []string
	CreatedAt    time.Time
}

// ClientHasRedirectURI reports whether uri is registered for the client.
func ClientHasRedirectURI(client Client, uri string) bool {
	for _, registered := range client.RedirectURIs {
		if registered == uri {
			return true
		}
	}
	return false
}

// CreateClientForRoot creates a Client node owned by the RootSession for the given email.
func CreateClientForRoot(ctx context.Context, email string, client Client) error {
	driver, err := Neo4j()
	if err != nil {
		return err
	}

	neo4jSession := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer neo4jSession.Close(ctx)

	_, err = neo4jSession.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (root:RootSession {email: $email})
			CREATE (c:Client {
				client_id: $clientID,
				client_secret: $clientSecret,
				name: $name,
				redirect_uris: $redirectURIs,
				created_at: datetime()
			})
			CREATE (root)-[:OWNS {created_at: datetime()}]->(c)
			RETURN c.client_id AS client_id
		`, map[string]any{
			"email":        email,
			"clientID":     client.ClientID,
			"clientSecret": client.ClientSecret,
			"name":         client.Name,
			"redirectURIs": client.RedirectURIs,
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return nil, nil
		}

		return nil, ErrRootNotFound
	})
	if err != nil {
		return err
	}

	return nil
}

// GetClientsForRoot returns all Client nodes owned by the RootSession for the given email.
func GetClientsForRoot(ctx context.Context, email string) ([]Client, error) {
	driver, err := Neo4j()
	if err != nil {
		return nil, err
	}

	neo4jSession := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer neo4jSession.Close(ctx)

	result, err := neo4jSession.Run(ctx, `
		MATCH (root:RootSession {email: $email})-[:OWNS]->(c:Client)
		RETURN c.client_id AS client_id,
		       c.name AS name,
		       coalesce(c.redirect_uris, []) AS redirect_uris
		ORDER BY c.created_at
	`, map[string]any{"email": email})
	if err != nil {
		return nil, err
	}

	var clients []Client
	for result.Next(ctx) {
		record := result.Record()
		clientIDVal, _ := record.Get("client_id")
		nameVal, _ := record.Get("name")
		redirectURIsVal, _ := record.Get("redirect_uris")

		clientIDStr, ok1 := clientIDVal.(string)
		nameStr, ok2 := nameVal.(string)
		if !ok1 || !ok2 {
			continue
		}

		clients = append(clients, Client{
			ClientID:     clientIDStr,
			Name:         nameStr,
			RedirectURIs: parseScopeSlice(redirectURIsVal),
		})
	}

	return clients, result.Err()
}

// GetClientByID returns the Client node for the given client_id.
func GetClientByID(ctx context.Context, clientID string) (Client, error) {
	driver, err := Neo4j()
	if err != nil {
		return Client{}, err
	}

	neo4jSession := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer neo4jSession.Close(ctx)

	result, err := neo4jSession.Run(ctx, `
		MATCH (c:Client {client_id: $clientID})
		RETURN c.client_id AS client_id,
		       c.client_secret AS client_secret,
		       c.name AS name,
		       coalesce(c.redirect_uris, []) AS redirect_uris
	`, map[string]any{"clientID": clientID})
	if err != nil {
		return Client{}, err
	}

	if result.Next(ctx) {
		record := result.Record()
		clientIDVal, _ := record.Get("client_id")
		clientSecretVal, _ := record.Get("client_secret")
		nameVal, _ := record.Get("name")
		redirectURIsVal, _ := record.Get("redirect_uris")

		clientIDStr, ok1 := clientIDVal.(string)
		clientSecretStr, ok2 := clientSecretVal.(string)
		nameStr, ok3 := nameVal.(string)
		if !ok1 || !ok2 || !ok3 {
			return Client{}, ErrClientNotFound
		}

		return Client{
			ClientID:     clientIDStr,
			ClientSecret: clientSecretStr,
			Name:         nameStr,
			RedirectURIs: parseScopeSlice(redirectURIsVal),
		}, nil
	}

	return Client{}, ErrClientNotFound
}

// VerifyClientCredentials reports whether the client_id and client_secret match.
func VerifyClientCredentials(ctx context.Context, clientID, clientSecret string) (bool, error) {
	client, err := GetClientByID(ctx, clientID)
	if err != nil {
		return false, err
	}

	return subtle.ConstantTimeCompare([]byte(client.ClientSecret), []byte(clientSecret)) == 1, nil
}
