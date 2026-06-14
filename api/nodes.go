package api

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/maslotwi/graph-auth/db"
)

// RegisterNodeRoutes registers the session node endpoints.
func RegisterNodeRoutes(app *fiber.App) {
	nodes := app.Group("/api/nodes", RequireSession)

	nodes.Get("/me", HandleGetMe)
	nodes.Get("/tree", HandleGetTree)
	nodes.Post("/:id/invalidate", HandleInvalidateNode)
}

// HandleGetMe returns the session node for the authenticated token.
//
// @Summary             Get Current Node
// @Description         Returns the session node associated with the authenticated Bearer token.
// @Tags                Nodes
// @Produce             json
// @Param               Authorization header string true "Bearer <token>"
// @Success             200 {object} NodeResponse "Current session node"
// @Failure             401 {object} ErrorResponse "Missing or invalid session"
// @Failure             404 {object} ErrorResponse "Node not found"
// @Router              /api/nodes/me [get]
func HandleGetMe(c fiber.Ctx) error {
	token := c.Locals("sessionToken").(string)

	node, err := db.GetSessionNode(context.Background(), token)
	if err != nil {
		if errors.Is(err, db.ErrNodeNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "node_not_found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "lookup_failed"})
	}

	return c.JSON(NodeResponse{Node: toAPINode(node)})
}

// HandleGetTree returns all session nodes in the account tree.
//
// @Summary             Get Session Tree
// @Description         Returns every session node in the account tree that the authenticated token belongs to.
// @Tags                Nodes
// @Produce             json
// @Param               Authorization header string true "Bearer <token>"
// @Success             200 {object} NodeTreeResponse "All nodes in the account tree"
// @Failure             401 {object} ErrorResponse "Missing or invalid session"
// @Failure             500 {object} ErrorResponse "Lookup failed"
// @Router              /api/nodes/tree [get]
func HandleGetTree(c fiber.Ctx) error {
	token := c.Locals("sessionToken").(string)

	nodes, err := db.GetSessionTree(context.Background(), token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "lookup_failed"})
	}

	apiNodes := make([]APINode, len(nodes))
	for i, n := range nodes {
		apiNodes[i] = toAPINode(&n)
	}

	return c.JSON(fiber.Map{"nodes": apiNodes})
}

// HandleInvalidateNode invalidates a target session node and all its descendants.
// The target must belong to the same account tree as the caller.
//
// @Summary             Invalidate Node
// @Description         Permanently deactivates the target session and every session it has spawned. The target must be in the caller's account tree.
// @Tags                Nodes
// @Produce             json
// @Param               Authorization header string true "Bearer <token>"
// @Param               id path string true "Session token of the node to invalidate"
// @Success             200 {object} map[string]string "Node invalidated"
// @Failure             401 {object} ErrorResponse "Missing or invalid session"
// @Failure             403 {object} ErrorResponse "Target not in caller's tree"
// @Failure             500 {object} ErrorResponse "Invalidation failed"
// @Router              /api/nodes/{id}/invalidate [post]
func HandleInvalidateNode(c fiber.Ctx) error {
	callerToken := c.Locals("sessionToken").(string)
	targetToken := c.Params("id")

	err := db.InvalidateNodeInTree(context.Background(), callerToken, targetToken)
	if err != nil {
		if errors.Is(err, db.ErrNotAuthorized) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "not_authorized"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "invalidate_failed"})
	}

	return c.JSON(fiber.Map{"message": "node_invalidated"})
}

// APINode is the JSON representation of a session node.
type APINode struct {
	ID            string   `json:"id"`
	Label         string   `json:"label"`
	IsRoot        bool     `json:"isRoot"`
	Permissions   []string `json:"permissions"`
	Status        string   `json:"status"`
	PredecessorID *string  `json:"predecessorId,omitempty"`
}

// NodeResponse wraps a single node for /me.
type NodeResponse struct {
	Node APINode `json:"node"`
}

// NodeTreeResponse wraps the list of nodes for /tree.
type NodeTreeResponse struct {
	Nodes []APINode `json:"nodes"`
}

func toAPINode(n *db.NodeRecord) APINode {
	node := APINode{
		ID:          n.ID,
		Label:       n.Label,
		IsRoot:      n.IsRoot,
		Permissions: n.Permissions,
		Status:      "active",
	}
	if !n.IsActive {
		node.Status = "invalidated"
	}
	if n.Permissions == nil {
		node.Permissions = []string{}
	}
	if n.PredecessorID != "" {
		node.PredecessorID = &n.PredecessorID
	}
	return node
}
