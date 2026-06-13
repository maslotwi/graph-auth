import { useCallback, useEffect, useState } from "react"
import { X } from "lucide-react"
import { toast } from "sonner"
import {
  ReactFlow,
  Background,
  Handle,
  Position,
  type Node,
  type Edge,
  type NodeProps,
  type NodeMouseHandler,
} from "@xyflow/react"
import "@xyflow/react/dist/style.css"

import { createInvite, getNodeTree, invalidateNode } from "@/api/nodes"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { QRCode } from "@/components/ui/qr-code"
import { Separator } from "@/components/ui/separator"
import { useAuth } from "@/hooks/useAuth"
import { cn } from "@/lib/utils"
import { ApiError } from "@/types/api"
import type { GraphNode } from "@/types/node"

// ── Layout ────────────────────────────────────────────────────────────────────

const NODE_WIDTH = 200
const NODE_HEIGHT = 110
const H_GAP = 40
const V_GAP = 80

type AuthNodeData = GraphNode & { isCurrentNode: boolean }

function buildLayout(
  nodes: GraphNode[],
  currentNodeId: string | null
): { rfNodes: Node[]; rfEdges: Edge[] } {
  const childrenOf = new Map<string | undefined, GraphNode[]>()
  for (const n of nodes) {
    childrenOf.set(n.predecessorId, [
      ...(childrenOf.get(n.predecessorId) ?? []),
      n,
    ])
  }

  const positioned: Node[] = []

  function subtreeWidth(id: string): number {
    const kids = childrenOf.get(id) ?? []
    if (kids.length === 0) return NODE_WIDTH
    const total =
      kids.reduce((sum, k) => sum + subtreeWidth(k.id), 0) +
      H_GAP * (kids.length - 1)
    return Math.max(NODE_WIDTH, total)
  }

  function place(node: GraphNode, x: number, y: number) {
    const data: AuthNodeData = {
      ...node,
      isCurrentNode: node.id === currentNodeId,
    }
    positioned.push({ id: node.id, type: "authNode", position: { x, y }, data })

    const kids = childrenOf.get(node.id) ?? []
    let cursor = x - subtreeWidth(node.id) / 2 + NODE_WIDTH / 2
    for (const kid of kids) {
      const w = subtreeWidth(kid.id)
      place(kid, cursor + w / 2 - NODE_WIDTH / 2, y + NODE_HEIGHT + V_GAP)
      cursor += w + H_GAP
    }
  }

  const roots = childrenOf.get(undefined) ?? []
  let rootCursor = 0
  for (const root of roots) {
    const w = subtreeWidth(root.id)
    place(root, rootCursor, 0)
    rootCursor += w + H_GAP
  }

  const edges: Edge[] = nodes
    .filter((n) => n.predecessorId)
    .map((n) => ({
      id: `${n.predecessorId}-${n.id}`,
      source: n.predecessorId!,
      target: n.id,
      type: "smoothstep",
    }))

  return { rfNodes: positioned, rfEdges: edges }
}

// ── Custom node ───────────────────────────────────────────────────────────────

function AuthNode({ data, selected }: NodeProps) {
  const node = data as AuthNodeData
  return (
    <div
      className={cn(
        "rounded-lg border bg-card text-card-foreground shadow-sm transition-opacity",
        node.status === "invalidated" && "opacity-40",
        node.isCurrentNode && !selected && "ring-2 ring-amber-500",
        selected && "ring-2 ring-primary",
      )}
      style={{ width: NODE_WIDTH, minHeight: NODE_HEIGHT, padding: "10px 12px" }}
    >
      <Handle type="target" position={Position.Top} />
      <div className="mb-1 flex items-center gap-1.5">
        {node.isRoot && (
          <Badge className="text-[10px]" variant="default">Root</Badge>
        )}
        {node.isCurrentNode && (
          <Badge className="text-[10px] border-amber-500 text-amber-500" variant="outline">
            You
          </Badge>
        )}
        <Badge
          className="text-[10px]"
          variant={node.status === "active" ? "secondary" : "destructive"}
        >
          {node.status}
        </Badge>
      </div>
      <p className="truncate text-sm font-semibold">{node.label}</p>
      <div className="mt-1.5 flex flex-wrap gap-1">
        {node.permissions.map((p) => (
          <Badge key={p} variant="outline" className="text-[10px]">
            {p}
          </Badge>
        ))}
      </div>
      <Handle type="source" position={Position.Bottom} />
    </div>
  )
}

const nodeTypes = { authNode: AuthNode }

// ── Node detail panel ─────────────────────────────────────────────────────────

type NodeDetailPanelProps = {
  node: GraphNode
  isCurrentNode: boolean
  onClose: () => void
  onInvalidated: () => void
}

function NodeDetailPanel({
  node,
  isCurrentNode,
  onClose,
  onInvalidated,
}: NodeDetailPanelProps) {
  const [isInvalidating, setIsInvalidating] = useState(false)
  const [inviteUrl, setInviteUrl] = useState<string | null>(null)
  const [isGenerating, setIsGenerating] = useState(false)

  const canInvalidate =
    !node.isRoot && !isCurrentNode && node.status === "active"
  const canInvite = node.status === "active"

  async function handleInvalidate() {
    setIsInvalidating(true)
    try {
      await invalidateNode(node.id)
      toast.success(`${node.label} invalidated`)
      onInvalidated()
    } catch (err) {
      toast.error(
        err instanceof ApiError ? err.message : "Failed to invalidate node."
      )
    } finally {
      setIsInvalidating(false)
    }
  }

  async function handleGenerateInvite() {
    setIsGenerating(true)
    try {
      const { token } = await createInvite(node.id)
      setInviteUrl(`${window.location.origin}/verify?token=${token}`)
    } catch (err) {
      toast.error(
        err instanceof ApiError ? err.message : "Failed to generate invite."
      )
    } finally {
      setIsGenerating(false)
    }
  }

  const invalidateHint = node.isRoot
    ? "Root nodes cannot be invalidated."
    : isCurrentNode
      ? "Cannot invalidate your current session node."
      : node.status === "invalidated"
        ? "Already invalidated."
        : null

  return (
    <div className="absolute right-0 top-0 z-10 flex h-full w-72 flex-col border-l border-border bg-card shadow-lg">
      <div className="flex items-center justify-between px-4 py-3">
        <h3 className="text-sm font-semibold">Node details</h3>
        <Button variant="ghost" size="icon" onClick={onClose} className="h-7 w-7">
          <X className="h-4 w-4" />
        </Button>
      </div>
      <Separator />
      <div className="flex flex-1 flex-col gap-4 overflow-y-auto p-4">
        <div>
          <p className="mb-1 text-xs text-muted-foreground">Label</p>
          <p className="text-sm font-medium">{node.label}</p>
        </div>
        <div>
          <p className="mb-1.5 text-xs text-muted-foreground">Badges</p>
          <div className="flex flex-wrap gap-1.5">
            {node.isRoot && <Badge variant="default">Root</Badge>}
            {isCurrentNode && (
              <Badge variant="outline" className="border-amber-500 text-amber-500">
                Current session
              </Badge>
            )}
            <Badge variant={node.status === "active" ? "secondary" : "destructive"}>
              {node.status}
            </Badge>
          </div>
        </div>
        <div>
          <p className="mb-1.5 text-xs text-muted-foreground">Permissions</p>
          <div className="flex flex-wrap gap-1">
            {node.permissions.map((p) => (
              <Badge key={p} variant="outline" className="text-xs">
                {p}
              </Badge>
            ))}
          </div>
        </div>
        <div>
          <p className="mb-1 text-xs text-muted-foreground">Node ID</p>
          <p className="truncate font-mono text-xs text-muted-foreground">
            {node.id}
          </p>
        </div>

        <Separator />

        <div className="flex flex-col gap-2">
          <p className="text-xs text-muted-foreground">Add a child device</p>
          {inviteUrl ? (
            <div className="flex flex-col items-center gap-2">
              <div className="rounded-lg border bg-muted/30 p-3">
                <QRCode value={inviteUrl} size={160} />
              </div>
              <p className="text-center text-xs text-muted-foreground">
                Scan with a new device to join as a child node
              </p>
              <Button
                variant="ghost"
                size="sm"
                className="w-full text-xs"
                onClick={() => setInviteUrl(null)}
              >
                Revoke &amp; close
              </Button>
            </div>
          ) : (
            <Button
              variant="outline"
              size="sm"
              className="w-full"
              disabled={!canInvite || isGenerating}
              onClick={handleGenerateInvite}
            >
              {isGenerating ? "Generating…" : "Generate invite QR"}
            </Button>
          )}
        </div>
      </div>
      <Separator />
      <div className="p-4">
        <Button
          variant="destructive"
          className="w-full"
          disabled={!canInvalidate || isInvalidating}
          onClick={handleInvalidate}
        >
          {isInvalidating ? "Invalidating…" : "Invalidate node"}
        </Button>
        {invalidateHint && (
          <p className="mt-2 text-center text-xs text-muted-foreground">
            {invalidateHint}
          </p>
        )}
      </div>
    </div>
  )
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default function GraphPage() {
  const { currentNode } = useAuth()
  const [allNodes, setAllNodes] = useState<GraphNode[]>([])
  const [rfNodes, setRfNodes] = useState<Node[]>([])
  const [rfEdges, setRfEdges] = useState<Edge[]>([])
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  const loadTree = useCallback(() => {
    return getNodeTree()
      .then(({ nodes }) => {
        setAllNodes(nodes)
        const { rfNodes: n, rfEdges: e } = buildLayout(
          nodes,
          currentNode?.id ?? null
        )
        setRfNodes(n)
        setRfEdges(e)
      })
      .catch((err) => {
        setError(err instanceof ApiError ? err.message : "Failed to load graph.")
      })
      .finally(() => setLoading(false))
  }, [currentNode?.id])

  useEffect(() => {
    void loadTree()
  }, [loadTree])

  const onNodeClick: NodeMouseHandler = useCallback((_event, node) => {
    setSelectedNodeId((prev) => (prev === node.id ? null : node.id))
  }, [])

  const selectedNode = allNodes.find((n) => n.id === selectedNodeId) ?? null

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        Loading graph…
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-destructive">
        {error}
      </div>
    )
  }

  return (
    <div className="relative" style={{ height: "calc(100vh - 120px)" }}>
      <ReactFlow
        nodes={rfNodes}
        edges={rfEdges}
        nodeTypes={nodeTypes}
        onNodeClick={onNodeClick}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        proOptions={{ hideAttribution: false }}
      >
        <Background />
      </ReactFlow>

      {selectedNode && (
        <NodeDetailPanel
          node={selectedNode}
          isCurrentNode={selectedNode.id === currentNode?.id}
          onClose={() => setSelectedNodeId(null)}
          onInvalidated={() => {
            void loadTree()
            setSelectedNodeId(null)
          }}
        />
      )}
    </div>
  )
}
