import { apiClient } from "./client"

export type ClientCredentials = {
  client_id: string
  client_secret: string
  name: string
  redirect_uris: string[]
}

export function createClient(
  name: string,
  redirectUris: string[]
): Promise<ClientCredentials> {
  return apiClient<ClientCredentials>("/api/clients", {
    method: "POST",
    body: { name, redirect_uris: redirectUris },
  })
}

export function getClientInfo(clientId: string): Promise<{ name: string }> {
  return apiClient<{ name: string }>(`/api/clients/${clientId}`, { auth: false })
}
