import type { GraphNode } from "./node"

export type RegisterRequest = {
  email: string
}

export type RegisterResponse = {
  message: string
}

export type VerifyRequest = {
  token: string
  name?: string
  scopes?: string[]
}

export type VerifyResponse = {
  sessionToken: string
  email: string
}

export type MeNodeResponse = {
  node: GraphNode
}

export type NodeTreeResponse = {
  nodes: GraphNode[]
}

export type DelegationCodeResponse = {
  code: string
  link: string
  expires_in: number
}

export type ConsumeCodeResponse = {
  session_token: string
  scopes: string[]
  status: string
}
