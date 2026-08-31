import { useSearchParams } from "react-router-dom";
import AcceptDeviceFlow from "@/components/devices/AcceptDeviceFlow";
import { setPendingDeviceCode } from "@/utils/navigation";
import LoginLayoutCard from "@/components/layout/LoginLayoutCard";

/**
 * The page an accept-device link lands on. The code comes from the query string, so the flow
 * starts already filled in.
 */
export default function AcceptDevice() {
  const [searchParams] = useSearchParams();
  const code = searchParams.get("code") ?? "";

  if (code) setPendingDeviceCode(code);

  return (
    <LoginLayoutCard>
      <AcceptDeviceFlow initialCode={code} />
    </LoginLayoutCard>
  );
}
