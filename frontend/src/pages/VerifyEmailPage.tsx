import { useState } from "react"
import { Link, useNavigate, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useAuth } from "@/hooks/useAuth"
import { ApiError } from "@/types/api"
import { PERMISSIONS, type Permission } from "@/types/node"

export default function VerifyEmailPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { verify } = useAuth()
  const token = searchParams.get("token")

  const [deviceName, setDeviceName] = useState("My Device")
  const [scopes, setScopes] = useState<Permission[]>([...PERMISSIONS])
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  function toggleScope(scope: Permission, checked: boolean) {
    setScopes((current) =>
      checked ? [...current, scope] : current.filter((s) => s !== scope)
    )
  }

  async function handleSubmit(event: { preventDefault(): void }) {
    event.preventDefault()
    if (!token) return
    if (scopes.length === 0) {
      setError("Select at least one scope.")
      return
    }
    setError(null)
    setIsSubmitting(true)
    try {
      await verify(token, deviceName.trim() || "My Device", scopes)
      toast.success("Device verified")
      const returnTo = searchParams.get("return")
      void navigate(returnTo ?? "/", { replace: true })
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Verification failed.")
    } finally {
      setIsSubmitting(false)
    }
  }

  if (!token) {
    return (
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Invalid link</CardTitle>
          <CardDescription>
            This verification link is missing a token.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Alert variant="destructive">
            <AlertDescription>Verification token is missing.</AlertDescription>
          </Alert>
        </CardContent>
        <CardFooter>
          <Button className="w-full" render={<Link to="/register" />}>
            Back to registration
          </Button>
        </CardFooter>
      </Card>
    )
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Set up this device</CardTitle>
        <CardDescription>
          Name this device and choose its access scopes.
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit}>
        <CardContent className="flex flex-col gap-4">
          {error ? (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          ) : null}
          <div className="flex flex-col gap-2">
            <Label htmlFor="device-name">Device name</Label>
            <Input
              id="device-name"
              value={deviceName}
              onChange={(e) => setDeviceName(e.target.value)}
              placeholder="My Laptop"
              required
              autoFocus
            />
          </div>
          <div className="flex flex-col gap-3">
            <Label>Scopes</Label>
            {PERMISSIONS.map((scope) => (
              <label
                key={scope}
                className="flex items-center gap-2 text-sm capitalize"
              >
                <Checkbox
                  checked={scopes.includes(scope)}
                  onCheckedChange={(checked) =>
                    toggleScope(scope, checked === true)
                  }
                />
                {scope}
              </label>
            ))}
          </div>
        </CardContent>
        <CardFooter>
          <Button className="w-full" type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Verifying…" : "Activate this device"}
          </Button>
        </CardFooter>
      </form>
    </Card>
  )
}
