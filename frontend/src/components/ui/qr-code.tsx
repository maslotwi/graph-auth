import QRCodeSVG from "react-qr-code"

type QRCodeProps = {
  value: string
  size?: number
  className?: string
}

export function QRCode({ value, size = 180, className }: QRCodeProps) {
  return (
    <div className={className}>
      <QRCodeSVG
        value={value}
        size={size}
        bgColor="transparent"
        fgColor="currentColor"
        style={{ width: "100%", height: "auto" }}
      />
    </div>
  )
}
