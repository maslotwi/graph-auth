import { useEffect } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import { confirmSSO } from "@/api/auth"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { useAuth } from "@/hooks/useAuth"
import { ApiError } from "@/types/api"

export default function SSOConsentPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { isAuthenticated, isLoading } = useAuth()

  const clientId = searchParams.get("client_id") ?? ""
  const redirectUri = searchParams.get("redirect_uri") ?? ""
  const state = searchParams.get("state") ?? ""

  useEffect(() => {
    if (isLoading) return
    if (!isAuthenticated) {
      const returnTo = encodeURIComponent(window.location.pathname + window.location.search)
      void navigate(`/login?return=${returnTo}`, { replace: true })
    }
  }, [isAuthenticated, isLoading, navigate])

  if (!clientId || !redirectUri) {
    return (
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Invalid request</CardTitle>
          <CardDescription>
            Missing OAuth2 parameters. This link may be broken.
          </CardDescription>
        </CardHeader>
      </Card>
    )
  }

  async function handleConfirm() {
    try {
      const { redirect_to } = await confirmSSO({ client_id: clientId, redirect_uri: redirectUri, state })
      window.location.href = redirect_to
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to confirm login.")
    }
  }

  function handleDeny() {
    const url = `${redirectUri}?error=access_denied&state=${encodeURIComponent(state)}`
    window.location.href = url
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Sign in request</CardTitle>
        <CardDescription>
          An application wants to use your graph-auth identity.
        </CardDescription>
      </CardHeader>
      <Separator />
      <CardContent className="flex flex-col gap-4 pt-6">
        <div className="rounded-lg border bg-muted/30 px-4 py-3">
          <p className="text-xs text-muted-foreground">Application</p>
          <p className="mt-0.5 font-semibold">{clientId}</p>
        </div>
        <div className="rounded-lg border bg-muted/30 px-4 py-3">
          <p className="text-xs text-muted-foreground">Redirect to</p>
          <p className="mt-0.5 truncate font-mono text-xs">{redirectUri}</p>
        </div>
        <p className="text-sm text-muted-foreground">
          Confirming will share your identity with this application and redirect
          your browser back to it.
        </p>
      </CardContent>
      <Separator />
      <CardFooter className="flex gap-2 pt-4">
        <Button variant="outline" className="flex-1" onClick={handleDeny}>
          Deny
        </Button>
        <Button className="flex-1" onClick={handleConfirm}>
          Confirm
        </Button>
      </CardFooter>
    </Card>
  )
}
