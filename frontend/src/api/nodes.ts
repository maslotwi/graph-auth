import { apiClient } from "./client"
import type {
  CreateRootNodeRequest,
  CreateRootNodeResponse,
  MeNodeResponse,
  NodeTreeResponse,
} from "@/types/auth"

export function createRootNode(
  data: CreateRootNodeRequest
): Promise<CreateRootNodeResponse> {
  return apiClient<CreateRootNodeResponse>("/api/nodes/root", {
    method: "POST",
    body: data,
  })
}

export function getCurrentNode(): Promise<MeNodeResponse> {
  return apiClient<MeNodeResponse>("/api/nodes/me")
}

export function getNodeTree(): Promise<NodeTreeResponse> {
  return apiClient<NodeTreeResponse>("/api/nodes/tree")
}

export function invalidateNode(id: string): Promise<{ message: string }> {
  return apiClient(`/api/nodes/${id}/invalidate`, { method: "POST" })
}

export function createInvite(id: string): Promise<{ token: string }> {
  return apiClient(`/api/nodes/${id}/invite`, { method: "POST" })
}
