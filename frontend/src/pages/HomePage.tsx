import { useCallback, useEffect, useRef, useState } from "react"
import { toast } from "sonner"

import { createClient, getClients, type ClientCredentials, type ClientSummary } from "@/api/clients"
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
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { QRCode } from "@/components/ui/qr-code"
import { Separator } from "@/components/ui/separator"
import { useAuth } from "@/hooks/useAuth"
import { ApiError } from "@/types/api"
import type { GraphNode } from "@/types/node"

const CODE_TTL = 120

type CodeData = { code: string; link: string }

// ── Devices column ────────────────────────────────────────────────────────────

function DevicesSection() {
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

  useEffect(() => { void loadNodes() }, [loadNodes])

  function clearTimer() {
    if (timerRef.current) { clearInterval(timerRef.current); timerRef.current = null }
  }

  function startCountdown() {
    clearTimer()
    setSecondsLeft(CODE_TTL)
    setIsExpired(false)
    timerRef.current = setInterval(() => {
      setSecondsLeft((s) => {
        if (s <= 1) { clearTimer(); setIsExpired(true); return 0 }
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
      toast.error(err instanceof ApiError ? err.message : "Failed to generate code.")
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
      toast.error(err instanceof ApiError ? err.message : "Failed to invalidate node.")
    }
  }

  const minutes = Math.floor(secondsLeft / 60)
  const secs = secondsLeft % 60
  const countdown = `${minutes}:${secs.toString().padStart(2, "0")}`
  const isRunning = codeData !== null && !isExpired
  const otherNodes = nodes.filter((n) => n.id !== currentNode?.id)

  // Build the set of ancestor IDs for the current node so we can
  // hide the Invalidate button on nodes that are above us in the tree.
  const ancestorIds = new Set<string>()
  if (currentNode) {
    const nodeMap = new Map(nodes.map((n) => [n.id, n]))
    let cursor: GraphNode | undefined = currentNode
    while (cursor?.predecessorId) {
      ancestorIds.add(cursor.predecessorId)
      cursor = nodeMap.get(cursor.predecessorId)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardHeader>
          <CardTitle>My devices</CardTitle>
          <CardDescription>All sessions in your account tree.</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {currentNode && (
            <>
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">This device</p>
              <div className="flex items-center gap-3 rounded-lg border border-amber-500/40 bg-amber-500/5 px-4 py-3">
                <div className="flex flex-col gap-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="text-sm font-medium truncate">{currentNode.label}</span>
                    {currentNode.isRoot && <Badge className="text-[10px]">Root</Badge>}
                    <Badge variant="secondary" className="text-[10px]">active</Badge>
                  </div>
                  <div className="flex flex-wrap gap-1">
                    {currentNode.permissions.map((p) => (
                      <Badge key={p} variant="outline" className="text-[10px]">{p}</Badge>
                    ))}
                  </div>
                </div>
              </div>
              {otherNodes.length > 0 && (
                <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide pt-1">Other devices</p>
              )}
            </>
          )}
          {isLoadingNodes ? (
            <p className="text-sm text-muted-foreground">Loading…</p>
          ) : otherNodes.length === 0 && !currentNode ? (
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
                      <Badge key={p} variant="outline" className="text-[10px]">{p}</Badge>
                    ))}
                  </div>
                </div>
                {node.status === "active" && !ancestorIds.has(node.id) && (
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
                <span className={`text-sm tabular-nums ${secondsLeft <= 30 ? "text-destructive" : "text-muted-foreground"}`}>
                  Expires in {countdown}
                </span>
              </div>
              <div className="rounded-xl border bg-muted/30 p-4">
                <QRCode value={codeData.link} size={180} />
              </div>
              <Button variant="outline" size="sm" onClick={() => void handleGenerate()} disabled={isGenerating}>
                Refresh code
              </Button>
            </>
          ) : (
            <div className="flex flex-col items-center gap-3">
              {isExpired && codeData && (
                <p className="text-sm text-muted-foreground">Code expired.</p>
              )}
              <Button onClick={() => void handleGenerate()} disabled={isGenerating}>
                {isGenerating ? "Generating…" : "Generate code"}
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

// ── Apps column ───────────────────────────────────────────────────────────────

const apiBase =
  (import.meta.env.VITE_ISSUER_URL as string) ||
  (import.meta.env.VITE_API_BASE_URL as string) ||
  window.location.origin

const OAUTH_ENDPOINTS = [
  { label: "Authorization", value: `${apiBase}/api/oauth/authorize` },
  { label: "Token",         value: `${apiBase}/api/oauth/token` },
  { label: "Userinfo",      value: `${apiBase}/api/oauth/userinfo` },
  { label: "Discovery",     value: `${apiBase}/.well-known/openid-configuration` },
]

function CopyRow({ label, value }: { label: string; value: string }) {
  const [copied, setCopied] = useState(false)

  function handleCopy() {
    const done = () => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }

    if (navigator.clipboard) {
      navigator.clipboard.writeText(value).then(done).catch(() => fallback())
    } else {
      fallback()
    }

    function fallback() {
      const el = document.createElement("textarea")
      el.value = value
      el.style.cssText = "position:fixed;opacity:0;pointer-events:none"
      document.body.appendChild(el)
      el.focus()
      el.select()
      const ok = document.execCommand("copy")
      document.body.removeChild(el)
      if (ok) done()
      else toast.error("Copy failed — please select and copy manually.")
    }
  }

  return (
    <div className="flex flex-col gap-1">
      <Label className="text-xs text-muted-foreground">{label}</Label>
      <div className="flex items-center gap-2">
        <code className="flex-1 rounded bg-muted px-2 py-1.5 font-mono text-xs break-all">
          {value}
        </code>
        <Button
          type="button"
          variant="outline"
          size="sm"
          className="shrink-0"
          onClick={handleCopy}
        >
          {copied ? "Copied!" : "Copy"}
        </Button>
      </div>
    </div>
  )
}

function EndpointsPanel() {
  const [open, setOpen] = useState(false)

  return (
    <div className="flex flex-col gap-2">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex items-center justify-between text-xs font-medium uppercase tracking-wide text-muted-foreground hover:text-foreground transition-colors"
      >
        <span>Provider endpoints</span>
        <span className="text-base leading-none">{open ? "−" : "+"}</span>
      </button>
      {open && OAUTH_ENDPOINTS.map(({ label, value }) => (
        <CopyRow key={label} label={label} value={value} />
      ))}
    </div>
  )
}

function AppRow({ app }: { app: ClientSummary }) {
  return (
    <div className="flex flex-col gap-1 rounded-lg border px-4 py-3">
      <div className="flex items-center justify-between gap-2">
        <span className="text-sm font-medium">{app.name}</span>
        <span className="font-mono text-[10px] text-muted-foreground truncate max-w-[140px]">
          {app.client_id}
        </span>
      </div>
      {app.redirect_uris.length > 0 && (
        <div className="flex flex-wrap gap-1 pt-0.5">
          {app.redirect_uris.map((uri) => (
            <span key={uri} className="rounded bg-muted px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">
              {uri}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}

function AppsSection() {
  const { currentNode } = useAuth()
  const [apps, setApps] = useState<ClientSummary[]>([])
  const [isLoadingApps, setIsLoadingApps] = useState(true)
  const [name, setName] = useState("")
  const [redirectUrisRaw, setRedirectUrisRaw] = useState("")
  const [isCreating, setIsCreating] = useState(false)
  const [credentials, setCredentials] = useState<ClientCredentials | null>(null)

  const hasClientsScope = currentNode?.permissions.includes("clients") ?? false

  const loadApps = useCallback(async () => {
    try {
      const list = await getClients()
      setApps(list)
    } catch {
      // not a hard failure — list stays empty
    } finally {
      setIsLoadingApps(false)
    }
  }, [])

  useEffect(() => { void loadApps() }, [loadApps])

  const parsedUris = redirectUrisRaw
    .split("\n")
    .map((s) => s.trim())
    .filter(Boolean)

  async function handleCreate(e: { preventDefault(): void }) {
    e.preventDefault()
    const trimmed = name.trim()
    if (!trimmed || parsedUris.length === 0) return
    setIsCreating(true)
    try {
      const creds = await createClient(trimmed, parsedUris)
      setCredentials(creds)
      setName("")
      setRedirectUrisRaw("")
      void loadApps()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to create app.")
    } finally {
      setIsCreating(false)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardHeader>
          <CardTitle>My apps</CardTitle>
          <CardDescription>
            Applications connected to your Graph Auth identity.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {isLoadingApps ? (
            <p className="text-sm text-muted-foreground">Loading…</p>
          ) : apps.length === 0 ? (
            <p className="text-sm text-muted-foreground">No apps registered yet.</p>
          ) : (
            apps.map((app) => <AppRow key={app.client_id} app={app} />)
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Register an app</CardTitle>
          <CardDescription>
            Create an OAuth2 client to integrate your application with Graph Auth
            sign-in.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-5">
          <EndpointsPanel />

          <Separator />

          {!hasClientsScope ? (
            <p className="text-sm text-muted-foreground">
              Your session needs the{" "}
              <Badge variant="outline" className="text-[10px]">clients</Badge>{" "}
              scope to register apps.
            </p>
          ) : credentials ? (
            <div className="flex flex-col gap-4">
              <div className="rounded-lg border border-amber-500/40 bg-amber-500/5 px-3 py-2">
                <p className="text-xs text-amber-600 dark:text-amber-400">
                  Save the client secret now — it will not be shown again.
                </p>
              </div>
              <CopyRow label="App name" value={credentials.name} />
              <CopyRow label="Client ID" value={credentials.client_id} />
              <CopyRow label="Client Secret" value={credentials.client_secret} />
              <CopyRow label="Redirect URIs" value={credentials.redirect_uris.join("\n")} />
              <Button variant="outline" size="sm" onClick={() => setCredentials(null)}>
                Done
              </Button>
            </div>
          ) : (
            <form onSubmit={handleCreate} className="flex flex-col gap-3">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="app-name">App name</Label>
                <Input
                  id="app-name"
                  placeholder="My Application"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="redirect-uris">Redirect URIs</Label>
                <textarea
                  id="redirect-uris"
                  rows={3}
                  placeholder={"https://myapp.com/callback\nhttp://localhost:3000/callback"}
                  value={redirectUrisRaw}
                  onChange={(e) => setRedirectUrisRaw(e.target.value)}
                  className="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring font-mono resize-none"
                />
                <p className="text-xs text-muted-foreground">One URI per line.</p>
              </div>
              <Button
                type="submit"
                disabled={isCreating || !name.trim() || parsedUris.length === 0}
                className="self-start"
              >
                {isCreating ? "Registering…" : "Register app"}
              </Button>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default function HomePage() {
  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
      <DevicesSection />
      <AppsSection />
    </div>
  )
}
