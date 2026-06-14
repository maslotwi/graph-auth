import { useCallback, useEffect, useRef, useState } from "react"
import { toast } from "sonner"

import { generateDelegationCode, getNodeTree, invalidateNode } from "@/api/nodes"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { QRCode } from "@/components/ui/qr-code"
import { Separator } from "@/components/ui/separator"
import { useAuth } from "@/hooks/useAuth"
import { ApiError } from "@/types/api"
import type { GraphNode } from "@/types/node"

const CODE_TTL = 120

type CodeData = {
  code: string
  link: string
}

export default function MyDevicesPage() {
  const { currentNode } = useAuth()
  const [nodes, setNodes] = useState<GraphNode[]>([])
  const [isLoadingNodes, setIsLoadingNodes] = useState(true)
  const [codeData, setCodeData] = useState<CodeData | null>(null)
  const [secondsLeft, setSecondsLeft] = useState(0)
  const [isExpired, setIsExpired] = useState(false)
  const [isGenerating, setIsGenerating] = useState(false)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const loadNodes = useCallback(async () => {
    try {
      const { nodes: tree } = await getNodeTree()
      setNodes(tree)
    } catch {
      toast.error("Failed to load devices.")
    } finally {
      setIsLoadingNodes(false)
    }
  }, [])

  useEffect(() => {
    void loadNodes()
  }, [loadNodes])

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
      const result = await generateDelegationCode()
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

  async function handleInvalidate(node: GraphNode) {
    try {
      await invalidateNode(node.id)
      toast.success(`${node.label} invalidated`)
      void loadNodes()
    } catch (err) {
      toast.error(
        err instanceof ApiError ? err.message : "Failed to invalidate node."
      )
    }
  }

  const minutes = Math.floor(secondsLeft / 60)
  const seconds = secondsLeft % 60
  const countdown = `${minutes}:${seconds.toString().padStart(2, "0")}`
  const isRunning = codeData !== null && !isExpired

  const otherNodes = nodes.filter((n) => n.id !== currentNode?.id)

  return (
    <div className="flex flex-col gap-4 max-w-lg">
      <Card>
        <CardHeader>
          <CardTitle>My devices</CardTitle>
          <CardDescription>
            All sessions in your account tree.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {isLoadingNodes ? (
            <p className="text-sm text-muted-foreground">Loading…</p>
          ) : otherNodes.length === 0 ? (
            <p className="text-sm text-muted-foreground">No other devices yet.</p>
          ) : (
            otherNodes.map((node) => (
              <div
                key={node.id}
                className="flex items-center justify-between gap-3 rounded-lg border px-4 py-3"
              >
                <div className="flex flex-col gap-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="text-sm font-medium truncate">{node.label}</span>
                    {node.isRoot && <Badge className="text-[10px]">Root</Badge>}
                    <Badge
                      variant={node.status === "active" ? "secondary" : "destructive"}
                      className="text-[10px]"
                    >
                      {node.status}
                    </Badge>
                  </div>
                  <div className="flex flex-wrap gap-1">
                    {node.permissions.map((p) => (
                      <Badge key={p} variant="outline" className="text-[10px]">
                        {p}
                      </Badge>
                    ))}
                  </div>
                </div>
                {node.status === "active" && (
                  <Button
                    variant="destructive"
                    size="sm"
                    className="shrink-0"
                    onClick={() => void handleInvalidate(node)}
                  >
                    Invalidate
                  </Button>
                )}
              </div>
            ))
          )}
        </CardContent>
      </Card>

      <Separator />

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
