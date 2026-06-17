import { isSdkError } from "@/api/errors";
import { getConfig } from "@/env";

const FALLBACK =
  "An error occurred while accepting the device. Please try again.";

const LICENSE_402 =
  "Your instance has reached its device limit. Please update your license to accept more devices or contact the instance administrator.";
const BILLING_402 =
  "Your subscription plan has reached its device limit. Please update your billing plan to accept more devices.";
const NAMESPACE_403 =
  "You do not have permission to accept devices in this namespace.";
const RENAME_409 =
  "A device with this name already exists in the namespace. Please rename the device and try again.";

/**
 * Translate an error thrown by the accept-device SDK call into a user-facing
 * message.  The 402 branch is split three ways:
 *   - enterprise && !cloud  → license copy (on-premises)
 *   - cloud                 → billing/subscription copy
 *   - community             → generic fallback
 */
export function getAcceptDeviceErrorMessage(err: unknown): string {
  if (!isSdkError(err)) return FALLBACK;

  switch (err.status) {
    case 402: {
      const { enterprise, cloud } = getConfig();
      // Cloud implies Enterprise, so `{ cloud: true, enterprise: false }` never occurs; check cloud first.
      if (cloud) return BILLING_402;
      if (enterprise) return LICENSE_402;
      return FALLBACK;
    }
    case 403:
      return NAMESPACE_403;
    case 409:
      return RENAME_409;
    default:
      return FALLBACK;
  }
}
