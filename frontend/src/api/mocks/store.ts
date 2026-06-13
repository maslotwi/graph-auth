import type { GraphNode } from "@/types/node"
import { PERMISSIONS } from "@/types/node"

type Session = {
  email: string
  node: GraphNode | null
}

const sessions = new Map<string, Session>()
const nodeChildren = new Map<string, GraphNode[]>()
const invites = new Map<string, string>() // token → parentNodeId

export function createSession(email: string): string {
  const token = `mock-token-${crypto.randomUUID()}`
  sessions.set(token, { email, node: null })
  return token
}

export function getSession(token: string): Session | undefined {
  return sessions.get(token)
}

export function createRootNode(
  token: string,
  label: string,
  permissions: GraphNode["permissions"]
): GraphNode | null {
  const session = sessions.get(token)
  if (!session || session.node) return null

  const node: GraphNode = {
    id: `node-${crypto.randomUUID()}`,
    label,
    isRoot: true,
    permissions: permissions.length > 0 ? permissions : [...PERMISSIONS],
    status: "active",
  }

  session.node = node
  seedDemoChildren(node)
  return node
}

export function createChildNode(
  parentId: string,
  label: string,
  permissions: GraphNode["permissions"]
): GraphNode {
  const node: GraphNode = {
    id: `node-${crypto.randomUUID()}`,
    label,
    isRoot: false,
    permissions,
    status: "active",
    predecessorId: parentId,
  }

  const siblings = nodeChildren.get(parentId) ?? []
  nodeChildren.set(parentId, [...siblings, node])
  return node
}

export function getNodeTree(token: string): GraphNode[] | null {
  const session = sessions.get(token)
  if (!session?.node) return null

  const all: GraphNode[] = []

  function collect(nodeId: string) {
    const children = nodeChildren.get(nodeId) ?? []
    for (const child of children) {
      all.push(child)
      collect(child.id)
    }
  }

  all.push(session.node)
  collect(session.node.id)
  return all
}

export function invalidateNode(id: string): boolean {
  for (const session of sessions.values()) {
    if (session.node?.id === id) {
      session.node.status = "invalidated"
      return true
    }
  }
  for (const children of nodeChildren.values()) {
    const node = children.find((n) => n.id === id)
    if (node) {
      node.status = "invalidated"
      return true
    }
  }
  return false
}

function seedDemoChildren(root: GraphNode) {
  const laptop = createChildNode(root.id, "Laptop", ["read", "write", "admin"])
  createChildNode(root.id, "Phone", ["read"])
  const work = createChildNode(laptop.id, "Work Profile", ["read", "write", "sso"])
  createChildNode(work.id, "CI Runner", ["read", "write"])
}

export function createInvite(parentNodeId: string): string {
  const token = `invite-${crypto.randomUUID()}`
  invites.set(token, parentNodeId)
  return token
}

export function redeemInvite(token: string): string | null {
  const parentId = invites.get(token) ?? null
  if (parentId) invites.delete(token)
  return parentId
}

export function resetMockStore(): void {
  sessions.clear()
  nodeChildren.clear()
  invites.clear()
}
