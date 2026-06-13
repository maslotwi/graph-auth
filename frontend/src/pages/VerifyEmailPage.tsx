import { useEffect, useState } from "react"
import { Link, useNavigate, useSearchParams } from "react-router-dom"

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

export default function VerifyEmailPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { verify } = useAuth()
  const [error, setError] = useState<string | null>(null)
  const [isVerifying, setIsVerifying] = useState(true)

  useEffect(() => {
    const token = searchParams.get("token")
    if (!token) {
      setError("Verification token is missing.")
      setIsVerifying(false)
      return
    }

    async function runVerification() {
      try {
        const response = await verify(token!)
        const returnTo = searchParams.get("return")
        const destination = returnTo
          ? returnTo
          : response.requiresRootSetup
            ? "/setup/root"
            : "/"
        void navigate(destination, { replace: true })
      } catch (err) {
        const message =
          err instanceof ApiError ? err.message : "Verification failed."
        setError(message)
        setIsVerifying(false)
      }
    }

    void runVerification()
  }, [searchParams, verify, navigate])

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Verifying email</CardTitle>
        <CardDescription>
          {isVerifying
            ? "Please wait while we verify your email link."
            : error
              ? "We could not verify your email."
              : "Redirecting..."}
        </CardDescription>
      </CardHeader>
      <CardContent>
        {error ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : null}
      </CardContent>
      {error ? (
        <CardFooter>
          <Button className="w-full" render={<Link to="/register" />}>
            Back to registration
          </Button>
        </CardFooter>
      ) : null}
    </Card>
  )
}
