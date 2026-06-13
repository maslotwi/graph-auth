import { useEffect, useState } from "react"
import {
  ReactFlow,
  Background,
  Handle,
  Position,
  type Node,
  type Edge,
  type NodeProps,
} from "@xyflow/react"
import "@xyflow/react/dist/style.css"

import { getNodeTree } from "@/api/nodes"
import { Badge } from "@/components/ui/badge"
import { ApiError } from "@/types/api"
import type { GraphNode } from "@/types/node"

// ── Layout ────────────────────────────────────────────────────────────────────

const NODE_WIDTH = 200
const NODE_HEIGHT = 110
const H_GAP = 40
const V_GAP = 80

function buildLayout(nodes: GraphNode[]): { rfNodes: Node[]; rfEdges: Edge[] } {
  const childrenOf = new Map<string | undefined, GraphNode[]>()
  for (const n of nodes) {
    const key = n.predecessorId
    childrenOf.set(key, [...(childrenOf.get(key) ?? []), n])
  }

  const positioned: Node[] = []

  function subtreeWidth(id: string): number {
    const kids = childrenOf.get(id) ?? []
    if (kids.length === 0) return NODE_WIDTH
    const childrenTotal =
      kids.reduce((sum, k) => sum + subtreeWidth(k.id), 0) +
      H_GAP * (kids.length - 1)
    return Math.max(NODE_WIDTH, childrenTotal)
  }

  function place(node: GraphNode, x: number, y: number) {
    positioned.push({
      id: node.id,
      type: "authNode",
      position: { x, y },
      data: node,
    })

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

function AuthNode({ data }: NodeProps) {
  const node = data as GraphNode
  return (
    <div
      className="rounded-lg border bg-card text-card-foreground shadow-sm"
      style={{ width: NODE_WIDTH, minHeight: NODE_HEIGHT, padding: "10px 12px" }}
    >
      <Handle type="target" position={Position.Top} />
      <div className="mb-1 flex items-center gap-1.5">
        {node.isRoot && (
          <Badge className="text-[10px]" variant="default">
            Root
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

// ── Page ──────────────────────────────────────────────────────────────────────

export default function GraphPage() {
  const [rfNodes, setRfNodes] = useState<Node[]>([])
  const [rfEdges, setRfEdges] = useState<Edge[]>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getNodeTree()
      .then(({ nodes }) => {
        const { rfNodes: n, rfEdges: e } = buildLayout(nodes)
        setRfNodes(n)
        setRfEdges(e)
      })
      .catch((err) => {
        setError(err instanceof ApiError ? err.message : "Failed to load graph.")
      })
      .finally(() => setLoading(false))
  }, [])

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
    <div style={{ height: "calc(100vh - 120px)" }}>
      <ReactFlow
        nodes={rfNodes}
        edges={rfEdges}
        nodeTypes={nodeTypes}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        proOptions={{ hideAttribution: false }}
      >
        <Background />
      </ReactFlow>
    </div>
  )
}
