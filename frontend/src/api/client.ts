import { apiUrl } from "@/lib/api"
import { ApiError, type ApiErrorBody } from "@/types/api"

const SESSION_TOKEN_KEY = "graph-auth:session-token"

let authToken: string | null = null

export function getSessionToken(): string | null {
  if (authToken) return authToken
  return sessionStorage.getItem(SESSION_TOKEN_KEY)
}

export function setSessionToken(token: string | null): void {
  authToken = token
  if (token) {
    sessionStorage.setItem(SESSION_TOKEN_KEY, token)
  } else {
    sessionStorage.removeItem(SESSION_TOKEN_KEY)
  }
}

export function clearSessionToken(): void {
  authToken = null
  sessionStorage.removeItem(SESSION_TOKEN_KEY)
}

type RequestOptions = Omit<RequestInit, "body"> & {
  body?: unknown
  auth?: boolean
}

export async function apiClient<T>(
  path: string,
  options: RequestOptions = {}
): Promise<T> {
  const { body, auth = true, headers, ...init } = options

  const requestHeaders = new Headers(headers)

  if (body !== undefined) {
    requestHeaders.set("Content-Type", "application/json")
  }

  if (auth) {
    const token = getSessionToken()
    if (token) {
      requestHeaders.set("Authorization", `Bearer ${token}`)
    }
  }

  const url = apiUrl(path)
  const response = await fetch(url, {
    ...init,
    method: init.method ?? "GET",
    headers: requestHeaders,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (!response.ok) {
    let message = response.statusText
    try {
      const errorBody = (await response.json()) as ApiErrorBody
      message = errorBody.message ?? errorBody.error ?? message
    } catch {
      // ignore non-JSON error bodies
    }
    throw new ApiError(message, response.status)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}
