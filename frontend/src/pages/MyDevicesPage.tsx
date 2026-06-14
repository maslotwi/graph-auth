import { useEffect, useRef, useState } from "react"
import { toast } from "sonner"

import { generateDelegationCode } from "@/api/nodes"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { QRCode } from "@/components/ui/qr-code"
import { useAuth } from "@/hooks/useAuth"
import { ApiError } from "@/types/api"

const CODE_TTL = 120

type CodeData = {
  code: string
  link: string
}

export default function MyDevicesPage() {
  const { currentNode } = useAuth()
  const [codeData, setCodeData] = useState<CodeData | null>(null)
  const [secondsLeft, setSecondsLeft] = useState(0)
  const [isExpired, setIsExpired] = useState(false)
  const [isGenerating, setIsGenerating] = useState(false)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  function clearTimer() {
    if (timerRef.current) {
      clearInterval(timerRef.current)
      timerRef.current = null
    }
  }

  function startCountdown() {
    clearTimer()
    setSecondsLeft(CODE_TTL)
    setIsExpired(false)
    timerRef.current = setInterval(() => {
      setSecondsLeft((s) => {
        if (s <= 1) {
          clearTimer()
          setIsExpired(true)
          return 0
        }
        return s - 1
      })
    }, 1000)
  }

  useEffect(() => clearTimer, [])

  async function handleGenerate() {
    setIsGenerating(true)
    try {
      const result = await generateDelegationCode(
        currentNode?.id ?? "",
        currentNode?.permissions
      )
      setCodeData({ code: result.code, link: result.link })
      startCountdown()
    } catch (err) {
      toast.error(
        err instanceof ApiError ? err.message : "Failed to generate code."
      )
    } finally {
      setIsGenerating(false)
    }
  }

  const minutes = Math.floor(secondsLeft / 60)
  const seconds = secondsLeft % 60
  const countdown = `${minutes}:${seconds.toString().padStart(2, "0")}`
  const isRunning = codeData !== null && !isExpired

  return (
    <div className="flex flex-col gap-4 max-w-lg">
      <Card>
        <CardHeader>
          <CardTitle>My devices</CardTitle>
          <CardDescription>
            Manage devices connected to your account.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">No devices yet.</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Add a device</CardTitle>
          <CardDescription>
            Generate a one-time code and enter it on the new device at{" "}
            <span className="font-mono text-foreground">/join</span>.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col items-center gap-6">
          {isRunning && codeData ? (
            <>
              <div className="flex flex-col items-center gap-2">
                <span className="font-mono text-5xl font-bold tracking-[0.3em]">
                  {codeData.code}
                </span>
                <span
                  className={`text-sm tabular-nums ${
                    secondsLeft <= 30
                      ? "text-destructive"
                      : "text-muted-foreground"
                  }`}
                >
                  Expires in {countdown}
                </span>
              </div>
              <div className="rounded-xl border bg-muted/30 p-4">
                <QRCode value={codeData.link} size={180} />
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => void handleGenerate()}
                disabled={isGenerating}
              >
                Refresh code
              </Button>
            </>
          ) : (
            <div className="flex flex-col items-center gap-3">
              {isExpired && codeData ? (
                <p className="text-sm text-muted-foreground">Code expired.</p>
              ) : null}
              <Button
                onClick={() => void handleGenerate()}
                disabled={isGenerating}
              >
                {isGenerating ? "Generating…" : "Generate code"}
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
