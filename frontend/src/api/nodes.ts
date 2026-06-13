import { apiClient } from "./client"
import type {
  CreateRootNodeRequest,
  CreateRootNodeResponse,
  MeNodeResponse,
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
