import { Callout } from "@shellhub/design-system/primitives";
import { hasPendingDeviceCode } from "@/utils/navigation";

export default function PendingDeviceCallout() {
  if (!hasPendingDeviceCode()) return null;

  return (
    <Callout variant="info">
      You have a device waiting to be accepted. You&apos;ll be redirected to
      complete the setup after signing in.
    </Callout>
  );
}
