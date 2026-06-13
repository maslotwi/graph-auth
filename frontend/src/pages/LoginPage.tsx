import { Link, useSearchParams } from "react-router-dom"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { QRCode } from "@/components/ui/qr-code"

export default function LoginPage() {
  const [searchParams] = useSearchParams()
  const returnTo = searchParams.get("return") ?? ""

  const verifyHref = returnTo
    ? `/verify?token=demo&return=${encodeURIComponent(returnTo)}`
    : "/verify?token=demo"

  const mockLoginUrl = `${window.location.origin}${verifyHref}`

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Log in</CardTitle>
        <CardDescription>
          Scan this code with an authorized device to approve this session.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-3">
        <div className="rounded-xl border bg-muted/30 p-4">
          <QRCode value={mockLoginUrl} size={200} />
        </div>
        <p className="text-xs text-muted-foreground">
          This code expires in 5 minutes
        </p>
      </CardContent>
      <CardFooter className="flex flex-col gap-2">
        <Button className="w-full" render={<Link to="/register" />}>
          Create an account instead
        </Button>
        <p className="text-sm text-muted-foreground">
          Already verified?{" "}
          <Link to={verifyHref} className="text-foreground underline">
            Continue setup
          </Link>
        </p>
      </CardFooter>
    </Card>
  )
}
