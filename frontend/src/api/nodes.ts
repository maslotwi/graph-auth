import { apiClient } from "./client"
import type {
  ConsumeCodeResponse,
  DelegationCodeResponse,
  MeNodeResponse,
  NodeTreeResponse,
} from "@/types/auth"

export function getCurrentNode(): Promise<MeNodeResponse> {
  return apiClient<MeNodeResponse>("/api/nodes/me")
}

export function getNodeTree(): Promise<NodeTreeResponse> {
  return apiClient<NodeTreeResponse>("/api/nodes/tree")
}

export function invalidateNode(id: string): Promise<{ message: string }> {
  return apiClient(`/api/nodes/${id}/invalidate`, { method: "POST" })
}

export function generateDelegationCode(): Promise<DelegationCodeResponse> {
  return apiClient<DelegationCodeResponse>("/api/auth/session/generate-code", {
    method: "POST",
  })
}

export function consumeDelegationCode(
  code: string,
  deviceName: string,
  scopes: string[]
): Promise<ConsumeCodeResponse> {
  return apiClient<ConsumeCodeResponse>("/api/auth/session/consume-code", {
    method: "POST",
    auth: false,
    body: { code, device_name: deviceName, scopes },
  })
}
