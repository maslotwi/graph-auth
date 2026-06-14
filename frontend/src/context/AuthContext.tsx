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
import type { VerifyResponse } from "@/types/auth"
import type { GraphNode } from "@/types/node"

const EMAIL_KEY = "graph-auth:email"

type AuthContextValue = {
  sessionToken: string | null
  email: string | null
  currentNode: GraphNode | null
  isAuthenticated: boolean
  isLoading: boolean
  register: (email: string) => Promise<void>
  verify: (token: string, name?: string, scopes?: string[]) => Promise<VerifyResponse>
  loginWithToken: (token: string) => Promise<void>
  logout: () => void
}

export const AuthContext = createContext<AuthContextValue | null>(null)

function persistSession(token: string, email: string): void {
  setSessionToken(token)
  sessionStorage.setItem(EMAIL_KEY, email)
}

function clearPersistedSession(): void {
  clearSessionToken()
  sessionStorage.removeItem(EMAIL_KEY)
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [sessionToken, setSessionTokenState] = useState<string | null>(() =>
    getSessionToken()
  )
  const [email, setEmail] = useState<string | null>(() =>
    sessionStorage.getItem(EMAIL_KEY)
  )
  const [currentNode, setCurrentNode] = useState<GraphNode | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  const bootstrap = useCallback(async () => {
    const token = getSessionToken()
    if (!token) {
      setIsLoading(false)
      return
    }

    setSessionTokenState(token)
    setEmail(sessionStorage.getItem(EMAIL_KEY))

    try {
      const { node } = await nodesApi.getCurrentNode()
      setCurrentNode(node)
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        clearPersistedSession()
        setSessionTokenState(null)
        setEmail(null)
        setCurrentNode(null)
      }
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    void bootstrap()
  }, [bootstrap])

  const applyVerifyResponse = useCallback((response: VerifyResponse) => {
    persistSession(response.sessionToken, response.email)
    setSessionTokenState(response.sessionToken)
    setEmail(response.email)
    setCurrentNode(null)
  }, [])

  const register = useCallback(async (registerEmail: string) => {
    await authApi.register({ email: registerEmail })
  }, [])

  const verify = useCallback(
    async (token: string, name?: string, scopes?: string[]) => {
      const response = await authApi.verifyEmail({ token, name, scopes })
      applyVerifyResponse(response)
      return response
    },
    [applyVerifyResponse]
  )

  const loginWithToken = useCallback(async (token: string) => {
    setSessionToken(token)
    setSessionTokenState(token)
    setEmail(null)
    sessionStorage.removeItem(EMAIL_KEY)
    try {
      const { node } = await nodesApi.getCurrentNode()
      setCurrentNode(node)
    } catch {
      setCurrentNode(null)
    }
  }, [])

  const logout = useCallback(() => {
    clearPersistedSession()
    setSessionTokenState(null)
    setEmail(null)
    setCurrentNode(null)
  }, [])

  const value = useMemo<AuthContextValue>(
    () => ({
      sessionToken,
      email,
      currentNode,
      isAuthenticated: Boolean(sessionToken),
      isLoading,
      register,
      verify,
      loginWithToken,
      logout,
    }),
    [
      sessionToken,
      email,
      currentNode,
      isLoading,
      register,
      verify,
      loginWithToken,
      logout,
    ]
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
