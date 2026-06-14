import { apiClient } from "./client"

export type ClientCredentials = {
  client_id: string
  client_secret: string
  name: string
}

export function createClient(name: string): Promise<ClientCredentials> {
  return apiClient<ClientCredentials>("/api/clients", {
    method: "POST",
    body: { name },
  })
}
