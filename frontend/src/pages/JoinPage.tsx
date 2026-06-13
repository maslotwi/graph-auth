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
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useAuth } from "@/hooks/useAuth"
import { ApiError } from "@/types/api"

export default function JoinPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { loginWithToken } = useAuth()
  const [code, setCode] = useState(searchParams.get("code") ?? "")
  const [deviceName, setDeviceName] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    const codeFromUrl = searchParams.get("code")
    if (codeFromUrl) setCode(codeFromUrl)
  }, [searchParams])

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (code.length !== 6) {
      setError("Enter a 6-digit code.")
      return
    }
    setError(null)
    setIsSubmitting(true)
    try {
      const { session_token } = await consumeDelegationCode(
        code.trim(),
        deviceName.trim() || "New Device"
      )
      await loginWithToken(session_token)
      toast.success("Device linked successfully")
      void navigate("/", { replace: true })
    } catch (err) {
      setError(
        err instanceof ApiError ? err.message : "Invalid or expired code."
      )
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Join as a device</CardTitle>
        <CardDescription>
          Enter the 6-digit code shown on an authorized device.
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
              autoFocus={!searchParams.get("code")}
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="device-name">
              Device name{" "}
              <span className="text-muted-foreground">(optional)</span>
            </Label>
            <Input
              id="device-name"
              placeholder="My Laptop"
              value={deviceName}
              onChange={(e) => setDeviceName(e.target.value)}
            />
          </div>
        </CardContent>
        <CardFooter>
          <Button className="w-full" type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Linking…" : "Link this device"}
          </Button>
        </CardFooter>
      </form>
    </Card>
  )
}
