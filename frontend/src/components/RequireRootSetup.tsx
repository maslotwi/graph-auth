import { Navigate, Outlet } from "react-router-dom"

import { useAuth } from "@/hooks/useAuth"

export function RequireRootSetup() {
  const { isAuthenticated, isLoading, requiresRootSetup, currentNode } =
    useAuth()

  if (isLoading) {
    return (
      <div className="flex min-h-svh items-center justify-center text-sm text-muted-foreground">
        Loading...
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/register" replace />
  }

  if (!requiresRootSetup && currentNode) {
    return <Navigate to="/" replace />
  }

  return <Outlet />
}
