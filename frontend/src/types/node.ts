export const PERMISSIONS = ["read", "write", "admin", "sso"] as const

export type Permission = (typeof PERMISSIONS)[number]

export type NodeStatus = "active" | "invalidated"

export type GraphNode = {
  id: string
  label: string
  isRoot: boolean
  permissions: Permission[]
  status: NodeStatus
  predecessorId?: string
}
