import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

export default function MyDevicesPage() {
  return (
    <Card className="max-w-lg">
      <CardHeader>
        <CardTitle>My devices</CardTitle>
        <CardDescription>
          Manage devices connected to your account.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">No devices yet.</p>
      </CardContent>
    </Card>
  )
}
