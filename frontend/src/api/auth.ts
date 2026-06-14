import { apiClient } from "./client"
import type {
  RegisterRequest,
  RegisterResponse,
  VerifyRequest,
  VerifyResponse,
} from "@/types/auth"

export function register(data: RegisterRequest): Promise<RegisterResponse> {
  return apiClient<RegisterResponse>("/api/auth/register", {
    method: "POST",
    body: data,
    auth: false,
  })
}

export function verifyEmail(data: VerifyRequest): Promise<VerifyResponse> {
  return apiClient<VerifyResponse>("/api/auth/verify", {
    method: "POST",
    body: data,
    auth: false,
  })
}

export function invalidateSession(): Promise<{ message: string }> {
  return apiClient("/api/auth/session/invalidate", { method: "POST" })
}

export function confirmSSO(data: {
  client_id: string
  redirect_uri: string
  state: string
  scope?: string
}): Promise<{ status: string; redirect_to: string }> {
  return apiClient("/api/oauth/authorize", { method: "POST", body: data })
}
