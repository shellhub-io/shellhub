import QRCode from "qrcode";
import { useEffect, useRef } from "react";

interface QRCodeDisplayProps {
  data: string; // otpauth:// URL
  size?: number;
}

export function QRCodeDisplay({ data, size = 200 }: QRCodeDisplayProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    if (canvasRef.current && data) {
      // Validate that data is an otpauth:// URL to prevent phishing attacks
      const validatedData = data.startsWith("otpauth://") ? data : "";

      if (!validatedData) {
        console.error("Invalid QR code data: must be an otpauth:// URL");
        return;
      }

      QRCode.toCanvas(canvasRef.current, validatedData, { width: size }, (error) => {
        if (error) {
          console.error("Error generating QR code:", error);
        }
      });
    }
  }, [data, size]);

  return <canvas ref={canvasRef} className="mx-auto rounded-lg" />;
}
