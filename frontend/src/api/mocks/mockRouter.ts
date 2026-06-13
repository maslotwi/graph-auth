import { apiUrl } from "@/lib/api"
import type { GraphNode } from "@/types/node"
import {
  consumeDelegationCode,
  createChildNode,
  createDelegationCode,
  createRootNode,
  createSession,
  getNodeTree,
  getSession,
  invalidateNode,
} from "./store"

type MockRequest = {
  path: string
  method: string
  headers: Headers
  body?: unknown
}

function jsonResponse(data: unknown, status = 200): Response {
  return Response.json(data, { status })
}

function getBearerToken(headers: Headers): string | null {
  const authHeader = headers.get("Authorization")
  return authHeader?.replace("Bearer ", "") ?? null
}

function normalizePath(path: string): string {
  const base = apiUrl("")
  if (base && base !== "/" && path.startsWith(base)) {
    return path.slice(base.length) || "/"
  }
  return path
}

export async function handleMockRequest(
  request: MockRequest
): Promise<Response | null> {
  const path = normalizePath(request.path)
  const method = request.method.toUpperCase()

  if (method === "POST" && path === "/api/auth/register") {
    const body = request.body as { email?: string }
    console.info("[mock] register:", body.email)
    return jsonResponse({
      message: "Check your email for a verification link.",
    })
  }

  if (method === "POST" && path === "/api/auth/verify") {
    const body = request.body as { token?: string }
    if (!body.token) {
      return jsonResponse({ message: "Verification token is required." }, 400)
    }
    const email = "user@example.com"
    const sessionToken = createSession(email)
    return jsonResponse({ sessionToken, email, requiresRootSetup: true })
  }

  if (method === "POST" && path === "/api/auth/session/generate-code") {
    const token = getBearerToken(request.headers)
    if (!token) return jsonResponse({ message: "Unauthorized" }, 401)
    const session = getSession(token)
    if (!session?.node) return jsonResponse({ message: "Unauthorized" }, 401)
    const body = request.body as {
      scopes?: GraphNode["permissions"]
      node_id?: string
    }
    const parentNodeId = body.node_id ?? session.node.id
    const scopes = body.scopes ?? session.node.permissions
    return jsonResponse(createDelegationCode(parentNodeId, scopes))
  }

  if (method === "POST" && path === "/api/auth/session/consume-code") {
    const body = request.body as { code?: string; device_name?: string }
    if (!body.code) return jsonResponse({ message: "Code is required." }, 400)
    const result = consumeDelegationCode(
      body.code,
      body.device_name ?? "New Device"
    )
    if (!result)
      return jsonResponse({ message: "Code expired or invalid." }, 401)
    return jsonResponse(result)
  }

  if (method === "POST" && path === "/api/nodes/root") {
    const token = getBearerToken(request.headers)
    if (!token) {
      return jsonResponse({ message: "Unauthorized" }, 401)
    }

    const session = getSession(token)
    if (!session) {
      return jsonResponse({ message: "Unauthorized" }, 401)
    }

    const body = request.body as {
      label?: string
      permissions?: GraphNode["permissions"]
    }

    const node = createRootNode(
      token,
      body.label?.trim() || "Root Node",
      body.permissions ?? []
    )

    if (!node) {
      return jsonResponse({ message: "Root node already exists." }, 409)
    }

    return jsonResponse({ node })
  }

  if (method === "GET" && path === "/api/nodes/me") {
    const token = getBearerToken(request.headers)
    if (!token) {
      return jsonResponse({ message: "Unauthorized" }, 401)
    }

    const session = getSession(token)
    if (!session) {
      return jsonResponse({ message: "Unauthorized" }, 401)
    }

    if (!session.node) {
      return jsonResponse({ message: "Root node not set up yet." }, 404)
    }

    return jsonResponse({ node: session.node })
  }

  if (method === "GET" && path === "/api/nodes/tree") {
    const token = getBearerToken(request.headers)
    if (!token) {
      return jsonResponse({ message: "Unauthorized" }, 401)
    }

    const nodes = getNodeTree(token)
    if (!nodes) {
      return jsonResponse({ message: "Root node not set up yet." }, 404)
    }

    return jsonResponse({ nodes })
  }

  if (method === "POST" && path === "/api/nodes/child") {
    const token = getBearerToken(request.headers)
    if (!token) {
      return jsonResponse({ message: "Unauthorized" }, 401)
    }

    const session = getSession(token)
    if (!session?.node) {
      return jsonResponse({ message: "Root node not set up yet." }, 404)
    }

    const body = request.body as {
      parentId?: string
      label?: string
      permissions?: GraphNode["permissions"]
    }

    if (!body.parentId || !body.label) {
      return jsonResponse({ message: "parentId and label are required." }, 400)
    }

    const node = createChildNode(
      body.parentId,
      body.label.trim(),
      body.permissions ?? []
    )

    return jsonResponse({ node })
  }

  const invalidateMatch = path.match(/^\/api\/nodes\/([^/]+)\/invalidate$/)
  if (method === "POST" && invalidateMatch) {
    const token = getBearerToken(request.headers)
    if (!token) return jsonResponse({ message: "Unauthorized" }, 401)
    if (!getSession(token))
      return jsonResponse({ message: "Unauthorized" }, 401)

    const nodeId = invalidateMatch[1]
    const ok = invalidateNode(nodeId)
    if (!ok) return jsonResponse({ message: "Node not found." }, 404)
    return jsonResponse({ message: "Node invalidated." })
  }

  if (method === "POST" && path === "/api/oauth/confirm") {
    const token = getBearerToken(request.headers)
    if (!token) return jsonResponse({ message: "Unauthorized" }, 401)
    if (!getSession(token))
      return jsonResponse({ message: "Unauthorized" }, 401)
    const body = request.body as {
      client_id?: string
      redirect_uri?: string
      state?: string
    }
    if (!body.redirect_uri)
      return jsonResponse({ message: "redirect_uri is required." }, 400)
    const code = `mock-code-${crypto.randomUUID().slice(0, 8)}`
    const redirectTo = `${body.redirect_uri}?code=${code}&state=${body.state ?? ""}`
    return jsonResponse({ status: "success", redirect_to: redirectTo })
  }

  if (method === "GET" && path === "/api/health") {
    return jsonResponse({ status: "ok" })
  }

  return null
}
