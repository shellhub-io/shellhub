import { isSdkError } from "@/api/errors";
import { isCloud, isEnterprise } from "@/env";

const FALLBACK =
  "An error occurred while accepting the device. Please try again.";

const LICENSE_402 =
  "Your instance has reached its device limit. Please update your license to accept more devices or contact the instance administrator.";
const BILLING_402 =
  "Your subscription plan has reached its device limit. Please update your billing plan to accept more devices.";
const SUBSCRIPTION_402 =
  "Your subscription needs attention before this namespace can accept devices. Open Billing to finish or repair it — your device count is not the problem.";
const NAMESPACE_403 =
  "You do not have permission to accept devices in this namespace.";
const RENAME_409 =
  "A device with this name already exists in the namespace. Please rename the device and try again.";

/**
 * The message the API sends when the namespace's subscription — not its device count — denies
 * the device. The API has no machine-readable error code yet, so the sentinel message is the
 * only signal that separates the two 402s; a change to it falls back to the device-limit copy.
 */
const SUBSCRIPTION_BLOCKED_MESSAGE =
  "the namespace's subscription blocks new devices";

/**
 * Translate an error thrown by the accept-device SDK call into a user-facing
 * message.  The 402 branch is split three ways:
 *   - enterprise && !cloud  → license copy (on-premises)
 *   - cloud                 → billing copy, itself split by whether the namespace is at its
 *                             device limit or held by its subscription
 *   - community             → generic fallback
 */
export function getAcceptDeviceErrorMessage(err: unknown): string {
  if (!isSdkError(err)) return FALLBACK;

  switch (err.status) {
    case 402: {
      if (isCloud()) {
        return err.message === SUBSCRIPTION_BLOCKED_MESSAGE
          ? SUBSCRIPTION_402
          : BILLING_402;
      }
      if (isEnterprise()) return LICENSE_402;
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
