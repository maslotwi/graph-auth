import { Link, Outlet } from "react-router-dom"

import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"

export default function PublicLayout() {
  return (
    <div className="flex min-h-svh flex-col">
      <header className="flex items-center justify-between px-6 py-4">
        <Link to="/" className="font-medium">
          Graph Auth
        </Link>
        <div className="flex items-center gap-2">
          <Button variant="ghost" render={<Link to="/login" />}>
            Log in
          </Button>
          <Button render={<Link to="/register" />}>Sign up</Button>
        </div>
      </header>
      <Separator />
      <main className="flex flex-1 items-center justify-center p-6">
        <Outlet />
      </main>
    </div>
  )
}
