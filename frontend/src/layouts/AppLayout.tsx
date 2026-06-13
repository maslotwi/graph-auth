import { Link, Outlet } from "react-router-dom"

import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { useAuth } from "@/hooks/useAuth"

export default function AppLayout() {
  const { logout } = useAuth()

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
        </nav>
        <Button variant="outline" onClick={logout}>
          Log out
        </Button>
      </header>
      <Separator />
      <main className="flex-1 p-6">
        <Outlet />
      </main>
    </div>
  )
}
