import QRCode from "qrcode";
import { useEffect, useRef } from "react";

interface QRCodeDisplayProps {
  data: string; // otpauth:// URL
  size?: number;
}

/**
 * Renders an MFA secret as a QR code on a canvas. Drawn in the browser, so the secret is never
 * sent anywhere to be turned into an image.
 */
export function QRCodeDisplay({ data, size = 200 }: QRCodeDisplayProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    if (canvasRef.current && data) {
      QRCode.toCanvas(canvasRef.current, data, { width: size }, (error) => {
        if (error) {
          console.error("Error generating QR code:", error);
        }
      });
    }
  }, [data, size]);

  return <canvas ref={canvasRef} className="mx-auto rounded-lg" />;
}
