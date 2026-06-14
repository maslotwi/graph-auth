import { Navigate, useSearchParams } from "react-router-dom"

export default function LoginPage() {
  const [searchParams] = useSearchParams()
  const code = searchParams.get("code")
  return <Navigate to={code ? `/join?code=${code}` : "/join"} replace />
}
