import { useState } from "react"
import { useNavigate } from "react-router-dom"
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
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useAuth } from "@/hooks/useAuth"
import { ApiError } from "@/types/api"
import { PERMISSIONS, type Permission } from "@/types/node"

export default function RootNodeSetupPage() {
  const navigate = useNavigate()
  const { createRoot, email } = useAuth()
  const [label, setLabel] = useState("My Device")
  const [permissions, setPermissions] = useState<Permission[]>([...PERMISSIONS])
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  function togglePermission(permission: Permission, checked: boolean) {
    setPermissions((current) =>
      checked
        ? [...current, permission]
        : current.filter((value) => value !== permission)
    )
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)

    if (permissions.length === 0) {
      setError("Select at least one permission for your root node.")
      return
    }

    setIsSubmitting(true)

    try {
      await createRoot({ label: label.trim(), permissions })
      toast.success("Root node created")
      void navigate("/", { replace: true })
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "Failed to create root node."
      setError(message)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-svh items-center justify-center p-6">
      <Card className="w-full max-w-lg">
        <CardHeader>
          <CardTitle>Set up your root node</CardTitle>
          <CardDescription>
            This is your root identity in the graph
            {email ? ` for ${email}` : ""}. Child nodes will inherit permissions
            from their predecessors.
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
              <Label htmlFor="label">Node label</Label>
              <Input
                id="label"
                value={label}
                onChange={(event) => setLabel(event.target.value)}
                placeholder="My Laptop"
                required
              />
            </div>
            <div className="flex flex-col gap-3">
              <Label>Permissions</Label>
              {PERMISSIONS.map((permission) => (
                <label
                  key={permission}
                  className="flex items-center gap-2 text-sm capitalize"
                >
                  <Checkbox
                    checked={permissions.includes(permission)}
                    onCheckedChange={(checked) =>
                      togglePermission(permission, checked === true)
                    }
                  />
                  {permission}
                </label>
              ))}
            </div>
          </CardContent>
          <CardFooter>
            <Button className="w-full" type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Creating..." : "Create root node"}
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  )
}
