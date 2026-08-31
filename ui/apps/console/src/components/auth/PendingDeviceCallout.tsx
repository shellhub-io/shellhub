import { Callout } from "@shellhub/design-system/primitives";
import { hasPendingDeviceCode } from "@/utils/navigation";

/**
 * Reminds a signing-in user that a device enrolment is waiting for them. Renders nothing when
 * there is none, so the sign-in screen stays plain in the ordinary case.
 */
export default function PendingDeviceCallout() {
  if (!hasPendingDeviceCode()) return null;

  return (
    <Callout variant="info">
      You have a device waiting to be accepted. You&apos;ll be redirected to
      complete the setup after signing in.
    </Callout>
  );
}
