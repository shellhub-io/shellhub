import { useSearchParams } from "react-router-dom";
import AcceptDeviceFlow from "@/components/devices/AcceptDeviceFlow";
import { setPendingDeviceCode } from "@/utils/navigation";
import LoginLayoutCard from "@/components/layout/LoginLayoutCard";

/**
 * Full-page accept surface, reached from the URL the agent prints
 * (`/accept-device?code=...`) or opened without a code to type one in. The flow
 * itself (resolve → preview → accept) lives in AcceptDeviceFlow, shared with the
 * pairing modal on the Add Device page.
 */
export default function AcceptDevice() {
  const [searchParams] = useSearchParams();
  const code = searchParams.get("code") ?? "";

  if (code) setPendingDeviceCode(code);

  return (
    <LoginLayoutCard>
      <AcceptDeviceFlow code={code} />
    </LoginLayoutCard>
  );
}
