import { Link } from "react-router-dom"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

export default function LoginPage() {
  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Log in</CardTitle>
        <CardDescription>
          Graph Auth uses passwordless login. Scan a QR code from an authorized
          node or open your verification email link.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="rounded-lg border border-dashed p-6 text-center text-sm text-muted-foreground">
          QR scanner coming soon
        </div>
      </CardContent>
      <CardFooter className="flex flex-col gap-2">
        <Button className="w-full" render={<Link to="/register" />}>
          Create an account
        </Button>
        <p className="text-sm text-muted-foreground">
          Already verified?{" "}
          <Link to="/verify?token=demo" className="text-foreground underline">
            Continue setup
          </Link>
        </p>
      </CardFooter>
    </Card>
  )
}
