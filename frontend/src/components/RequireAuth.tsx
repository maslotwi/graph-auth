import { Navigate, Outlet, useLocation } from "react-router-dom"

import { useAuth } from "@/hooks/useAuth"

export function RequireAuth() {
  const { isAuthenticated, isLoading, requiresRootSetup, currentNode } =
    useAuth()
  const location = useLocation()

  if (isLoading) {
    return (
      <div className="flex min-h-svh items-center justify-center text-sm text-muted-foreground">
        Loading...
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/register" replace state={{ from: location }} />
  }

  if (requiresRootSetup || !currentNode) {
    return <Navigate to="/setup/root" replace />
  }

  return <Outlet />
}
