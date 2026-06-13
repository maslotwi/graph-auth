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

export function confirmSSO(data: {
  client_id: string
  redirect_uri: string
  state: string
}): Promise<{ status: string; redirect_to: string }> {
  return apiClient("/api/oauth/confirm", { method: "POST", body: data })
}
