import {
  createContext,
  useCallback,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react"

import * as authApi from "@/api/auth"
import {
  clearSessionToken,
  getSessionToken,
  setSessionToken,
} from "@/api/client"
import * as nodesApi from "@/api/nodes"
import { ApiError } from "@/types/api"
import type { CreateRootNodeRequest, VerifyResponse } from "@/types/auth"
import type { GraphNode } from "@/types/node"

const EMAIL_KEY = "graph-auth:email"
const REQUIRES_ROOT_SETUP_KEY = "graph-auth:requires-root-setup"

type AuthContextValue = {
  sessionToken: string | null
  email: string | null
  currentNode: GraphNode | null
  requiresRootSetup: boolean
  isAuthenticated: boolean
  isLoading: boolean
  register: (email: string) => Promise<void>
  verify: (token: string) => Promise<VerifyResponse>
  loginWithToken: (token: string) => Promise<void>
  createRoot: (data: CreateRootNodeRequest) => Promise<void>
  logout: () => void
}

export const AuthContext = createContext<AuthContextValue | null>(null)

function readRequiresRootSetup(): boolean {
  return sessionStorage.getItem(REQUIRES_ROOT_SETUP_KEY) === "true"
}

function persistSession(
  token: string,
  email: string,
  requiresRootSetup: boolean
): void {
  setSessionToken(token)
  sessionStorage.setItem(EMAIL_KEY, email)
  sessionStorage.setItem(
    REQUIRES_ROOT_SETUP_KEY,
    requiresRootSetup ? "true" : "false"
  )
}

function clearPersistedSession(): void {
  clearSessionToken()
  sessionStorage.removeItem(EMAIL_KEY)
  sessionStorage.removeItem(REQUIRES_ROOT_SETUP_KEY)
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [sessionToken, setSessionTokenState] = useState<string | null>(() =>
    getSessionToken()
  )
  const [email, setEmail] = useState<string | null>(() =>
    sessionStorage.getItem(EMAIL_KEY)
  )
  const [currentNode, setCurrentNode] = useState<GraphNode | null>(null)
  const [requiresRootSetup, setRequiresRootSetup] = useState(() =>
    readRequiresRootSetup()
  )
  const [isLoading, setIsLoading] = useState(true)

  const bootstrap = useCallback(async () => {
    const token = getSessionToken()
    if (!token) {
      setIsLoading(false)
      return
    }

    setSessionTokenState(token)
    setEmail(sessionStorage.getItem(EMAIL_KEY))
    setRequiresRootSetup(readRequiresRootSetup())

    try {
      const { node } = await nodesApi.getCurrentNode()
      setCurrentNode(node)
      setRequiresRootSetup(false)
      sessionStorage.setItem(REQUIRES_ROOT_SETUP_KEY, "false")
    } catch (error) {
      if (error instanceof ApiError && error.status === 404) {
        setCurrentNode(null)
        setRequiresRootSetup(true)
        sessionStorage.setItem(REQUIRES_ROOT_SETUP_KEY, "true")
      } else if (error instanceof ApiError && error.status === 401) {
        clearPersistedSession()
        setSessionTokenState(null)
        setEmail(null)
        setCurrentNode(null)
        setRequiresRootSetup(false)
      }
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    void bootstrap()
  }, [bootstrap])

  const applyVerifyResponse = useCallback((response: VerifyResponse) => {
    persistSession(
      response.sessionToken,
      response.email,
      response.requiresRootSetup
    )
    setSessionTokenState(response.sessionToken)
    setEmail(response.email)
    setRequiresRootSetup(response.requiresRootSetup)
    setCurrentNode(null)
  }, [])

  const register = useCallback(async (registerEmail: string) => {
    await authApi.register({ email: registerEmail })
  }, [])

  const verify = useCallback(
    async (token: string) => {
      const response = await authApi.verifyEmail({ token })
      applyVerifyResponse(response)
      return response
    },
    [applyVerifyResponse]
  )

  const loginWithToken = useCallback(async (token: string) => {
    setSessionToken(token)
    setSessionTokenState(token)
    setEmail(null)
    setRequiresRootSetup(false)
    sessionStorage.removeItem(EMAIL_KEY)
    sessionStorage.setItem(REQUIRES_ROOT_SETUP_KEY, "false")
    try {
      const { node } = await nodesApi.getCurrentNode()
      setCurrentNode(node)
    } catch {
      setCurrentNode(null)
    }
  }, [])

  const createRoot = useCallback(async (data: CreateRootNodeRequest) => {
    const { node } = await nodesApi.createRootNode(data)
    setCurrentNode(node)
    setRequiresRootSetup(false)
    sessionStorage.setItem(REQUIRES_ROOT_SETUP_KEY, "false")
  }, [])

  const logout = useCallback(() => {
    clearPersistedSession()
    setSessionTokenState(null)
    setEmail(null)
    setCurrentNode(null)
    setRequiresRootSetup(false)
  }, [])

  const value = useMemo<AuthContextValue>(
    () => ({
      sessionToken,
      email,
      currentNode,
      requiresRootSetup,
      isAuthenticated: Boolean(sessionToken),
      isLoading,
      register,
      verify,
      loginWithToken,
      createRoot,
      logout,
    }),
    [
      sessionToken,
      email,
      currentNode,
      requiresRootSetup,
      isLoading,
      register,
      verify,
      loginWithToken,
      createRoot,
      logout,
    ]
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
