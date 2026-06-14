import { useEffect, useState } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import { confirmSSO } from "@/api/auth"
import { getClientInfo } from "@/api/clients"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { useAuth } from "@/hooks/useAuth"
import { ApiError } from "@/types/api"

const SCOPE_LABELS: Record<string, { label: string; description: string }> = {
  openid:  { label: "Identity",      description: "Verify who you are" },
  profile: { label: "Profile",       description: "Read your display name and picture" },
  email:   { label: "Email",         description: "Read your email address" },
  read:    { label: "Read",          description: "Read access to your account data" },
}

export default function SSOConsentPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { isAuthenticated, isLoading, email } = useAuth()

  const clientId   = searchParams.get("client_id")   ?? ""
  const redirectUri = searchParams.get("redirect_uri") ?? ""
  const state      = searchParams.get("state")        ?? ""
  const scope      = searchParams.get("scope")        ?? ""

  const [appName, setAppName] = useState<string | null>(null)
  const [isConfirming, setIsConfirming] = useState(false)

  const requestedScopes = scope
    .split(" ")
    .map((s) => s.trim())
    .filter(Boolean)

  const redirectHost = (() => {
    try { return new URL(redirectUri).hostname } catch { return redirectUri }
  })()

  useEffect(() => {
    if (isLoading) return
    if (!isAuthenticated) {
      const returnTo = encodeURIComponent(window.location.pathname + window.location.search)
      void navigate(`/join?return=${returnTo}`, { replace: true })
    }
  }, [isAuthenticated, isLoading, navigate])

  useEffect(() => {
    if (!clientId) return
    getClientInfo(clientId)
      .then(({ name }) => setAppName(name))
      .catch(() => setAppName(null))
  }, [clientId])

  if (!clientId || !redirectUri) {
    return (
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Invalid request</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            Missing OAuth2 parameters. This link may be broken.
          </p>
        </CardContent>
      </Card>
    )
  }

  async function handleConfirm() {
    setIsConfirming(true)
    try {
      const { redirect_to } = await confirmSSO({ client_id: clientId, redirect_uri: redirectUri, state, scope })
      window.location.href = redirect_to
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to confirm login.")
      setIsConfirming(false)
    }
  }

  function handleDeny() {
    window.location.href = `${redirectUri}?error=access_denied&state=${encodeURIComponent(state)}`
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader className="pb-4">
        <CardTitle className="text-xl">
          {appName ?? clientId}
          <span className="ml-1 font-normal text-muted-foreground"> wants to sign you in</span>
        </CardTitle>
      </CardHeader>

      <Separator />

      <CardContent className="flex flex-col gap-4 pt-5">
        <div className="flex flex-col gap-1 rounded-lg border bg-muted/30 px-4 py-3">
          <p className="text-xs text-muted-foreground">Signed in as</p>
          <p className="text-sm font-medium">{email ?? "—"}</p>
        </div>

        {requestedScopes.length > 0 && (
          <div className="flex flex-col gap-2">
            <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Permissions requested
            </p>
            {requestedScopes.map((s) => {
              const meta = SCOPE_LABELS[s]
              return (
                <div key={s} className="flex items-center gap-3 rounded-lg border px-3 py-2.5">
                  <Badge variant="secondary" className="shrink-0 text-[10px]">
                    {meta?.label ?? s}
                  </Badge>
                  <span className="text-sm text-muted-foreground">
                    {meta?.description ?? s}
                  </span>
                </div>
              )
            })}
          </div>
        )}

        <div className="flex flex-col gap-1 rounded-lg border bg-muted/30 px-4 py-3">
          <p className="text-xs text-muted-foreground">Redirects to</p>
          <p className="font-mono text-xs">{redirectHost}</p>
        </div>
      </CardContent>

      <Separator />

      <CardFooter className="flex gap-2 pt-4">
        <Button variant="outline" className="flex-1" onClick={handleDeny} disabled={isConfirming}>
          Deny
        </Button>
        <Button className="flex-1" onClick={() => void handleConfirm()} disabled={isConfirming}>
          {isConfirming ? "Confirming…" : "Confirm"}
        </Button>
      </CardFooter>
    </Card>
  )
}
