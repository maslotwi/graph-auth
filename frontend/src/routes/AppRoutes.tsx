import { lazy, Suspense, type ComponentType, type ReactNode } from "react"
import { type RouteObject } from "react-router-dom"

import { RequireAuth } from "@/components/RequireAuth"
import { RequireRootSetup } from "@/components/RequireRootSetup"
import AppLayout from "@/layouts/AppLayout"
import PublicLayout from "@/layouts/PublicLayout"

function lazyPage(
  importer: () => Promise<{ default: ComponentType }>
) {
  return lazy(importer)
}

const HomePage = lazyPage(() => import("@/pages/HomePage"))
const LoginPage = lazyPage(() => import("@/pages/LoginPage"))
const RegisterPage = lazyPage(() => import("@/pages/RegisterPage"))
const MyDevicesPage = lazyPage(() => import("@/pages/MyDevicesPage"))
const CheckEmailPage = lazyPage(() => import("@/pages/CheckEmailPage"))
const VerifyEmailPage = lazyPage(() => import("@/pages/VerifyEmailPage"))
const RootNodeSetupPage = lazyPage(() => import("@/pages/RootNodeSetupPage"))
const GraphPage = lazyPage(() => import("@/pages/GraphPage"))
const JoinPage = lazyPage(() => import("@/pages/JoinPage"))
const SSOConsentPage = lazyPage(() => import("@/pages/SSOConsentPage"))

function PageLoader() {
  return (
    <div className="flex min-h-[40vh] items-center justify-center text-sm text-muted-foreground">
      Loading...
    </div>
  )
}

function withSuspense(element: ReactNode) {
  return <Suspense fallback={<PageLoader />}>{element}</Suspense>
}

export const routes: RouteObject[] = [
  {
    element: <PublicLayout />,
    children: [
      { path: "/login", element: withSuspense(<LoginPage />) },
      { path: "/register", element: withSuspense(<RegisterPage />) },
      { path: "/check-email", element: withSuspense(<CheckEmailPage />) },
      { path: "/verify", element: withSuspense(<VerifyEmailPage />) },
      { path: "/join", element: withSuspense(<JoinPage />) },
      { path: "/sso/consent", element: withSuspense(<SSOConsentPage />) },
    ],
  },
  {
    element: <RequireRootSetup />,
    children: [
      {
        path: "/setup/root",
        element: withSuspense(<RootNodeSetupPage />),
      },
    ],
  },
  {
    element: <RequireAuth />,
    children: [
      {
        element: <AppLayout />,
        children: [
          { path: "/", element: withSuspense(<HomePage />) },
          { path: "/devices", element: withSuspense(<MyDevicesPage />) },
          { path: "/graph", element: withSuspense(<GraphPage />) },
        ],
      },
    ],
  },
]
