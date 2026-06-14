import { useState } from "react"
import { Link, Outlet, useNavigate } from "react-router-dom"
import { toast } from "sonner"

import * as authApi from "@/api/auth"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { useAuth } from "@/hooks/useAuth"

export default function AppLayout() {
  const { logout } = useAuth()
  const navigate = useNavigate()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [isLoggingOut, setIsLoggingOut] = useState(false)

  async function handleConfirmLogout() {
    setIsLoggingOut(true)
    try {
      await authApi.invalidateSession()
    } catch {
      toast.error("Failed to reach the server, but your local session has been cleared.")
    } finally {
      logout()
      setIsLoggingOut(false)
      void navigate("/register", { replace: true })
    }
  }

  return (
    <div className="flex min-h-svh flex-col">
      <header className="flex items-center justify-between px-6 py-4">
        <nav className="flex items-center gap-4">
          <Link to="/" className="font-medium">
            Graph Auth
          </Link>
          <Button variant="ghost" render={<Link to="/devices" />}>
            My devices
          </Button>
          <Button variant="ghost" render={<Link to="/graph" />}>
            Graph
          </Button>
        </nav>
        <Button variant="outline" onClick={() => setDialogOpen(true)}>
          Log out
        </Button>
      </header>
      <Separator />
      <main className="flex-1 p-6">
        <Outlet />
      </main>

      <AlertDialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Log out this device?</AlertDialogTitle>
            <AlertDialogDescription>
              This will permanently{" "}
              <strong className="text-foreground">invalidate this device session</strong> and{" "}
              <strong className="text-foreground">all devices that were authorized through it</strong>.
              {" "}You cannot log back in with this session — to use Graph Auth again on this device you will need to register or be invited by another active device.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isLoggingOut}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => void handleConfirmLogout()}
              disabled={isLoggingOut}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isLoggingOut ? "Logging out…" : "Yes, log out and invalidate"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
