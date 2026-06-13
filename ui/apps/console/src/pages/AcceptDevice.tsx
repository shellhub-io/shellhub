import { useSearchParams } from "react-router-dom";
import AcceptDeviceFlow from "@/components/devices/AcceptDeviceFlow";

/**
 * Full-page accept surface, reached from the URL the agent prints
 * (`/accept-device?code=...`) or opened without a code to type one in. The flow
 * itself (resolve → preview → accept) lives in AcceptDeviceFlow, shared with the
 * pairing modal on the Add Device page.
 */
export default function AcceptDevice() {
  const [searchParams] = useSearchParams();
  const code = searchParams.get("code") ?? "";

  return (
    <div className="w-full max-w-md mx-auto animate-fade-in">
      <div className="bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm">
        <AcceptDeviceFlow code={code} />
      </div>
    </div>
  );
}
