import type { GraphNode, Permission } from "./node"

export type RegisterRequest = {
  email: string
}

export type RegisterResponse = {
  message: string
}

export type VerifyRequest = {
  token: string
}

export type VerifyResponse = {
  sessionToken: string
  email: string
  requiresRootSetup: boolean
}

export type CreateRootNodeRequest = {
  label: string
  permissions: Permission[]
}

export type CreateRootNodeResponse = {
  node: GraphNode
}

export type MeNodeResponse = {
  node: GraphNode
}

export type NodeTreeResponse = {
  nodes: GraphNode[]
}
