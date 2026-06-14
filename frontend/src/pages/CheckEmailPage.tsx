import { useState } from "react"
import { Link, useSearchParams } from "react-router-dom"
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
import { useAuth } from "@/hooks/useAuth"
import { ApiError } from "@/types/api"

export default function CheckEmailPage() {
  const [searchParams] = useSearchParams()
  const email = searchParams.get("email") ?? "your inbox"
  const { register } = useAuth()
  const [error, setError] = useState<string | null>(null)
  const [isResending, setIsResending] = useState(false)

  async function handleResend() {
    if (!searchParams.get("email")) return

    setError(null)
    setIsResending(true)

    try {
      await register(searchParams.get("email")!.trim())
      toast.success("Verification email resent")
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "Failed to resend email."
      setError(message)
    } finally {
      setIsResending(false)
    }
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Check your email</CardTitle>
        <CardDescription>
          We sent a verification link to <strong>{email}</strong>. Open it to
          set up this device and activate your account.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {error ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : (
          <p className="text-sm text-muted-foreground">
            The link will take you to device setup. No password is required.
          </p>
        )}
      </CardContent>
      <CardFooter className="flex flex-col gap-2">
        {searchParams.get("email") ? (
          <Button
            variant="outline"
            className="w-full"
            onClick={() => void handleResend()}
            disabled={isResending}
          >
            {isResending ? "Resending..." : "Resend verification email"}
          </Button>
        ) : null}
        <Button
          variant="ghost"
          className="w-full"
          render={<Link to="/register" />}
        >
          Use a different email
        </Button>
      </CardFooter>
    </Card>
  )
}
