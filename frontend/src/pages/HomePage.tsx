import { Link } from "react-router-dom"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { useAuth } from "@/hooks/useAuth"

export default function HomePage() {
  const { currentNode, email } = useAuth()

  if (!currentNode) {
    return null
  }

  return (
    <div className="mx-auto flex w-full max-w-2xl flex-col gap-4">
      <Card>
        <CardHeader>
          <CardTitle>Dashboard</CardTitle>
          <CardDescription>
            Signed in as {email ?? "unknown"} via node {currentNode.label}
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <div className="flex flex-wrap items-center gap-2">
            {currentNode.isRoot ? <Badge>Root node</Badge> : null}
            <Badge
              variant={
                currentNode.status === "active" ? "secondary" : "destructive"
              }
            >
              {currentNode.status}
            </Badge>
          </div>
          <div>
            <p className="mb-2 text-sm font-medium">Permissions</p>
            <div className="flex flex-wrap gap-2">
              {currentNode.permissions.map((permission) => (
                <Badge key={permission} variant="outline">
                  {permission}
                </Badge>
              ))}
            </div>
          </div>
          <Button variant="outline" render={<Link to="/graph" />}>
            Graph view
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
