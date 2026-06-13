import type { GraphNode } from "@/types/node"
import { PERMISSIONS } from "@/types/node"

type Session = {
  email: string
  node: GraphNode | null
}

const sessions = new Map<string, Session>()

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
  return node
}

export function resetMockStore(): void {
  sessions.clear()
}
