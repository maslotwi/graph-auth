import { useEffect, useState } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import { consumeDelegationCode } from "@/api/nodes"
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

type Step = "code" | "details"

export default function JoinPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { loginWithToken } = useAuth()

  const [step, setStep] = useState<Step>(
    searchParams.get("code") ? "details" : "code"
  )
  const [code, setCode] = useState(searchParams.get("code") ?? "")
  const [deviceName, setDeviceName] = useState("")
  const [scopes, setScopes] = useState<Permission[]>([...PERMISSIONS])
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    const codeFromUrl = searchParams.get("code")
    if (codeFromUrl) {
      setCode(codeFromUrl)
      setStep("details")
    }
  }, [searchParams])

  function toggleScope(scope: Permission, checked: boolean) {
    setScopes((current) =>
      checked ? [...current, scope] : current.filter((s) => s !== scope)
    )
  }

  function handleCodeSubmit(event: { preventDefault(): void }) {
    event.preventDefault()
    if (code.length !== 6) {
      setError("Enter a 6-digit code.")
      return
    }
    setError(null)
    setStep("details")
  }

  async function handleDetailsSubmit(event: { preventDefault(): void }) {
    event.preventDefault()
    if (scopes.length === 0) {
      setError("Select at least one permission.")
      return
    }
    setError(null)
    setIsSubmitting(true)
    try {
      const { session_token } = await consumeDelegationCode(
        code.trim(),
        deviceName.trim() || "New Device",
        scopes
      )
      await loginWithToken(session_token)
      toast.success("Device linked successfully")
      void navigate("/", { replace: true })
    } catch (err) {
      setError(
        err instanceof ApiError ? err.message : "Invalid or expired code."
      )
      setStep("code")
    } finally {
      setIsSubmitting(false)
    }
  }

  if (step === "code") {
    return (
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Join as a device</CardTitle>
          <CardDescription>
            Enter the 6-digit code shown on an authorized device.
          </CardDescription>
        </CardHeader>
        <form onSubmit={handleCodeSubmit}>
          <CardContent className="flex flex-col gap-4">
            {error ? (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            ) : null}
            <div className="flex flex-col gap-2">
              <Label htmlFor="code">6-digit code</Label>
              <Input
                id="code"
                inputMode="numeric"
                pattern="[0-9]{6}"
                maxLength={6}
                placeholder="123456"
                value={code}
                onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))}
                className="text-center font-mono text-2xl tracking-[0.5em]"
                required
                autoFocus
              />
            </div>
          </CardContent>
          <CardFooter>
            <Button className="w-full" type="submit">
              Continue
            </Button>
          </CardFooter>
        </form>
      </Card>
    )
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Set up this device</CardTitle>
        <CardDescription>
          Name this device and choose its permissions.
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleDetailsSubmit}>
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
              placeholder="My Phone"
              value={deviceName}
              onChange={(e) => setDeviceName(e.target.value)}
              autoFocus
            />
          </div>
          <div className="flex flex-col gap-3">
            <Label>Permissions</Label>
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
        <CardFooter className="flex gap-2">
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              setError(null)
              setStep("code")
            }}
          >
            Back
          </Button>
          <Button className="flex-1" type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Linking…" : "Link this device"}
          </Button>
        </CardFooter>
      </form>
    </Card>
  )
}
